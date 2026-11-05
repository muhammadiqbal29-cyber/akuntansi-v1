package main

import (
	"fmt"
	"strings"

	"keuangan_nawawi/database"
	"keuangan_nawawi/models"

	"gorm.io/gorm"
)

// GenerateTagihanSPP membuat tagihan SPP rutin bulanan.
func (a *App) GenerateTagihanSPP(entitasID int, bulan int, tahun int) error {
	var siswaAktif []models.Siswa
	err := database.DB.Where("entitas_id = ? AND status = 'Aktif'", entitasID).Find(&siswaAktif).Error
	if err != nil {
		return fmt.Errorf("gagal mengambil data siswa: %v", err)
	}

	count := 0
	for _, siswa := range siswaAktif {
		var tagihan models.TagihanSiswa
		// Cek apakah tagihan SPP sudah ada untuk bulan & tahun ini
		err := database.DB.Where("siswa_id = ? AND jenis_tagihan = 'SPP' AND bulan = ? AND tahun = ?", siswa.ID, bulan, tahun).First(&tagihan).Error
		if err == gorm.ErrRecordNotFound {
			tagihanBaru := models.TagihanSiswa{
				EntitasID:    entitasID,
				SiswaID:      siswa.ID,
				JenisTagihan: "SPP",
				Bulan:        &bulan,
				Tahun:        &tahun,
				TotalTagihan: siswa.TarifSPP,
				SisaTagihan:  siswa.TarifSPP,
				Status:       "Belum Lunas",
			}
			database.DB.Create(&tagihanBaru)
			count++
		}
	}

	if count == 0 {
		return fmt.Errorf("semua siswa aktif sudah memiliki tagihan SPP untuk periode tersebut")
	}

	return nil
}

// GenerateTagihanBebasMassal membuat tagihan custom (Uang Gedung, Seragam, dll) untuk sekelompok siswa (atau semua siswa)
func (a *App) GenerateTagihanBebasMassal(entitasID int, targetKelas string, jenisTagihan string, nominal int64) error {
	query := database.DB.Where("entitas_id = ? AND status = 'Aktif'", entitasID)
	if targetKelas != "" {
		query = query.Where("kelas = ?", targetKelas)
	}

	var siswaList []models.Siswa
	if err := query.Find(&siswaList).Error; err != nil {
		return err
	}

	count := 0
	for _, siswa := range siswaList {
		var existing int64
		database.DB.Model(&models.TagihanSiswa{}).Where("siswa_id = ? AND jenis_tagihan = ?", siswa.ID, jenisTagihan).Count(&existing)
		
		if existing == 0 {
			tagihanBaru := models.TagihanSiswa{
				EntitasID:    entitasID,
				SiswaID:      siswa.ID,
				JenisTagihan: jenisTagihan,
				TotalTagihan: nominal,
				SisaTagihan:  nominal,
				Status:       "Belum Lunas",
			}
			database.DB.Create(&tagihanBaru)
			count++
		}
	}

	if count == 0 {
		return fmt.Errorf("Semua siswa di target tersebut sudah memiliki tagihan ini")
	}
	return nil
}

// GetTagihanSiswa mengambil semua tagihan untuk siswa tertentu.
func (a *App) GetTagihanSiswa(siswaID int) ([]models.TagihanSiswa, error) {
	var tagihan []models.TagihanSiswa
	err := database.DB.Where("siswa_id = ?", siswaID).Order("created_at DESC").Find(&tagihan).Error
	return tagihan, err
}

// GetTunggakanSiswa mengambil semua tagihan (termasuk yang Lunas dan Belum Lunas).
func (a *App) GetTunggakanSiswa(entitasID int) ([]models.TagihanSiswa, error) {
	var tunggakan []models.TagihanSiswa
	err := database.DB.Preload("Siswa").Where("entitas_id = ?", entitasID).Order("created_at ASC").Find(&tunggakan).Error
	return tunggakan, err
}

// BayarTagihanSiswa memproses pembayaran cicilan atau pelunasan tagihan.
func (a *App) BayarTagihanSiswa(tagihanID int, nominalBayar int64, akunPenerimaID int, tanggal string) error {
	var tagihan models.TagihanSiswa
	err := database.DB.Preload("Siswa").First(&tagihan, tagihanID).Error
	if err != nil {
		return fmt.Errorf("tagihan tidak ditemukan: %v", err)
	}

	if tagihan.Status == "Lunas" {
		return fmt.Errorf("tagihan ini sudah lunas sepenuhnya")
	}

	if nominalBayar <= 0 {
		return fmt.Errorf("nominal bayar harus lebih dari 0")
	}

	if nominalBayar != tagihan.SisaTagihan {
		return fmt.Errorf("pembayaran harus lunas/penuh (sebesar Rp %d) karena cicilan tidak diperbolehkan", tagihan.SisaTagihan)
	}

	// 1. Siapkan Akun Pendapatan spesifik untuk jenis tagihan ini
	akunPendapatanID, err := getOrCreateAkunPendapatan(tagihan.EntitasID, tagihan.JenisTagihan)
	if err != nil {
		return fmt.Errorf("gagal menyiapkan akun pendapatan %s: %v", tagihan.JenisTagihan, err)
	}

	// Mulai Transaksi Database
	tx := database.DB.Begin()

	// 2. Buat Transaksi Jurnal Umum
	var keterangan string
	if tagihan.JenisTagihan == "SPP" && tagihan.Bulan != nil && tagihan.Tahun != nil {
		keterangan = fmt.Sprintf("Pembayaran %s Bulan %d Tahun %d - %s (%s)", tagihan.JenisTagihan, *tagihan.Bulan, *tagihan.Tahun, tagihan.Siswa.NamaLengkap, tagihan.Siswa.NIS)
	} else {
		keterangan = fmt.Sprintf("Pembayaran %s - %s (%s)", tagihan.JenisTagihan, tagihan.Siswa.NamaLengkap, tagihan.Siswa.NIS)
	}
	
	transaksi := models.Transaksi{
		EntitasID:   tagihan.EntitasID,
		Tanggal:     tanggal,
		NoReferensi: fmt.Sprintf("INV/%d/%d", tagihan.Siswa.ID, tagihan.ID),
		Keterangan:  keterangan,
		TotalDebit:  nominalBayar,
		TotalKredit: nominalBayar,
	}

	if err := tx.Create(&transaksi).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal membuat transaksi jurnal: %v", err)
	}

	// 3. Detail Jurnal: Debit (Kas/Bank Penerima)
	jurnalDebit := models.JurnalDetail{
		TransaksiID: transaksi.ID,
		AkunID:      akunPenerimaID,
		Debit:       nominalBayar,
		Kredit:      0,
	}
	if err := tx.Create(&jurnalDebit).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal membuat detail jurnal (debit): %v", err)
	}

	// 4. Detail Jurnal: Kredit (Pendapatan Tagihan)
	jurnalKredit := models.JurnalDetail{
		TransaksiID: transaksi.ID,
		AkunID:      akunPendapatanID,
		Debit:       0,
		Kredit:      nominalBayar,
	}
	if err := tx.Create(&jurnalKredit).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal membuat detail jurnal (kredit): %v", err)
	}

	// 5. Perbarui saldo akun
	// Saldo Penerima (Kas/Bank)
	var akunPenerima models.AkunKeuangan
	if err := tx.First(&akunPenerima, akunPenerimaID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal mengambil data akun penerima: %v", err)
	}
	if akunPenerima.Sifat == "Debit" {
		akunPenerima.Saldo += nominalBayar
	} else {
		akunPenerima.Saldo -= nominalBayar
	}
	if err := tx.Save(&akunPenerima).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal update saldo akun kas/bank: %v", err)
	}

	// Saldo Pendapatan
	var akunPendapatan models.AkunKeuangan
	if err := tx.First(&akunPendapatan, akunPendapatanID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal mengambil data akun pendapatan: %v", err)
	}
	if akunPendapatan.Sifat == "Kredit" {
		akunPendapatan.Saldo += nominalBayar
	} else {
		akunPendapatan.Saldo -= nominalBayar
	}
	if err := tx.Save(&akunPendapatan).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal update saldo akun pendapatan: %v", err)
	}

	// 6. Buat Riwayat Pembayaran Tagihan
	riwayat := models.RiwayatPembayaranTagihan{
		TagihanSiswaID: tagihan.ID,
		NominalBayar:   nominalBayar,
		TanggalBayar:   tanggal,
		AkunPenerimaID: akunPenerimaID,
		TransaksiID:    transaksi.ID,
	}
	if err := tx.Create(&riwayat).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal membuat riwayat pembayaran: %v", err)
	}

	// 7. Update Status dan Sisa Tagihan Siswa
	tagihan.SisaTagihan -= nominalBayar
	if tagihan.SisaTagihan == 0 {
		tagihan.Status = "Lunas"
	} else {
		tagihan.Status = "Sebagian"
	}

	if err := tx.Save(&tagihan).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal memperbarui status tagihan: %v", err)
	}

	// Commit Transaksi
	return tx.Commit().Error
}

// mapJenisTagihanToAkun memetakan jenis tagihan ke kode akun dan nama akun yang sesuai
// berdasarkan Bagan Akun ISAK 35 yang sudah di-seed oleh GenerateISAK35.
func mapJenisTagihanToAkun(jenisTagihan string) (kodeAkun string, namaAkun string) {
	jenis := strings.ToLower(jenisTagihan)
	switch {
	case strings.Contains(jenis, "spp") || strings.Contains(jenis, "syahriyah"):
		return "4110", "Pendapatan SPP"
	case strings.Contains(jenis, "semester"):
		return "4120", "Pendapatan Semesteran"
	case strings.Contains(jenis, "seragam"):
		return "4130", "Pendapatan Seragam"
	case strings.Contains(jenis, "buku") || strings.Contains(jenis, "modul") || strings.Contains(jenis, "kitab") || strings.Contains(jenis, "lks"):
		return "4140", "Pendapatan Buku"
	case strings.Contains(jenis, "pangkal") || strings.Contains(jenis, "pembangunan") || strings.Contains(jenis, "wakaf sarana"):
		return "4150", "Pendapatan Uang Pangkal"
	case strings.Contains(jenis, "asrama") || strings.Contains(jenis, "boarding"):
		return "4160", "Pendapatan Asrama"
	case strings.Contains(jenis, "infaq") || strings.Contains(jenis, "infak") || strings.Contains(jenis, "shadaqah"):
		return "4170", "Pendapatan Infaq"
	case strings.Contains(jenis, "kegiatan") || strings.Contains(jenis, "ekstra"):
		return "4120", "Pendapatan Semesteran"
	default:
		return "4110", "Pendapatan SPP"
	}
}

func getOrCreateAkunPendapatan(entitasID int, jenisTagihan string) (int, error) {
	kodeAkun, namaAkun := mapJenisTagihanToAkun(jenisTagihan)

	var akun models.AkunKeuangan
	err := database.DB.Where("entitas_id = ? AND kode_akun = ?", entitasID, kodeAkun).First(&akun).Error
	if err == nil {
		return akun.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return 0, err
	}

	akunBaru := models.AkunKeuangan{
		EntitasID: entitasID,
		KodeAkun:  kodeAkun,
		NamaAkun:  namaAkun,
		Kategori:  "Pendapatan",
		Sifat:     "Kredit",
		Saldo:     0,
	}
	if err := database.DB.Create(&akunBaru).Error; err != nil {
		return 0, err
	}
	return akunBaru.ID, nil
}

// BayarBanyakTagihan memproses pembayaran massal (banyak tagihan) dalam 1 transaksi jurnal
func (a *App) BayarBanyakTagihan(tagihanIDs []int, rekeningBankID int, tanggal string) error {
	if len(tagihanIDs) == 0 {
		return fmt.Errorf("tidak ada tagihan yang dipilih")
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	var tagihans []models.TagihanSiswa
	if err := tx.Preload("Siswa").Where("id IN ?", tagihanIDs).Find(&tagihans).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal mengambil data tagihan: %v", err)
	}

	if len(tagihans) != len(tagihanIDs) {
		tx.Rollback()
		return fmt.Errorf("beberapa tagihan tidak ditemukan")
	}

	// Validasi awal
	siswaID := tagihans[0].SiswaID
	entitasID := tagihans[0].EntitasID
	namaSiswa := tagihans[0].Siswa.NamaLengkap
	nisSiswa := tagihans[0].Siswa.NIS

	var totalBayar int64
	totalPerJenis := make(map[string]int64)

	for i := range tagihans {
		if tagihans[i].SiswaID != siswaID {
			tx.Rollback()
			return fmt.Errorf("tagihan yang dipilih harus dari siswa yang sama")
		}
		if tagihans[i].Status == "Lunas" {
			tx.Rollback()
			return fmt.Errorf("tagihan %s sudah lunas", tagihans[i].JenisTagihan)
		}
		
		nominal := tagihans[i].SisaTagihan
		totalBayar += nominal
		totalPerJenis[tagihans[i].JenisTagihan] += nominal

		// Update status tagihan
		tagihans[i].SisaTagihan = 0
		tagihans[i].Status = "Lunas"
		tagihans[i].Siswa = nil // hindari GORM meng-cascade ke tabel siswa
		if err := tx.Save(&tagihans[i]).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("gagal update tagihan: %v", err)
		}
	}

	// 1. Buat Transaksi Jurnal Umum
	keterangan := fmt.Sprintf("Pembayaran %d Tagihan - %s (%s)", len(tagihans), namaSiswa, nisSiswa)
	transaksi := models.Transaksi{
		EntitasID:   entitasID,
		Tanggal:     tanggal,
		NoReferensi: fmt.Sprintf("INV/BATCH/%d/%d", siswaID, tagihans[0].ID),
		Keterangan:  keterangan,
		TotalDebit:  totalBayar,
		TotalKredit: totalBayar,
	}

	if err := tx.Create(&transaksi).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("gagal membuat transaksi jurnal: %v", err)
	}

	// 3. Update Saldo Kas/Bank (Rekening Bank)
	var rekBank models.RekeningBank
	if err := tx.First(&rekBank, rekeningBankID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("rekening bank tidak ditemukan: %v", err)
	}
	rekBank.Saldo += totalBayar
	if err := tx.Save(&rekBank).Error; err != nil {
		tx.Rollback()
		return err
	}

	// Ambil Master Akun Bank untuk jurnalnya
	var masterBank models.AkunKeuangan
	if err := tx.Where("entitas_id = ? AND is_bawaan = ? AND nama_akun = ?", entitasID, true, "Bank").First(&masterBank).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("akun master Bank tidak ditemukan: %v", err)
	}

	if masterBank.Sifat == "Debit" {
		masterBank.Saldo += totalBayar
	} else {
		masterBank.Saldo -= totalBayar
	}
	if err := tx.Save(&masterBank).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 2. Jurnal Debit (Kas/Bank)
	jurnalDebit := models.JurnalDetail{
		TransaksiID:    transaksi.ID,
		AkunID:         masterBank.ID,
		RekeningBankID: &rekBank.ID,
		Debit:          totalBayar,
		Kredit:         0,
	}
	if err := tx.Create(&jurnalDebit).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 4. Jurnal Kredit (Dipecah berdasarkan masing-masing Jenis Tagihan)
	for jenis, nominal := range totalPerJenis {
		akunPendapatanID, err := getOrCreateAkunPendapatanTx(tx, entitasID, jenis)
		if err != nil {
			tx.Rollback()
			return err
		}

		jurnalKredit := models.JurnalDetail{
			TransaksiID: transaksi.ID,
			AkunID:      akunPendapatanID,
			Debit:       0,
			Kredit:      nominal,
		}
		if err := tx.Create(&jurnalKredit).Error; err != nil {
			tx.Rollback()
			return err
		}

		// Update Saldo Pendapatan
		var akunPendapatan models.AkunKeuangan
		if err := tx.First(&akunPendapatan, akunPendapatanID).Error; err != nil {
			tx.Rollback()
			return err
		}
		if akunPendapatan.Sifat == "Kredit" {
			akunPendapatan.Saldo += nominal
		} else {
			akunPendapatan.Saldo -= nominal
		}
		if err := tx.Save(&akunPendapatan).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// 5. Buat Riwayat Pembayaran Tagihan
	for i := range tagihans {
		riwayat := models.RiwayatPembayaranTagihan{
			TagihanSiswaID: tagihans[i].ID,
			NominalBayar:   tagihans[i].TotalTagihan,
			TanggalBayar:   tanggal,
			AkunPenerimaID: masterBank.ID, // Simpan master bank
			TransaksiID:    transaksi.ID,
		}
		if err := tx.Create(&riwayat).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("gagal membuat riwayat pembayaran: %v", err)
		}
	}

	return tx.Commit().Error
}

func getOrCreateAkunPendapatanTx(tx *gorm.DB, entitasID int, jenisTagihan string) (int, error) {
	kodeAkun, namaAkun := mapJenisTagihanToAkun(jenisTagihan)

	var akun models.AkunKeuangan
	err := tx.Where("entitas_id = ? AND kode_akun = ?", entitasID, kodeAkun).First(&akun).Error
	if err == nil {
		return akun.ID, nil
	}

	if err != gorm.ErrRecordNotFound {
		return 0, err
	}

	akunBaru := models.AkunKeuangan{
		EntitasID: entitasID,
		KodeAkun:  kodeAkun,
		NamaAkun:  namaAkun,
		Kategori:  "Pendapatan",
		Sifat:     "Kredit",
		Saldo:     0,
	}
	if err := tx.Create(&akunBaru).Error; err != nil {
		return 0, err
	}
	return akunBaru.ID, nil
}
