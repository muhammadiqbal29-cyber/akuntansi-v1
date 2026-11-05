package models

import "time"

type Entitas struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Nama      string    `gorm:"type:varchar(255);not null;unique" json:"nama"`
	Tipe      string    `gorm:"type:varchar(50);not null" json:"tipe"` // "Perusahaan" or "Sekolah"
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

type AkunKeuangan struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	EntitasID int       `gorm:"uniqueIndex:idx_entitas_kode;not null" json:"entitas_id"`
	KodeAkun  string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_entitas_kode" json:"kode_akun"`
	NamaAkun  string    `gorm:"type:varchar(255);not null" json:"nama_akun"`
	Kategori  string    `gorm:"type:varchar(50);not null" json:"kategori"`
	Sifat     string    `gorm:"type:varchar(20);not null" json:"sifat"` // Saldo normal
	IsBawaan  bool      `gorm:"type:boolean;default:false" json:"is_bawaan"`
	Saldo     int64     `gorm:"type:integer;not null;default:0" json:"saldo"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type RekeningBank struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	EntitasID     int       `gorm:"index;not null" json:"entitas_id"`
	NamaRekening  string    `gorm:"type:varchar(255);not null" json:"nama_rekening"`
	NomorRekening string    `gorm:"type:varchar(100)" json:"nomor_rekening"`
	Saldo         int64     `gorm:"type:integer;not null;default:0" json:"saldo"`
	CreatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt     time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type Transaksi struct {
	ID          int       `gorm:"primaryKey;autoIncrement" json:"id"`
	EntitasID   int       `gorm:"index;not null" json:"entitas_id"`
	Tanggal     string    `gorm:"type:date;not null" json:"tanggal"` // YYYY-MM-DD
	NoReferensi string    `gorm:"type:varchar(100)" json:"no_referensi"`
	Keterangan  string    `gorm:"type:text;not null" json:"keterangan"`
	TotalDebit  int64     `gorm:"type:integer;not null;default:0" json:"total_debit"`
	TotalKredit int64     `gorm:"type:integer;not null;default:0" json:"total_kredit"`
	CreatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type JurnalDetail struct {
	ID             int           `gorm:"primaryKey;autoIncrement" json:"id"`
	TransaksiID    int           `gorm:"index;not null" json:"transaksi_id"`
	Transaksi      *Transaksi    `gorm:"foreignKey:TransaksiID;constraint:OnDelete:CASCADE" json:"transaksi,omitempty"`
	AkunID         int           `gorm:"index;not null" json:"akun_id"`
	Akun           *AkunKeuangan `gorm:"foreignKey:AkunID" json:"akun,omitempty"`
	RekeningBankID *int          `gorm:"index" json:"rekening_bank_id,omitempty"` // ID sub-ledger Kas/Bank jika relevan
	RekeningBank   *RekeningBank `gorm:"foreignKey:RekeningBankID" json:"rekening_bank,omitempty"`
	Debit          int64         `gorm:"type:integer;not null;default:0" json:"debit"`
	Kredit         int64         `gorm:"type:integer;not null;default:0" json:"kredit"`
}

// --- MODUL PENGGAJIAN ---
type Pegawai struct {
	ID                int       `gorm:"primaryKey;autoIncrement" json:"id"`
	EntitasID         int       `gorm:"index;not null" json:"entitas_id"`
	NIP               string    `gorm:"type:varchar(50)" json:"nip"`
	Nama              string    `gorm:"type:varchar(255);not null" json:"nama"`
	Jabatan           string    `gorm:"type:varchar(100)" json:"jabatan"`
	GajiPokok         int64     `gorm:"type:integer;not null;default:0" json:"gaji_pokok"`
	TanggalMulaiKerja string    `gorm:"type:date;not null" json:"tanggal_mulai_kerja"` // YYYY-MM-DD
	PeriodeKenaikan   int       `gorm:"type:int;not null;default:1" json:"periode_kenaikan"`   // Per berapa tahun
	NominalKenaikan   int64     `gorm:"type:integer;not null;default:0" json:"nominal_kenaikan"` // Kenaikan per periode
	Tunjangan         string    `gorm:"type:text" json:"tunjangan"` // JSON array string
	Potongan          string    `gorm:"type:text" json:"potongan"`  // JSON array string
	CreatedAt         time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type RiwayatGaji struct {
	ID              int       `gorm:"primaryKey;autoIncrement" json:"id"`
	PegawaiID       int       `gorm:"index;not null" json:"pegawai_id"`
	Pegawai         *Pegawai  `gorm:"foreignKey:PegawaiID;constraint:OnDelete:CASCADE" json:"pegawai,omitempty"`
	Bulan           int       `gorm:"type:int;not null" json:"bulan"` // 1-12
	Tahun           int       `gorm:"type:int;not null" json:"tahun"` // e.g., 2026
	TotalGaji       int64     `gorm:"type:integer;not null;default:0" json:"total_gaji"`
	DetailTunjangan string    `gorm:"type:text" json:"detail_tunjangan"`
	DetailPotongan  string    `gorm:"type:text" json:"detail_potongan"`
	TanggalBayar    string    `gorm:"type:date;not null" json:"tanggal_bayar"`
	TransaksiID     int       `gorm:"index;not null" json:"transaksi_id"` // Terhubung ke Jurnal Umum
	CreatedAt       time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// Modul Siswa
type Siswa struct {
	ID           int       `gorm:"primaryKey;autoIncrement" json:"id"`
	EntitasID    int       `gorm:"index;not null" json:"entitas_id"`
	NIS          string    `gorm:"type:varchar(20);not null" json:"nis"`
	NISN         string    `gorm:"type:varchar(10)" json:"nisn"`
	NamaLengkap  string    `gorm:"type:varchar(100);not null" json:"nama_lengkap"`
	JenisKelamin string    `gorm:"type:varchar(10);not null" json:"jenis_kelamin"` // "L" atau "P"
	TempatLahir  string    `gorm:"type:varchar(50)" json:"tempat_lahir"`
	TanggalLahir string    `gorm:"type:date" json:"tanggal_lahir"`
	Agama        string    `gorm:"type:varchar(20)" json:"agama"`
	Alamat       string    `gorm:"type:text" json:"alamat"`
	NamaAyah     string    `gorm:"type:varchar(100)" json:"nama_ayah"`
	NamaIbu      string    `gorm:"type:varchar(100)" json:"nama_ibu"`
	TeleponOrtu  string    `gorm:"type:varchar(15)" json:"telepon_ortu"`
	Kelas        string    `gorm:"type:varchar(20)" json:"kelas"`   // misal: "7A", "1B"
	Jenjang      string    `gorm:"type:varchar(10)" json:"jenjang"` // "SD" atau "SMP"
	TarifSPP     int64     `gorm:"type:integer;not null;default:0" json:"tarif_spp"`
	Status       string    `gorm:"type:varchar(20);default:'Aktif'" json:"status"` // "Aktif" / "Keluar"
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type TagihanSiswa struct {
	ID           int    `gorm:"primaryKey;autoIncrement" json:"id"`
	EntitasID    int    `gorm:"index;not null" json:"entitas_id"`
	SiswaID      int    `gorm:"index;not null" json:"siswa_id"`
	Siswa        *Siswa `gorm:"foreignKey:SiswaID;constraint:OnDelete:CASCADE" json:"siswa,omitempty"`
	JenisTagihan string `gorm:"type:varchar(100);not null" json:"jenis_tagihan"` // e.g., "SPP", "Uang Pangkal"
	Bulan        *int   `gorm:"type:int" json:"bulan,omitempty"`                 // null untuk tagihan bebas
	Tahun        *int   `gorm:"type:int" json:"tahun,omitempty"`
	TotalTagihan int64  `gorm:"type:integer;not null;default:0" json:"total_tagihan"`
	SisaTagihan  int64  `gorm:"type:integer;not null;default:0" json:"sisa_tagihan"`
	Status       string `gorm:"type:varchar(20);not null;default:'Belum Lunas'" json:"status"` // "Belum Lunas", "Sebagian", "Lunas"
	CreatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

type RiwayatPembayaranTagihan struct {
	ID             int       `gorm:"primaryKey;autoIncrement" json:"id"`
	TagihanSiswaID int       `gorm:"index;not null" json:"tagihan_siswa_id"`
	NominalBayar   int64     `gorm:"type:integer;not null;default:0" json:"nominal_bayar"`
	TanggalBayar   string    `gorm:"type:date;not null" json:"tanggal_bayar"`
	AkunPenerimaID int       `gorm:"index;not null" json:"akun_penerima_id"` // Ke mana uang masuk (Kas/Bank)
	TransaksiID    int       `gorm:"index;not null" json:"transaksi_id"`     // Terhubung ke Jurnal Umum
	CreatedAt      time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP" json:"created_at"`
}

type LaporanKonsolidasiItem struct {
	Uraian string           `json:"uraian"`
	Values map[string]int64 `json:"values"` // Map dari Nama Entitas -> Nilai (Rupiah bulat)
	Global int64            `json:"global"`
	IsBold bool             `json:"is_bold"`
	IsRed  bool             `json:"is_red"`
}

type LaporanKonsolidasiGroup struct {
	NamaGrup string                   `json:"nama_grup"`
	Items    []LaporanKonsolidasiItem `json:"items"`
	Total    LaporanKonsolidasiItem   `json:"total"` // Total baris
}

type LaporanKonsolidasiResponse struct {
	EntitasNames []string                  `json:"entitas_names"`
	Neraca       []LaporanKonsolidasiGroup `json:"neraca"`
	Aktivitas    []LaporanKonsolidasiGroup `json:"aktivitas"`
	ArusKas      []LaporanKonsolidasiGroup `json:"arus_kas"`
}
