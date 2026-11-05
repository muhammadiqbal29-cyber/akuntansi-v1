package database

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"keuangan_nawawi/models"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// getExeDir mengembalikan direktori tempat file executable berada (untuk deteksi migrasi).
func getExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		return "."
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "."
	}
	return filepath.Dir(exePath)
}

// GetAppDataDir mengembalikan folder penyimpanan data aplikasi standar desktop OS:
// - Windows : %APPDATA%\KeuanganNawawi (C:\Users\<User>\AppData\Roaming\KeuanganNawawi)
// - Linux   : ~/.config/KeuanganNawawi
// - macOS   : ~/Library/Application Support/KeuanganNawawi
func GetAppDataDir() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "."
		}
		configDir = homeDir
	}

	appDir := filepath.Join(configDir, "KeuanganNawawi")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		log.Printf("Gagal membuat direktori AppData: %v", err)
	}
	return appDir
}

// CopyFile menyalin file dari src ke dst
func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

// GetDbPath mengembalikan path lengkap file database utama
func GetDbPath() string {
	return filepath.Join(GetAppDataDir(), "keuangan_data.db")
}

// CheckpointWAL memaksa SQLite menulis seluruh log transaksi WAL ke file database utama
func CheckpointWAL() error {
	if DB != nil {
		return DB.Exec("PRAGMA wal_checkpoint(TRUNCATE);").Error
	}
	return nil
}

// CloseDB menutup koneksi database SQLite
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			return sqlDB.Close()
		}
	}
	return nil
}

func ConnectDB() {
	dbPath := GetDbPath()

	// Migrasi otomatis: jika di AppData belum ada database, tetapi di samping .exe ada database lama,
	// salin data lama tersebut ke AppData agar data user tidak hilang saat update.
	if fi, err := os.Stat(dbPath); os.IsNotExist(err) || (err == nil && fi.Size() == 0) {
		legacyPath := filepath.Join(getExeDir(), "keuangan_data.db")
		if legacyFi, err := os.Stat(legacyPath); err == nil && legacyFi.Size() > 0 {
			log.Printf("Memigrasikan database lama dari %s ke %s", legacyPath, dbPath)
			if err := CopyFile(legacyPath, dbPath); err != nil {
				log.Printf("Peringatan: Gagal menyalin database lama: %v", err)
			} else {
				log.Println("Migrasi database lama ke AppData berhasil.")
			}
		}
	}

	log.Printf("Menggunakan database di: %s", dbPath)

	database, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Gagal terhubung ke database: %v", err))
	}

	// Optimasi performa & reliabilitas SQLite (Poin 3):
	database.Exec("PRAGMA journal_mode = WAL;")  // Write-Ahead Logging (baca & tulis bisa simultan)
	database.Exec("PRAGMA synchronous = NORMAL;") // Keseimbangan ideal antara kecepatan disk & keselamatan data
	database.Exec("PRAGMA busy_timeout = 5000;")  // Tunggu hingga 5 detik jika ada antrian tulis
	database.Exec("PRAGMA foreign_keys = ON;")    // Integritas relasional antar tabel

	// Konfigurasi pool koneksi: serialisasi akses tulis SQLite agar tidak terjadi race condition
	sqlDB, err := database.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(1)
		sqlDB.SetMaxIdleConns(1)
	}

	// Hapus indeks unik global lama pada kode_akun jika ada (sekarang memakai indeks komposit dengan entitas_id)
	database.Exec("DROP INDEX IF EXISTS idx_akun_keuangans_kode_akun;")

	// Auto-Migrate tabel sesuai ISAK35 & Modul Operasional
	err = database.AutoMigrate(
		&models.Entitas{},
		&models.AkunKeuangan{},
		&models.RekeningBank{},
		&models.Transaksi{},
		&models.JurnalDetail{},
		&models.Pegawai{},
		&models.RiwayatGaji{},
		&models.Siswa{},
		&models.TagihanSiswa{},
		&models.RiwayatPembayaranTagihan{},
	)
	if err != nil {
		panic(fmt.Sprintf("Gagal migrasi database: %v", err))
	}

	DB = database
	log.Println("Koneksi database berhasil dan termigrasi (SQLite).")
}
