package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"keuangan_nawawi/database"
	"keuangan_nawawi/models"

	"github.com/wailsapp/wails/v2/pkg/options"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type TransaksiInput struct {
	EntitasID   int                   `json:"entitas_id"`
	Tanggal     string                `json:"tanggal"`
	NoReferensi string                `json:"no_referensi"`
	Keterangan  string                `json:"keterangan"`
	TotalDebit  int64                 `json:"total_debit"`
	TotalKredit int64                 `json:"total_kredit"`
	Details     []models.JurnalDetail `json:"details"`
}

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Inisialisasi Database
	database.ConnectDB()
}

// onSecondInstanceLaunch dipanggil saat ada instansi aplikasi kedua yang dibuka (Poin 2)
func (a *App) onSecondInstanceLaunch(secondInstanceData options.SecondInstanceData) {
	if a.ctx != nil {
		wailsRuntime.WindowShow(a.ctx)
		wailsRuntime.WindowUnminimise(a.ctx)
	}
}

// GetDatabaseInfo mengambil ringkasan informasi lokasi dan ukuran database (Poin 1)
func (a *App) GetDatabaseInfo() (map[string]interface{}, error) {
	dbPath := database.GetDbPath()
	fi, err := os.Stat(dbPath)
	sizeStr := "0 KB"
	modStr := "-"
	if err == nil {
		sizeKB := float64(fi.Size()) / 1024.0
		if sizeKB > 1024.0 {
			sizeStr = fmt.Sprintf("%.2f MB", sizeKB/1024.0)
		} else {
			sizeStr = fmt.Sprintf("%.1f KB", sizeKB)
		}
		modStr = fi.ModTime().Format("02 Jan 2006, 15:04 WIB")
	}
	return map[string]interface{}{
		"path":         dbPath,
		"size":         sizeStr,
		"lastModified": modStr,
	}, nil
}

// BackupDatabase membuka dialog simpan dan mencadangkan file database SQLite (Poin 1)
func (a *App) BackupDatabase() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("runtime context belum siap")
	}

	// Tulis semua log transaksi WAL ke file utama terlebih dahulu
	if err := database.CheckpointWAL(); err != nil {
		log.Printf("Peringatan checkpoint WAL: %v", err)
	}

	nowStr := time.Now().Format("2006-01-02_150405")
	defaultFilename := fmt.Sprintf("Backup_Keuangan_Nawawi_%s.db", nowStr)

	destPath, err := wailsRuntime.SaveFileDialog(a.ctx, wailsRuntime.SaveDialogOptions{
		Title:           "Pilih Lokasi Simpan Cadangan Database",
		DefaultFilename: defaultFilename,
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "Database SQLite (*.db)",
				Pattern:     "*.db",
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("gagal membuka dialog penyimpanan: %v", err)
	}
	if destPath == "" {
		// Pengguna membatalkan dialog
		return "", nil
	}

	srcPath := database.GetDbPath()
	if err := database.CopyFile(srcPath, destPath); err != nil {
		return "", fmt.Errorf("gagal menyalin file database: %v", err)
	}

	log.Printf("Database berhasil dicadangkan ke: %s", destPath)
	return destPath, nil
}

// RestoreDatabase membuka dialog file dan memulihkan database dari cadangan (Poin 1)
func (a *App) RestoreDatabase() (string, error) {
	if a.ctx == nil {
		return "", fmt.Errorf("runtime context belum siap")
	}

	srcPath, err := wailsRuntime.OpenFileDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "Pilih File Cadangan Database untuk Dipulihkan",
		Filters: []wailsRuntime.FileFilter{
			{
				DisplayName: "Database SQLite (*.db)",
				Pattern:     "*.db",
			},
		},
	})

	if err != nil {
		return "", fmt.Errorf("gagal membuka dialog file: %v", err)
	}
	if srcPath == "" {
		// Pengguna membatalkan
		return "", nil
	}

	// Validasi integritas dasar file cadangan
	fi, err := os.Stat(srcPath)
	if err != nil || fi.Size() < 100 {
		return "", fmt.Errorf("file cadangan kosong atau tidak dapat diakses")
	}

	fileHeader := make([]byte, 16)
	f, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %v", err)
	}
	_, _ = f.Read(fileHeader)
	f.Close()

	if !strings.HasPrefix(string(fileHeader), "SQLite format 3") {
		return "", fmt.Errorf("file yang dipilih bukan file database SQLite yang sah")
	}

	targetPath := database.GetDbPath()

	// Buat backup penyelamatan darurat jika ada kendala saat restore
	safetyBackup := filepath.Join(database.GetAppDataDir(), fmt.Sprintf("keuangan_data_safety_backup_%s.db", time.Now().Format("20060102_150405")))
	_ = database.CopyFile(targetPath, safetyBackup)

	// Tutup koneksi database aktif
	_ = database.CloseDB()

	// Bersihkan cache file WAL dan SHM lama
	_ = os.Remove(targetPath + "-wal")
	_ = os.Remove(targetPath + "-shm")

	// Salin file backup ke database utama
	if err := database.CopyFile(srcPath, targetPath); err != nil {
		_ = database.CopyFile(safetyBackup, targetPath)
		database.ConnectDB()
		return "", fmt.Errorf("gagal memulihkan database: %v", err)
	}

	// Reconnect database
	database.ConnectDB()
	log.Printf("Database berhasil dipulihkan dari: %s", srcPath)

	return srcPath, nil
}

// Dapatkan daftar entitas
func (a *App) GetEntitas() ([]models.Entitas, error) {
	var list []models.Entitas
	err := database.DB.Find(&list).Error
	return list, err
}

// Buat Entitas Baru
func (a *App) CreateEntitas(nama, tipe string) (models.Entitas, error) {
	entitas := models.Entitas{Nama: nama, Tipe: tipe}
	tx := database.DB.Begin()
	if err := tx.Create(&entitas).Error; err != nil {
		tx.Rollback()
		return entitas, err
	}
	
	// Seed akun Bank Bawaan
	bankAkun := models.AkunKeuangan{
		EntitasID: entitas.ID,
		KodeAkun:  "1110", // Contoh kode kas/bank
		NamaAkun:  "Bank",
		Kategori:  "Aset",
		Sifat:     "Debit",
		IsBawaan:  true,
	}
	if err := tx.Create(&bankAkun).Error; err != nil {
		tx.Rollback()
		return entitas, err
	}
	
	tx.Commit()
	return entitas, nil
}

// Dapatkan semua Akun Keuangan berdasarkan EntitasID
func (a *App) GetAkunKeuangan(entitasID int) ([]models.AkunKeuangan, error) {
	var akun []models.AkunKeuangan
	result := database.DB.Where("entitas_id = ?", entitasID).Find(&akun)
	return akun, result.Error
}

// Buat Akun Keuangan Baru
func (a *App) CreateAkunKeuangan(akun models.AkunKeuangan) (models.AkunKeuangan, error) {
	result := database.DB.Create(&akun)
	return akun, result.Error
}

// Update Akun Keuangan
func (a *App) UpdateAkunKeuangan(akun models.AkunKeuangan) (models.AkunKeuangan, error) {
	result := database.DB.Save(&akun)
	return akun, result.Error
}

// Hapus Akun Keuangan
func (a *App) DeleteAkunKeuangan(id int) error {
	var akun models.AkunKeuangan
	if err := database.DB.First(&akun, id).Error; err != nil {
		return err
	}
	if akun.IsBawaan {
		return fmt.Errorf("Akun ini adalah akun bawaan sistem dan tidak boleh dihapus")
	}

	// Cek apakah akun sudah dipakai di transaksi
	var count int64
	database.DB.Model(&models.JurnalDetail{}).Where("akun_id = ?", id).Count(&count)
	if count > 0 {
		return fmt.Errorf("Akun ini tidak bisa dihapus karena sudah memiliki riwayat transaksi/jurnal")
	}

	return database.DB.Delete(&models.AkunKeuangan{}, id).Error
}

// --- CRUD REKENING BANK ---

func (a *App) GetRekeningBank(entitasID int) ([]models.RekeningBank, error) {
	var list []models.RekeningBank
	err := database.DB.Where("entitas_id = ?", entitasID).Find(&list).Error
	return list, err
}

func (a *App) CreateRekeningBank(input models.RekeningBank) (models.RekeningBank, error) {
	tx := database.DB.Begin()
	if err := tx.Create(&input).Error; err != nil {
		tx.Rollback()
		return input, err
	}

	// Update Saldo Master Bank
	var masterBank models.AkunKeuangan
	if err := tx.Where("entitas_id = ? AND is_bawaan = ? AND nama_akun = ?", input.EntitasID, true, "Bank").First(&masterBank).Error; err == nil {
		masterBank.Saldo += input.Saldo
		tx.Save(&masterBank)
	}

	tx.Commit()
	return input, nil
}

func (a *App) UpdateRekeningBank(input models.RekeningBank) (models.RekeningBank, error) {
	var existing models.RekeningBank
	if err := database.DB.First(&existing, input.ID).Error; err != nil {
		return existing, err
	}
	existing.NamaRekening = input.NamaRekening
	existing.NomorRekening = input.NomorRekening
	err := database.DB.Save(&existing).Error
	return existing, err
}

func (a *App) DeleteRekeningBank(id int) error {
	tx := database.DB.Begin()
	// Cek apakah dipakai di JurnalDetail
	var count int64
	tx.Model(&models.JurnalDetail{}).Where("rekening_bank_id = ?", id).Count(&count)
	if count > 0 {
		tx.Rollback()
		return fmt.Errorf("Rekening ini tidak bisa dihapus karena memiliki riwayat jurnal")
	}

	var rek models.RekeningBank
	if err := tx.First(&rek, id).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Kurangi Saldo Master Bank
	var masterBank models.AkunKeuangan
	if err := tx.Where("entitas_id = ? AND is_bawaan = ? AND nama_akun = ?", rek.EntitasID, true, "Bank").First(&masterBank).Error; err == nil {
		masterBank.Saldo -= rek.Saldo
		tx.Save(&masterBank)
	}

	if err := tx.Delete(&models.RekeningBank{}, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit().Error
}

// --- CRUD SISWA ---

func (a *App) GetSiswa(entitasID int) ([]models.Siswa, error) {
	var list []models.Siswa
	err := database.DB.Where("entitas_id = ?", entitasID).Order("jenjang, kelas, nama_lengkap").Find(&list).Error
	return list, err
}

func (a *App) CreateSiswa(input models.Siswa) (models.Siswa, error) {
	err := database.DB.Create(&input).Error
	return input, err
}

func (a *App) UpdateSiswa(input models.Siswa) (models.Siswa, error) {
	var existing models.Siswa
	if err := database.DB.First(&existing, input.ID).Error; err != nil {
		return existing, err
	}
	existing.NIS = input.NIS
	existing.NISN = input.NISN
	existing.NamaLengkap = input.NamaLengkap
	existing.JenisKelamin = input.JenisKelamin
	existing.TempatLahir = input.TempatLahir
	existing.TanggalLahir = input.TanggalLahir
	existing.Agama = input.Agama
	existing.Alamat = input.Alamat
	existing.NamaAyah = input.NamaAyah
	existing.NamaIbu = input.NamaIbu
	existing.TeleponOrtu = input.TeleponOrtu
	existing.Kelas = input.Kelas
	existing.Jenjang = input.Jenjang
	existing.Status = input.Status
	err := database.DB.Save(&existing).Error
	return existing, err
}

func (a *App) DeleteSiswa(id int) error {
	return database.DB.Delete(&models.Siswa{}, id).Error
}

// Export Akun Keuangan ke string CSV
func (a *App) ExportAkunCSV(entitasID int) (string, error) {
	var akun []models.AkunKeuangan
	if err := database.DB.Where("entitas_id = ?", entitasID).Find(&akun).Error; err != nil {
		return "", err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)
	
	// Header
	writer.Write([]string{"Kode Akun", "Nama Akun", "Kategori", "Sifat", "Saldo Awal"})
	
	for _, ak := range akun {
		writer.Write([]string{
			ak.KodeAkun,
			ak.NamaAkun,
			ak.Kategori,
			ak.Sifat,
			fmt.Sprintf("%d", ak.Saldo),
		})
	}
	writer.Flush()
	
	return buf.String(), nil
}

// Import Akun Keuangan dari string CSV
func (a *App) ImportAkunCSV(entitasID int, csvData string) error {
	// Proteksi "Siluman": Batasi ukuran memori string dari backend (Maksimal 1 MB)
	// Jika ada yang membypass Frontend JS dan mengirim string raksasa, langsung tolak.
	if len(csvData) > 1024*1024 {
		return fmt.Errorf("ukuran data melampaui batas maksimal server (1 MB)")
	}

	reader := csv.NewReader(strings.NewReader(csvData))
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}
	
	if len(records) < 2 {
		return nil // Kosong atau hanya header
	}
	
	if len(records) > 501 { // 500 baris data + 1 header
		return fmt.Errorf("maksimal 500 baris data dalam satu kali import")
	}
	
	// Mulai dari indeks 1 untuk melewati header
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) < 5 {
			continue // Skip baris tidak lengkap
		}
		
		saldo, _ := strconv.ParseInt(row[4], 10, 64)
		
		akunBaru := models.AkunKeuangan{
			EntitasID: entitasID,
			KodeAkun:  row[0],
			NamaAkun:  row[1],
			Kategori:  row[2],
			Sifat:     row[3],
			Saldo:     saldo,
		}
		
		// Gunakan FirstOrCreate atau Create (asumsi bisa duplikat kalau user copas mentah, jadi kita cek kode akun dulu)
		var existing models.AkunKeuangan
		err := database.DB.Where("entitas_id = ? AND kode_akun = ?", entitasID, akunBaru.KodeAkun).First(&existing).Error
		if err != nil {
			// Belum ada, Create
			database.DB.Create(&akunBaru)
		}
	}
	return nil
}

// Generate ISAK 35 Template
func (a *App) GenerateISAK35(entitasID int) error {
	// Template standar ISAK 35 (Disesuaikan dengan Pos Anggaran Yayasan)
	akunTemplate := []models.AkunKeuangan{
		// Aset (1)
		{KodeAkun: "1111", NamaAkun: "Kas Tunai", Kategori: "Aset", Sifat: "Debit"},
		// (Kode 1110 "Bank" sudah otomatis dibuat oleh sistem)
		{KodeAkun: "1130", NamaAkun: "Piutang Usaha/SPP", Kategori: "Aset", Sifat: "Debit"},
		{KodeAkun: "1140", NamaAkun: "Persediaan (Seragam/Buku)", Kategori: "Aset", Sifat: "Debit"},
		// Liabilitas (2)
		{KodeAkun: "2110", NamaAkun: "Utang Usaha", Kategori: "Liabilitas", Sifat: "Kredit"},
		{KodeAkun: "2120", NamaAkun: "Utang Gaji", Kategori: "Liabilitas", Sifat: "Kredit"},
		// Aset Neto (3)
		{KodeAkun: "3110", NamaAkun: "Aset Neto Tanpa Pembatasan", Kategori: "Aset Neto", Sifat: "Kredit"},
		{KodeAkun: "3120", NamaAkun: "Aset Neto Dengan Pembatasan", Kategori: "Aset Neto", Sifat: "Kredit"},
		// Pendapatan (4) - Harus persis dengan format jenis tagihan
		{KodeAkun: "4110", NamaAkun: "Pendapatan SPP", Kategori: "Pendapatan", Sifat: "Kredit"},
		{KodeAkun: "4120", NamaAkun: "Pendapatan Semesteran", Kategori: "Pendapatan", Sifat: "Kredit"},
		{KodeAkun: "4130", NamaAkun: "Pendapatan Seragam", Kategori: "Pendapatan", Sifat: "Kredit"},
		{KodeAkun: "4140", NamaAkun: "Pendapatan Buku", Kategori: "Pendapatan", Sifat: "Kredit"},
		{KodeAkun: "4150", NamaAkun: "Pendapatan Uang Pangkal", Kategori: "Pendapatan", Sifat: "Kredit"},
		{KodeAkun: "4160", NamaAkun: "Pendapatan Asrama", Kategori: "Pendapatan", Sifat: "Kredit"},
		{KodeAkun: "4170", NamaAkun: "Pendapatan Infaq", Kategori: "Pendapatan", Sifat: "Kredit"},
		{KodeAkun: "4180", NamaAkun: "Pendapatan Koperasi", Kategori: "Pendapatan", Sifat: "Kredit"},
		{KodeAkun: "4190", NamaAkun: "Pendapatan Bantuan Operasional (BOS)", Kategori: "Pendapatan", Sifat: "Kredit"},
		// Beban (5)
		{KodeAkun: "5110", NamaAkun: "Beban Gaji, THR & Insentif Guru", Kategori: "Beban", Sifat: "Debit"},
		{KodeAkun: "5120", NamaAkun: "Beban Konsumsi & Makan Siswa", Kategori: "Beban", Sifat: "Debit"},
		{KodeAkun: "5130", NamaAkun: "Beban Kegiatan & Ujian Siswa", Kategori: "Beban", Sifat: "Debit"},
		{KodeAkun: "5140", NamaAkun: "Beban Kesehatan & Pengobatan Siswa", Kategori: "Beban", Sifat: "Debit"},
		{KodeAkun: "5150", NamaAkun: "Beban Pengadaan Seragam", Kategori: "Beban", Sifat: "Debit"},
		{KodeAkun: "5160", NamaAkun: "Beban Pengadaan Buku", Kategori: "Beban", Sifat: "Debit"},
		{KodeAkun: "5170", NamaAkun: "Beban Pembangunan & Sarpras", Kategori: "Beban", Sifat: "Debit"},
		{KodeAkun: "5180", NamaAkun: "Beban Perlengkapan Asrama (Ranjang, Kasur)", Kategori: "Beban", Sifat: "Debit"},
		{KodeAkun: "5190", NamaAkun: "Beban Utilitas (Listrik, Air, Internet)", Kategori: "Beban", Sifat: "Debit"},
		{KodeAkun: "5200", NamaAkun: "Beban Pemeliharaan & Alat Kebersihan", Kategori: "Beban", Sifat: "Debit"},
		{KodeAkun: "5210", NamaAkun: "Beban Operasional (Transport, ATK, Fotocopy)", Kategori: "Beban", Sifat: "Debit"},
	}

	for _, akun := range akunTemplate {
		akun.EntitasID = entitasID
		var existing models.AkunKeuangan
		err := database.DB.Where("entitas_id = ? AND kode_akun = ?", entitasID, akun.KodeAkun).First(&existing).Error
		if err != nil {
			database.DB.Create(&akun)
		}
	}
	return nil
}

// --- FUNGSI JURNAL UMUM (TRANSAKSI) ---

// Dapatkan daftar transaksi
func (a *App) GetTransaksi(entitasID int) ([]models.Transaksi, error) {
	var list []models.Transaksi
	err := database.DB.Where("entitas_id = ?", entitasID).Order("tanggal desc, id desc").Find(&list).Error
	return list, err
}

// Dapatkan Jurnal Detail untuk satu transaksi (jika diperlukan)
func (a *App) GetJurnalDetail(transaksiID int) ([]models.JurnalDetail, error) {
	var list []models.JurnalDetail
	err := database.DB.Preload("Akun").Where("transaksi_id = ?", transaksiID).Find(&list).Error
	return list, err
}

// Buat Transaksi Baru (Jurnal Umum) dengan Database Transaction
func (a *App) CreateTransaksi(input TransaksiInput) error {
	// Mulai Database Transaction (all or nothing)
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 1. Simpan Transaksi Header
	transaksi := models.Transaksi{
		EntitasID:   input.EntitasID,
		Tanggal:     input.Tanggal,
		NoReferensi: input.NoReferensi,
		Keterangan:  input.Keterangan,
		TotalDebit:  input.TotalDebit,
		TotalKredit: input.TotalKredit,
	}

	if err := tx.Create(&transaksi).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. Simpan Detail Jurnal dan Update Saldo Akun
	for _, det := range input.Details {
		det.TransaksiID = transaksi.ID
		
		if err := tx.Create(&det).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Cari akun terkait untuk diupdate saldonya
		var akun models.AkunKeuangan
		if err := tx.First(&akun, det.AkunID).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Logika Akuntansi (Sifat Saldo Normal)
		if strings.EqualFold(akun.Sifat, "Debit") {
			akun.Saldo += det.Debit
			akun.Saldo -= det.Kredit
		} else { // Kredit
			akun.Saldo += det.Kredit
			akun.Saldo -= det.Debit
		}

		if err := tx.Save(&akun).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Jika ada RekeningBankID, sync saldo sub-ledger
		if det.RekeningBankID != nil && *det.RekeningBankID > 0 {
			var rek models.RekeningBank
			if err := tx.First(&rek, *det.RekeningBankID).Error; err == nil {
				// Akun Bank sifatnya Debit: debit = masuk ke bank, kredit = keluar dari bank
				rek.Saldo += det.Debit
				rek.Saldo -= det.Kredit
				tx.Save(&rek)
			}
		}
	}

	// 3. Sukses, commit semua perubahan
	return tx.Commit().Error
}

// Hapus Transaksi (Rollback Saldo)
func (a *App) DeleteTransaksi(transaksiID int) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 1. Ambil detail jurnal untuk mereverse saldo akun
	var details []models.JurnalDetail
	if err := tx.Where("transaksi_id = ?", transaksiID).Find(&details).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, det := range details {
		var akun models.AkunKeuangan
		if err := tx.First(&akun, det.AkunID).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Reverse logika akuntansi
		if strings.EqualFold(akun.Sifat, "Debit") {
			akun.Saldo -= det.Debit
			akun.Saldo += det.Kredit
		} else { // Kredit
			akun.Saldo -= det.Kredit
			akun.Saldo += det.Debit
		}

		if err := tx.Save(&akun).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Reverse saldo sub-ledger rekening bank
		if det.RekeningBankID != nil && *det.RekeningBankID > 0 {
			var rek models.RekeningBank
			if err := tx.First(&rek, *det.RekeningBankID).Error; err == nil {
				rek.Saldo -= det.Debit
				rek.Saldo += det.Kredit
				tx.Save(&rek)
			}
		}
	}

	// 2. Hapus detail jurnal
	if err := tx.Where("transaksi_id = ?", transaksiID).Delete(&models.JurnalDetail{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2.5 Hapus riwayat gaji jika ada kaitannya dengan transaksi ini
	if err := tx.Where("transaksi_id = ?", transaksiID).Delete(&models.RiwayatGaji{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2.6 Rollback status TagihanSiswa jika transaksi ini dari pembayaran tagihan
	var riwayatPembayaran []models.RiwayatPembayaranTagihan
	tx.Where("transaksi_id = ?", transaksiID).Find(&riwayatPembayaran)
	for _, rp := range riwayatPembayaran {
		var tagihan models.TagihanSiswa
		if err := tx.First(&tagihan, rp.TagihanSiswaID).Error; err == nil {
			tagihan.SisaTagihan += rp.NominalBayar
			if tagihan.SisaTagihan >= tagihan.TotalTagihan {
				tagihan.Status = "Belum Lunas"
				tagihan.SisaTagihan = tagihan.TotalTagihan
			} else {
				tagihan.Status = "Sebagian"
			}
			tx.Save(&tagihan)
		}
	}
	// Hapus riwayat pembayaran tagihan terkait
	if err := tx.Where("transaksi_id = ?", transaksiID).Delete(&models.RiwayatPembayaranTagihan{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 3. Hapus transaksi header
	if err := tx.Delete(&models.Transaksi{}, transaksiID).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}

// Update Transaksi (Edit)
func (a *App) UpdateTransaksi(id int, input TransaksiInput) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 1. Ambil detail jurnal LAMA untuk mereverse saldo
	var oldDetails []models.JurnalDetail
	if err := tx.Where("transaksi_id = ?", id).Find(&oldDetails).Error; err != nil {
		tx.Rollback()
		return err
	}

	for _, det := range oldDetails {
		var akun models.AkunKeuangan
		if err := tx.First(&akun, det.AkunID).Error; err != nil {
			tx.Rollback()
			return err
		}
		// Reverse
		if strings.EqualFold(akun.Sifat, "Debit") {
			akun.Saldo -= det.Debit
			akun.Saldo += det.Kredit
		} else { // Kredit
			akun.Saldo -= det.Kredit
			akun.Saldo += det.Debit
		}
		if err := tx.Save(&akun).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 2. Hapus detail lama
	if err := tx.Where("transaksi_id = ?", id).Delete(&models.JurnalDetail{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 3. Update header transaksi
	var transaksi models.Transaksi
	if err := tx.First(&transaksi, id).Error; err != nil {
		tx.Rollback()
		return err
	}
	transaksi.Tanggal = input.Tanggal
	transaksi.NoReferensi = input.NoReferensi
	transaksi.Keterangan = input.Keterangan
	transaksi.TotalDebit = input.TotalDebit
	transaksi.TotalKredit = input.TotalKredit
	if err := tx.Save(&transaksi).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 4. Insert detail BARU dan update saldo
	for _, det := range input.Details {
		det.TransaksiID = transaksi.ID
		if err := tx.Create(&det).Error; err != nil {
			tx.Rollback()
			return err
		}

		var akun models.AkunKeuangan
		if err := tx.First(&akun, det.AkunID).Error; err != nil {
			tx.Rollback()
			return err
		}

		if strings.EqualFold(akun.Sifat, "Debit") {
			akun.Saldo += det.Debit
			akun.Saldo -= det.Kredit
		} else { // Kredit
			akun.Saldo += det.Kredit
			akun.Saldo -= det.Debit
		}

		if err := tx.Save(&akun).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// ==========================================
// MODUL PENGGAJIAN
// ==========================================

func (a *App) GetPegawai(entitasID int) ([]models.Pegawai, error) {
	var list []models.Pegawai
	err := database.DB.Where("entitas_id = ?", entitasID).Find(&list).Error
	return list, err
}



func (a *App) GetBulanGajiTerproses(tahun int, entitasID int) ([]int, error) {
	var bulanList []int
	err := database.DB.Model(&models.RiwayatGaji{}).
		Joins("JOIN pegawais ON pegawais.id = riwayat_gajis.pegawai_id").
		Where("riwayat_gajis.tahun = ? AND pegawais.entitas_id = ?", tahun, entitasID).
		Distinct("riwayat_gajis.bulan").
		Pluck("riwayat_gajis.bulan", &bulanList).Error
	return bulanList, err
}

type PegawaiInput struct {
	EntitasID         int    `json:"entitas_id"`
	NIP               string `json:"nip"`
	Nama              string `json:"nama"`
	Jabatan           string `json:"jabatan"`
	GajiPokok         int64  `json:"gaji_pokok"`
	TanggalMulaiKerja string `json:"tanggal_mulai_kerja"`
	PeriodeKenaikan   int    `json:"periode_kenaikan"`
	NominalKenaikan   int64  `json:"nominal_kenaikan"`
	Tunjangan         string `json:"tunjangan"`
	Potongan          string `json:"potongan"`
}

func (a *App) CreatePegawai(input PegawaiInput) error {
	pegawai := models.Pegawai{
		EntitasID:         input.EntitasID,
		NIP:               input.NIP,
		Nama:              input.Nama,
		Jabatan:           input.Jabatan,
		GajiPokok:         input.GajiPokok,
		TanggalMulaiKerja: input.TanggalMulaiKerja,
		PeriodeKenaikan:   input.PeriodeKenaikan,
		NominalKenaikan:   input.NominalKenaikan,
		Tunjangan:         input.Tunjangan,
		Potongan:          input.Potongan,
	}
	return database.DB.Create(&pegawai).Error
}

func (a *App) UpdatePegawai(id int, input PegawaiInput) error {
	var pegawai models.Pegawai
	if err := database.DB.First(&pegawai, id).Error; err != nil {
		return err
	}

	pegawai.NIP = input.NIP
	pegawai.Nama = input.Nama
	pegawai.Jabatan = input.Jabatan
	pegawai.GajiPokok = input.GajiPokok
	pegawai.TanggalMulaiKerja = input.TanggalMulaiKerja
	pegawai.PeriodeKenaikan = input.PeriodeKenaikan
	pegawai.NominalKenaikan = input.NominalKenaikan
	pegawai.Tunjangan = input.Tunjangan
	pegawai.Potongan = input.Potongan

	return database.DB.Save(&pegawai).Error
}

func (a *App) DeletePegawai(id int) error {
	return database.DB.Delete(&models.Pegawai{}, id).Error
}

func (a *App) GetRiwayatGajiBulanan(bulan int, tahun int, entitasID int) ([]models.RiwayatGaji, error) {
	var list []models.RiwayatGaji
	err := database.DB.Preload("Pegawai").Joins("JOIN pegawais ON pegawais.id = riwayat_gajis.pegawai_id").
		Where("pegawais.entitas_id = ? AND bulan = ? AND tahun = ?", entitasID, bulan, tahun).
		Find(&list).Error
	return list, err
}

type KomponenGaji struct {
	Nama    string `json:"nama"`
	Nominal int64  `json:"nominal"`
}

type ProsesGajiInput struct {
	EntitasID int    `json:"entitas_id"`
	Bulan     int    `json:"bulan"`
	Tahun          int    `json:"tahun"`
	RekeningBankID int    `json:"rekening_bank_id"`
	Tanggal   string `json:"tanggal"`
}

func (a *App) ProsesPenggajianBulanan(input ProsesGajiInput) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Cek apakah sudah pernah diproses bulan ini
	var count int64
	tx.Model(&models.RiwayatGaji{}).Joins("JOIN pegawais ON pegawais.id = riwayat_gajis.pegawai_id").
		Where("pegawais.entitas_id = ? AND bulan = ? AND tahun = ?", input.EntitasID, input.Bulan, input.Tahun).Count(&count)
	
	if count > 0 {
		tx.Rollback()
		return fmt.Errorf("Gaji bulan %d tahun %d sudah pernah diproses", input.Bulan, input.Tahun)
	}

	// 1. Ambil daftar pegawai
	var pegawais []models.Pegawai
	if err := tx.Where("entitas_id = ?", input.EntitasID).Find(&pegawais).Error; err != nil {
		tx.Rollback()
		return err
	}

	if len(pegawais) == 0 {
		tx.Rollback()
		return fmt.Errorf("Tidak ada data pegawai untuk diproses")
	}

	// 2. Hitung total gaji aktual (termasuk kenaikan berkala, tunjangan, dan potongan)
	var totalGaji int64
	gajiAktualMap := make(map[int]int64) // simpan gaji aktual per pegawai
	tunjanganMap := make(map[int]string)
	potonganMap := make(map[int]string)

	// Parse tanggal proses (default to today if empty)
	hitungDate, err := time.Parse("2006-01-02", input.Tanggal)
	if err != nil || input.Tanggal == "" {
		hitungDate = time.Now()
	}
	tanggalStr := hitungDate.Format("2006-01-02")

	for _, p := range pegawais {
		gajiAktual := p.GajiPokok
		if p.TanggalMulaiKerja != "" && p.PeriodeKenaikan > 0 && p.NominalKenaikan > 0 {
			tanggalMulai := p.TanggalMulaiKerja
			if len(tanggalMulai) >= 10 {
				tanggalMulai = tanggalMulai[:10]
			}
			mulaiDate, err := time.Parse("2006-01-02", tanggalMulai)
			if err == nil {
				// Hitung selisih tahun
				tahunKerja := hitungDate.Year() - mulaiDate.Year()
				// Penyesuaian jika bulan/hari belum lewat
				if hitungDate.Month() < mulaiDate.Month() || (hitungDate.Month() == mulaiDate.Month() && hitungDate.Day() < mulaiDate.Day()) {
					tahunKerja--
				}
				
				if tahunKerja > 0 {
					jumlahKenaikan := tahunKerja / p.PeriodeKenaikan
					gajiAktual += int64(jumlahKenaikan) * p.NominalKenaikan
				}
			}
		}

		// Kalkulasi Tunjangan
		var tunjanganList []KomponenGaji
		if p.Tunjangan != "" {
			json.Unmarshal([]byte(p.Tunjangan), &tunjanganList)
			for _, t := range tunjanganList {
				gajiAktual += t.Nominal
			}
		}

		// Kalkulasi Potongan
		var potonganList []KomponenGaji
		if p.Potongan != "" {
			json.Unmarshal([]byte(p.Potongan), &potonganList)
			for _, pot := range potonganList {
				gajiAktual -= pot.Nominal
			}
		}

		// Save details for snapshot
		tunjanganJSON, _ := json.Marshal(tunjanganList)
		potonganJSON, _ := json.Marshal(potonganList)
		tunjanganMap[p.ID] = string(tunjanganJSON)
		potonganMap[p.ID] = string(potonganJSON)

		gajiAktualMap[p.ID] = gajiAktual
		totalGaji += gajiAktual
	}

	// 3. Cari Akun Beban Gaji (Otomatis deteksi berdasarkan kategori "Beban")
	// Idealnya ini dipetakan secara statis, tapi kita pakai pencarian untuk kemudahan
	var akunBeban models.AkunKeuangan
	if err := tx.Where("entitas_id = ? AND nama_akun LIKE ?", input.EntitasID, "%Beban%Gaji%").First(&akunBeban).Error; err != nil {
		// Fallback cari kategori Beban
		if err := tx.Where("entitas_id = ? AND kategori = ?", input.EntitasID, "Beban").First(&akunBeban).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("Gagal menemukan akun Beban Gaji di Bagan Akun. Harap buat akun 'Beban Gaji'.")
		}
	}

	// Cari akun Bawaan "Bank" untuk Jurnal Buku Besar
	var akunBankUtama models.AkunKeuangan
	if err := tx.Where("entitas_id = ? AND is_bawaan = ? AND nama_akun = ?", input.EntitasID, true, "Bank").First(&akunBankUtama).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Gagal menemukan akun sistem 'Bank'. Pastikan akun bawaan Bank tersedia.")
	}

	// 4. Buat Transaksi Jurnal Umum
	transaksi := models.Transaksi{
		EntitasID:   input.EntitasID,
		Tanggal:     tanggalStr,
		NoReferensi: fmt.Sprintf("GAJI-%02d-%d", input.Bulan, input.Tahun),
		Keterangan:  fmt.Sprintf("Pembayaran Gaji Pegawai Bulan %d Tahun %d", input.Bulan, input.Tahun),
		TotalDebit:  totalGaji,
		TotalKredit: totalGaji,
	}

	if err := tx.Create(&transaksi).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 5. Insert Jurnal Detail
	details := []models.JurnalDetail{
		{TransaksiID: transaksi.ID, AkunID: akunBeban.ID, Debit: totalGaji},
		{TransaksiID: transaksi.ID, AkunID: akunBankUtama.ID, RekeningBankID: &input.RekeningBankID, Kredit: totalGaji},
	}
	for _, det := range details {
		if err := tx.Create(&det).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Update Saldo Akun Buku Besar
		var akun models.AkunKeuangan
		if err := tx.First(&akun, det.AkunID).Error; err != nil {
			tx.Rollback()
			return err
		}
		if strings.EqualFold(akun.Sifat, "Debit") {
			akun.Saldo += det.Debit
			akun.Saldo -= det.Kredit
		} else {
			akun.Saldo += det.Kredit
			akun.Saldo -= det.Debit
		}
		if err := tx.Save(&akun).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Update Saldo Sub-Ledger Rekening Bank
	var rekBank models.RekeningBank
	if err := tx.First(&rekBank, input.RekeningBankID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("Gagal menemukan rekening bank yang dipilih")
	}
	rekBank.Saldo -= totalGaji
	if err := tx.Save(&rekBank).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 6. Catat riwayat gaji untuk setiap pegawai
	for _, p := range pegawais {
		riwayat := models.RiwayatGaji{
			PegawaiID:    p.ID,
			Bulan:        input.Bulan,
			Tahun:        input.Tahun,
			TotalGaji:    gajiAktualMap[p.ID],
			DetailTunjangan: tunjanganMap[p.ID],
			DetailPotongan:  potonganMap[p.ID],
			TanggalBayar: tanggalStr,
			TransaksiID:  transaksi.ID,
		}
		if err := tx.Create(&riwayat).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("Gagal membuat catatan riwayat gaji untuk pegawai %s", p.Nama)
		}
	}

	return tx.Commit().Error
}

// GetLaporanKonsolidasi menarik laporan dari entitas tertentu berdasarkan periode bulanan, kuartal, atau tahunan
func (a *App) GetLaporanKonsolidasi(mode string, periode int, tahun int, entitasID int) (models.LaporanKonsolidasiResponse, error) {
	var response models.LaporanKonsolidasiResponse
	var entitasList []models.Entitas
	query := database.DB
	if entitasID > 0 {
		query = query.Where("id = ?", entitasID)
	}
	if err := query.Find(&entitasList).Error; err != nil {
		return response, err
	}

	for _, e := range entitasList {
		response.EntitasNames = append(response.EntitasNames, e.Nama)
	}

	var startDate, endDate string
	switch strings.ToLower(mode) {
	case "bulanan":
		if periode < 1 || periode > 12 {
			periode = 7 // Default ke Juli jika tidak valid
		}
		startDate = fmt.Sprintf("%04d-%02d-01", tahun, periode)
		endDate = fmt.Sprintf("%04d-%02d-31", tahun, periode)
	case "kuartal":
		switch periode {
		case 1:
			startDate = fmt.Sprintf("%04d-01-01", tahun)
			endDate = fmt.Sprintf("%04d-03-31", tahun)
		case 2:
			startDate = fmt.Sprintf("%04d-04-01", tahun)
			endDate = fmt.Sprintf("%04d-06-30", tahun)
		case 3:
			startDate = fmt.Sprintf("%04d-07-01", tahun)
			endDate = fmt.Sprintf("%04d-09-30", tahun)
		case 4:
			startDate = fmt.Sprintf("%04d-10-01", tahun)
			endDate = fmt.Sprintf("%04d-12-31", tahun)
		default:
			startDate = fmt.Sprintf("%04d-07-01", tahun)
			endDate = fmt.Sprintf("%04d-09-30", tahun)
		}
	case "tahunan":
		fallthrough
	default:
		startDate = fmt.Sprintf("%04d-01-01", tahun)
		endDate = fmt.Sprintf("%04d-12-31", tahun)
	}

	newItem := func(uraian string, isRed bool) models.LaporanKonsolidasiItem {
		return models.LaporanKonsolidasiItem{
			Uraian: uraian,
			Values: make(map[string]int64),
			IsRed:  isRed,
		}
	}

	calcGlobal := func(item *models.LaporanKonsolidasiItem) {
		for _, v := range item.Values {
			item.Global += v
		}
	}

	calcTotalGroup := func(group *models.LaporanKonsolidasiGroup, uraian string, isBold bool) {
		total := newItem(uraian, false)
		total.IsBold = isBold
		for _, e := range entitasList {
			var sum int64
			for _, item := range group.Items {
				if item.IsRed {
					sum -= item.Values[e.Nama]
				} else {
					sum += item.Values[e.Nama]
				}
			}
			total.Values[e.Nama] = sum
		}
		calcGlobal(&total)
		group.Total = total
	}

	// NERACA
	asetLancar := models.LaporanKonsolidasiGroup{NamaGrup: "ASET LANCAR"}
	itemKas := newItem("Kas dan Setara Kas (Bank & Tunai)", false)
	itemPiutang := newItem("Piutang Usaha & SPP", false)
	itemPersediaan := newItem("Persediaan (Seragam & Buku)", false)

	liabilitas := models.LaporanKonsolidasiGroup{NamaGrup: "LIABILITAS (KEWAJIBAN / UTANG)"}
	itemUtangPdk := newItem("Utang Usaha & Utang Gaji", false)

	asetNeto := models.LaporanKonsolidasiGroup{NamaGrup: "ASET NETO (MODAL / KEKAYAAN BERSIH)"}
	itemNetoTB := newItem("Aset Neto Tanpa Pembatasan (Modal Awal & Operasional)", false)
	itemNetoDB := newItem("Aset Neto Dengan Pembatasan (Dana Wakaf & BOS)", false)
	itemSurplusSebelum := newItem("Akumulasi Surplus (Defisit) Periode Sebelumnya", false)

	// AKTIVITAS
	pendapatan := models.LaporanKonsolidasiGroup{NamaGrup: "PENDAPATAN"}
	itemSpp := newItem("SPP dan Uang Pangkal", false)
	itemBos := newItem("Dana BOS & Bantuan Operasional", false)
	itemWakaf := newItem("Sumbangan Infaq & Wakaf", false)
	itemPendLain := newItem("Pendapatan Lainnya (Seragam, Buku, Koperasi)", false)

	beban := models.LaporanKonsolidasiGroup{NamaGrup: "BEBAN (PENGELUARAN)"}
	itemGaji := newItem("Beban Gaji, THR & Insentif Guru/Staf", false)
	itemListrik := newItem("Beban Utilitas (Listrik, Air, & Internet)", false)
	itemBebanBos := newItem("Beban Operasional (Konsumsi, Kegiatan, ATK, Kebersihan)", false)

	// ARUS KAS - AKTIVITAS OPERASI
	arusOperasi := models.LaporanKonsolidasiGroup{NamaGrup: "ARUS KAS DARI AKTIVITAS OPERASI"}
	itemKasSpp := newItem("Penerimaan Kas SPP, Syahriyah & Uang Pangkal", false)
	itemKasBos := newItem("Penerimaan Kas Dana BOS & Bantuan Pemerintah", false)
	itemKasUsaha := newItem("Penerimaan Kas Usaha Koperasi, Seragam & Buku", false)
	itemKasGaji := newItem("Pembayaran Gaji, THR & Insentif Guru/Staf", true)
	itemKasKonsumsi := newItem("Pembayaran Belanja Dapur & Konsumsi Siswa", true)
	itemKasUtilitas := newItem("Pembayaran Utilitas (Listrik, Air & Internet)", true)
	itemKasOperasiLain := newItem("Pembayaran Kegiatan, UKS, ATK & Kebersihan", true)

	// ARUS KAS - AKTIVITAS INVESTASI
	arusInvestasi := models.LaporanKonsolidasiGroup{NamaGrup: "ARUS KAS DARI AKTIVITAS INVESTASI"}
	itemKasPembangunan := newItem("Pembayaran Pembangunan & Sarana Prasarana", true)
	itemKasAsrama := newItem("Pengadaan Perlengkapan Asrama & Furnitur", true)

	// ARUS KAS - AKTIVITAS PENDANAAN
	arusPendanaan := models.LaporanKonsolidasiGroup{NamaGrup: "ARUS KAS DARI AKTIVITAS PENDANAAN"}
	itemKasWakaf := newItem("Penerimaan Sumbangan Infaq, Shadaqah & Wakaf", false)

	// ARUS KAS - REKONSILIASI KAS & BANK
	arusRekonsiliasi := models.LaporanKonsolidasiGroup{NamaGrup: "RINGKASAN ARUS KAS & BANK"}
	itemKenaikanKas := newItem("Kenaikan (Penurunan) Kas Bersih", false)
	itemKasAwal := newItem("Saldo Kas & Setara Kas Awal Periode", false)

	for _, entitas := range entitasList {
		// Hitung surplus periode sebelumnya (< startDate) untuk entitas ini
		var pendSebelum, bebanSebelum struct {
			Net int64
		}
		database.DB.Table("jurnal_details").
			Joins("JOIN transaksis ON transaksis.id = jurnal_details.transaksi_id").
			Joins("JOIN akun_keuangans ON akun_keuangans.id = jurnal_details.akun_id").
			Where("transaksis.entitas_id = ? AND transaksis.tanggal < ? AND akun_keuangans.kategori = ?", entitas.ID, startDate, "Pendapatan").
			Select("COALESCE(SUM(jurnal_details.kredit - jurnal_details.debit), 0) AS net").
			Scan(&pendSebelum)

		database.DB.Table("jurnal_details").
			Joins("JOIN transaksis ON transaksis.id = jurnal_details.transaksi_id").
			Joins("JOIN akun_keuangans ON akun_keuangans.id = jurnal_details.akun_id").
			Where("transaksis.entitas_id = ? AND transaksis.tanggal < ? AND akun_keuangans.kategori = ?", entitas.ID, startDate, "Beban").
			Select("COALESCE(SUM(jurnal_details.debit - jurnal_details.kredit), 0) AS net").
			Scan(&bebanSebelum)

		itemSurplusSebelum.Values[entitas.Nama] = pendSebelum.Net - bebanSebelum.Net

		var akuns []models.AkunKeuangan
		database.DB.Where("entitas_id = ?", entitas.ID).Find(&akuns)

		for _, akun := range akuns {
			var saldo int64
			if akun.Kategori == "Pendapatan" {
				var res struct{ Net int64 }
				database.DB.Table("jurnal_details").
					Joins("JOIN transaksis ON transaksis.id = jurnal_details.transaksi_id").
					Where("jurnal_details.akun_id = ? AND transaksis.tanggal >= ? AND transaksis.tanggal <= ?", akun.ID, startDate, endDate).
					Select("COALESCE(SUM(jurnal_details.kredit - jurnal_details.debit), 0) AS net").
					Scan(&res)
				saldo = res.Net
			} else if akun.Kategori == "Beban" {
				var res struct{ Net int64 }
				database.DB.Table("jurnal_details").
					Joins("JOIN transaksis ON transaksis.id = jurnal_details.transaksi_id").
					Where("jurnal_details.akun_id = ? AND transaksis.tanggal >= ? AND transaksis.tanggal <= ?", akun.ID, startDate, endDate).
					Select("COALESCE(SUM(jurnal_details.debit - jurnal_details.kredit), 0) AS net").
					Scan(&res)
				saldo = res.Net
			} else {
				saldo = akun.Saldo
			}

			switch akun.Kategori {
			case "Aset":
				if akun.KodeAkun == "1110" || akun.KodeAkun == "1111" {
					itemKas.Values[entitas.Nama] += saldo
				} else if akun.KodeAkun == "1130" {
					itemPiutang.Values[entitas.Nama] += saldo
				} else {
					itemPersediaan.Values[entitas.Nama] += saldo
				}
			case "Liabilitas":
				itemUtangPdk.Values[entitas.Nama] += saldo
			case "Aset Neto":
				if akun.KodeAkun == "3120" {
					itemNetoDB.Values[entitas.Nama] += saldo
				} else {
					itemNetoTB.Values[entitas.Nama] += saldo
				}
			case "Pendapatan":
				if akun.KodeAkun == "4110" || akun.KodeAkun == "4150" {
					itemSpp.Values[entitas.Nama] += saldo
					itemKasSpp.Values[entitas.Nama] += saldo
				} else if akun.KodeAkun == "4160" {
					itemPendLain.Values[entitas.Nama] += saldo
					itemKasSpp.Values[entitas.Nama] += saldo
				} else if akun.KodeAkun == "4190" {
					itemBos.Values[entitas.Nama] += saldo
					itemKasBos.Values[entitas.Nama] += saldo
				} else if akun.KodeAkun == "4170" {
					itemWakaf.Values[entitas.Nama] += saldo
					itemKasWakaf.Values[entitas.Nama] += saldo
				} else {
					itemPendLain.Values[entitas.Nama] += saldo
					itemKasUsaha.Values[entitas.Nama] += saldo
				}
			case "Beban":
				if akun.KodeAkun == "5110" {
					itemGaji.Values[entitas.Nama] += saldo
					itemKasGaji.Values[entitas.Nama] += saldo
				} else if akun.KodeAkun == "5120" {
					itemBebanBos.Values[entitas.Nama] += saldo
					itemKasKonsumsi.Values[entitas.Nama] += saldo
				} else if akun.KodeAkun == "5190" {
					itemListrik.Values[entitas.Nama] += saldo
					itemKasUtilitas.Values[entitas.Nama] += saldo
				} else if akun.KodeAkun == "5170" {
					itemBebanBos.Values[entitas.Nama] += saldo
					itemKasPembangunan.Values[entitas.Nama] += saldo
				} else if akun.KodeAkun == "5180" {
					itemBebanBos.Values[entitas.Nama] += saldo
					itemKasAsrama.Values[entitas.Nama] += saldo
				} else {
					itemBebanBos.Values[entitas.Nama] += saldo
					itemKasOperasiLain.Values[entitas.Nama] += saldo
				}
			}
		}
	}

	// Hitung global untuk komponen Aktivitas terlebih dahulu
	calcGlobal(&itemSpp); calcGlobal(&itemBos); calcGlobal(&itemWakaf); calcGlobal(&itemPendLain)
	calcGlobal(&itemGaji); calcGlobal(&itemListrik); calcGlobal(&itemBebanBos)

	pendapatan.Items = []models.LaporanKonsolidasiItem{itemSpp, itemBos, itemWakaf, itemPendLain}
	calcTotalGroup(&pendapatan, "Total Pendapatan", true)

	beban.Items = []models.LaporanKonsolidasiItem{itemGaji, itemListrik, itemBebanBos}
	calcTotalGroup(&beban, "Total Beban", true)

	// Surplus / Defisit Periode Berjalan (Pendapatan - Beban) dimasukkan ke Aset Neto
	itemSurplus := newItem("Surplus (Defisit) Periode Berjalan", false)
	for _, e := range entitasList {
		itemSurplus.Values[e.Nama] = pendapatan.Total.Values[e.Nama] - beban.Total.Values[e.Nama]
	}
	calcGlobal(&itemSurplus)

	// Hitung global untuk komponen Neraca
	calcGlobal(&itemKas); calcGlobal(&itemPiutang); calcGlobal(&itemPersediaan)
	calcGlobal(&itemUtangPdk)
	calcGlobal(&itemNetoTB); calcGlobal(&itemNetoDB); calcGlobal(&itemSurplusSebelum)

	asetLancar.Items = []models.LaporanKonsolidasiItem{itemKas, itemPiutang, itemPersediaan}
	calcTotalGroup(&asetLancar, "Total Aset (Aktiva)", true)

	liabilitas.Items = []models.LaporanKonsolidasiItem{itemUtangPdk}
	calcTotalGroup(&liabilitas, "Total Liabilitas (Kewajiban)", true)

	asetNeto.Items = []models.LaporanKonsolidasiItem{itemNetoTB, itemNetoDB, itemSurplusSebelum, itemSurplus}
	calcTotalGroup(&asetNeto, "Total Aset Neto (Modal)", true)

	// Hitung global dan grup untuk Arus Kas
	calcGlobal(&itemKasSpp); calcGlobal(&itemKasBos); calcGlobal(&itemKasUsaha)
	calcGlobal(&itemKasGaji); calcGlobal(&itemKasKonsumsi); calcGlobal(&itemKasUtilitas); calcGlobal(&itemKasOperasiLain)
	calcGlobal(&itemKasPembangunan); calcGlobal(&itemKasAsrama)
	calcGlobal(&itemKasWakaf)

	arusOperasi.Items = []models.LaporanKonsolidasiItem{itemKasSpp, itemKasBos, itemKasUsaha, itemKasGaji, itemKasKonsumsi, itemKasUtilitas, itemKasOperasiLain}
	calcTotalGroup(&arusOperasi, "Arus Kas Bersih dari Aktivitas Operasi", true)

	arusInvestasi.Items = []models.LaporanKonsolidasiItem{itemKasPembangunan, itemKasAsrama}
	calcTotalGroup(&arusInvestasi, "Arus Kas Bersih dari Aktivitas Investasi", true)

	arusPendanaan.Items = []models.LaporanKonsolidasiItem{itemKasWakaf}
	calcTotalGroup(&arusPendanaan, "Arus Kas Bersih dari Aktivitas Pendanaan", true)

	// Rekonsiliasi Kas
	for _, e := range entitasList {
		kenaikan := arusOperasi.Total.Values[e.Nama] + arusInvestasi.Total.Values[e.Nama] + arusPendanaan.Total.Values[e.Nama]
		itemKenaikanKas.Values[e.Nama] = kenaikan
		saldoAkhir := itemKas.Values[e.Nama]
		itemKasAwal.Values[e.Nama] = saldoAkhir - kenaikan
	}
	calcGlobal(&itemKenaikanKas)
	calcGlobal(&itemKasAwal)

	arusRekonsiliasi.Items = []models.LaporanKonsolidasiItem{itemKenaikanKas, itemKasAwal}
	calcTotalGroup(&arusRekonsiliasi, "Saldo Kas & Setara Kas Akhir Periode", true)

	response.Neraca = []models.LaporanKonsolidasiGroup{asetLancar, liabilitas, asetNeto}
	response.Aktivitas = []models.LaporanKonsolidasiGroup{pendapatan, beban}
	response.ArusKas = []models.LaporanKonsolidasiGroup{arusOperasi, arusInvestasi, arusPendanaan, arusRekonsiliasi}

	return response, nil
}
