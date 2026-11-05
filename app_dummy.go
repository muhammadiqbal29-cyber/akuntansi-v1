package main

import (
	"fmt"
	"log"

	"keuangan_nawawi/database"
	"keuangan_nawawi/models"

	"gorm.io/gorm"
)

func getAkunByKode(entitasID int, kode string) (*models.AkunKeuangan, error) {
	var akun models.AkunKeuangan
	err := database.DB.Where("entitas_id = ? AND kode_akun = ?", entitasID, kode).First(&akun).Error
	return &akun, err
}

// GenerateDataDummy membuat data simulasi komplit Sekolah Boarding School selama 1 tahun penuh (Juli 2025 s/d Juni 2026)
func (a *App) GenerateDataDummy() error {
	log.Println("Memulai generate data dummy Sekolah Boarding School selama 1 Tahun Penuh (Juli 2025 - Juni 2026)...")

	// 1. Bersihkan data lama
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.RiwayatPembayaranTagihan{})
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.TagihanSiswa{})
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Siswa{})
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.RiwayatGaji{})
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Pegawai{})
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.JurnalDetail{})
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Transaksi{})
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.RekeningBank{})
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.AkunKeuangan{})
	database.DB.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Entitas{})
	database.DB.Exec("DELETE FROM sqlite_sequence;") // Reset autoincrement ID

	// =========================================================================
	// ENTITAS 1: SMP ISLAM IMAM NAWAWI (UNIT SEKOLAH / AKADEMIK)
	// =========================================================================
	smp, err := a.CreateEntitas("SMP Islam Imam Nawawi", "Sekolah")
	if err != nil {
		return fmt.Errorf("gagal membuat entitas SMP: %v", err)
	}
	if err := a.GenerateISAK35(smp.ID); err != nil {
		return fmt.Errorf("gagal generate ISAK35 SMP: %v", err)
	}

	// Rekening Bank & Kas Sekolah
	rekBSISMP, _ := a.CreateRekeningBank(models.RekeningBank{
		EntitasID:     smp.ID,
		NamaRekening:  "Bank BSI - Operasional Sekolah",
		NomorRekening: "7123456789",
		Saldo:         85000000,
	})
	_, _ = a.CreateRekeningBank(models.RekeningBank{
		EntitasID:     smp.ID,
		NamaRekening:  "Bank Muamalat - SPP & Kegiatan",
		NomorRekening: "0001239876",
		Saldo:         45000000,
	})

	// Saldo Kas Tunai & Modal Awal Sekolah
	akunKasSMP, _ := getAkunByKode(smp.ID, "1111")
	if akunKasSMP != nil {
		akunKasSMP.Saldo = 15000000
		database.DB.Save(akunKasSMP)
	}
	akunBankSMP, _ := getAkunByKode(smp.ID, "1110")
	akunModalSMP, _ := getAkunByKode(smp.ID, "3110")
	if akunModalSMP != nil && akunBankSMP != nil {
		akunModalSMP.Saldo += 145000000 // 85jt BSI + 45jt Muamalat + 15jt Kas
		database.DB.Save(akunModalSMP)
	}

	// Data Siswa SMP
	siswasSMP := []models.Siswa{
		{EntitasID: smp.ID, NIS: "250701", NISN: "0130000001", NamaLengkap: "Muhammad Al-Fatih", JenisKelamin: "L", TempatLahir: "Jakarta", TanggalLahir: "2013-05-12", Agama: "Islam", Alamat: "Jl. Merdeka Raya No. 10, Jakarta", NamaAyah: "Hendra Wijaya", NamaIbu: "Aisyah", TeleponOrtu: "081234560001", Kelas: "7A", Jenjang: "SMP", TarifSPP: 450000, Status: "Aktif"},
		{EntitasID: smp.ID, NIS: "250702", NISN: "0130000002", NamaLengkap: "Zahra Nuraini", JenisKelamin: "P", TempatLahir: "Bandung", TanggalLahir: "2013-08-20", Agama: "Islam", Alamat: "Jl. Anggrek Asri No. 2, Jakarta", NamaAyah: "Rahmat Hidayat", NamaIbu: "Fatmawati", TeleponOrtu: "081234560002", Kelas: "7A", Jenjang: "SMP", TarifSPP: 450000, Status: "Aktif"},
		{EntitasID: smp.ID, NIS: "250703", NISN: "0130000003", NamaLengkap: "Ahmad Zaki", JenisKelamin: "L", TempatLahir: "Bekasi", TanggalLahir: "2013-02-14", Agama: "Islam", Alamat: "Jl. Patriot No. 5, Bekasi", NamaAyah: "Sulaiman", NamaIbu: "Halimah", TeleponOrtu: "081234560003", Kelas: "7B", Jenjang: "SMP", TarifSPP: 450000, Status: "Aktif"},
		{EntitasID: smp.ID, NIS: "250704", NISN: "0130000004", NamaLengkap: "Nadiya Sakinah", JenisKelamin: "P", TempatLahir: "Depok", TanggalLahir: "2013-11-19", Agama: "Islam", Alamat: "Jl. Margonda No. 88, Depok", NamaAyah: "Lukman Hakim", NamaIbu: "Rosmaniar", TeleponOrtu: "081234560004", Kelas: "7B", Jenjang: "SMP", TarifSPP: 450000, Status: "Aktif"},
		{EntitasID: smp.ID, NIS: "250705", NISN: "0120000005", NamaLengkap: "Yusuf Ramadhan", JenisKelamin: "L", TempatLahir: "Surabaya", TanggalLahir: "2012-10-15", Agama: "Islam", Alamat: "Jl. Kebon Jeruk No. 45, Jakarta", NamaAyah: "Bambang Sugita", NamaIbu: "Siti Aminah", TeleponOrtu: "081234560005", Kelas: "8A", Jenjang: "SMP", TarifSPP: 450000, Status: "Aktif"},
		{EntitasID: smp.ID, NIS: "250706", NISN: "0120000006", NamaLengkap: "Aisyah Humaira", JenisKelamin: "P", TempatLahir: "Bogor", TanggalLahir: "2012-04-18", Agama: "Islam", Alamat: "Jl. Melati Mekar No. 12, Depok", NamaAyah: "Fauzan Alawi", NamaIbu: "Mariam", TeleponOrtu: "081234560006", Kelas: "8A", Jenjang: "SMP", TarifSPP: 450000, Status: "Aktif"},
		{EntitasID: smp.ID, NIS: "250707", NISN: "0120000007", NamaLengkap: "Umar bin Khattab Putra", JenisKelamin: "L", TempatLahir: "Tangerang", TanggalLahir: "2012-07-21", Agama: "Islam", Alamat: "Jl. BSD Raya No. 19, Tangerang", NamaAyah: "Anwar Sanusi", NamaIbu: "Khadijah", TeleponOrtu: "081234560007", Kelas: "8B", Jenjang: "SMP", TarifSPP: 450000, Status: "Aktif"},
		{EntitasID: smp.ID, NIS: "250708", NISN: "0120000008", NamaLengkap: "Siti Fatimah", JenisKelamin: "P", TempatLahir: "Jakarta", TanggalLahir: "2012-09-09", Agama: "Islam", Alamat: "Jl. Tebet Barat No. 3, Jakarta", NamaAyah: "Ridwan Kamil", NamaIbu: "Atalia", TeleponOrtu: "081234560008", Kelas: "8B", Jenjang: "SMP", TarifSPP: 450000, Status: "Aktif"},
		{EntitasID: smp.ID, NIS: "250709", NISN: "0110000009", NamaLengkap: "Bilal bin Rabah", JenisKelamin: "L", TempatLahir: "Bekasi", TanggalLahir: "2011-11-03", Agama: "Islam", Alamat: "Jl. Mawar Raya No. 8, Bekasi", NamaAyah: "Abdullah", NamaIbu: "Aminah", TeleponOrtu: "081234560009", Kelas: "9A", Jenjang: "SMP", TarifSPP: 450000, Status: "Aktif"},
		{EntitasID: smp.ID, NIS: "250710", NISN: "0110000010", NamaLengkap: "Firdaus Azzahra", JenisKelamin: "P", TempatLahir: "Jakarta", TanggalLahir: "2011-12-25", Agama: "Islam", Alamat: "Jl. Dahlia Selatan No. 15, Tangerang", NamaAyah: "Ali bin Abi Thalib", NamaIbu: "Safiyyah", TeleponOrtu: "081234560010", Kelas: "9A", Jenjang: "SMP", TarifSPP: 450000, Status: "Aktif"},
	}
	for _, s := range siswasSMP {
		_, _ = a.CreateSiswa(s)
	}

	// Data Pegawai & Guru SMP
	pegawaisSMP := []PegawaiInput{
		{EntitasID: smp.ID, NIP: "19850101202001", Nama: "Ustadz Ahmad Abdullah, M.Pd.", Jabatan: "Kepala Sekolah & Guru Tahfiz", GajiPokok: 4800000, TanggalMulaiKerja: "2021-07-01", PeriodeKenaikan: 1, NominalKenaikan: 250000, Tunjangan: `[{"nama":"Tunjangan Jabatan","nominal":1200000}]`, Potongan: `[{"nama":"BPJS","nominal":180000}]`},
		{EntitasID: smp.ID, NIP: "19900215202107", Nama: "Siti Khadijah, S.Pd.", Jabatan: "Guru Matematika & Wali Kelas 7A", GajiPokok: 3600000, TanggalMulaiKerja: "2022-01-10", PeriodeKenaikan: 1, NominalKenaikan: 200000, Tunjangan: `[{"nama":"Tunjangan Wali Kelas","nominal":500000}]`, Potongan: `[{"nama":"BPJS","nominal":130000}]`},
		{EntitasID: smp.ID, NIP: "19930420202301", Nama: "Budi Santoso, S.Kom.", Jabatan: "Staf Tata Usaha & IT", GajiPokok: 3200000, TanggalMulaiKerja: "2023-01-05", PeriodeKenaikan: 2, NominalKenaikan: 150000, Tunjangan: `[{"nama":"Tunjangan IT","nominal":500000}]`, Potongan: `[{"nama":"BPJS","nominal":110000}]`},
	}
	for _, p := range pegawaisSMP {
		_ = a.CreatePegawai(p)
	}

	// =========================================================================
	// ENTITAS 2: YAYASAN PONDOK PESANTREN IMAM NAWAWI (UNIT ASRAMA & BOARDING)
	// =========================================================================
	yayasan, err := a.CreateEntitas("Yayasan Pondok Pesantren Imam Nawawi", "Perusahaan")
	if err != nil {
		return fmt.Errorf("gagal membuat entitas Yayasan: %v", err)
	}
	if err := a.GenerateISAK35(yayasan.ID); err != nil {
		return fmt.Errorf("gagal generate ISAK35 Yayasan: %v", err)
	}

	// Rekening Bank & Kas Yayasan
	rekBSIYys, _ := a.CreateRekeningBank(models.RekeningBank{
		EntitasID:     yayasan.ID,
		NamaRekening:  "Bank BSI - Wakaf & Infaq Pondok",
		NomorRekening: "7999888777",
		Saldo:         250000000,
	})
	rekMandiriYys, _ := a.CreateRekeningBank(models.RekeningBank{
		EntitasID:     yayasan.ID,
		NamaRekening:  "Bank Mandiri - Operasional Asrama",
		NomorRekening: "123000999888",
		Saldo:         120000000,
	})

	akunKasYys, _ := getAkunByKode(yayasan.ID, "1111")
	if akunKasYys != nil {
		akunKasYys.Saldo = 25000000
		database.DB.Save(akunKasYys)
	}
	akunBankYys, _ := getAkunByKode(yayasan.ID, "1110")
	akunModalYys, _ := getAkunByKode(yayasan.ID, "3110")
	if akunModalYys != nil && akunBankYys != nil {
		akunModalYys.Saldo += 395000000 // 250jt + 120jt + 25jt
		database.DB.Save(akunModalYys)
	}

	// Data Santri Boarding / Asrama
	santrisYys := []models.Siswa{
		{EntitasID: yayasan.ID, NIS: "PONDOK-001", NISN: "9900001001", NamaLengkap: "Abdullah Azzam (Santri Asrama)", JenisKelamin: "L", TempatLahir: "Solo", TanggalLahir: "2010-03-10", Agama: "Islam", Alamat: "Asrama Putra Gedung A No. 1", NamaAyah: "Sulaiman Rasyidin", TeleponOrtu: "081122330001", Kelas: "Asrama Putra", Jenjang: "Pesantren", TarifSPP: 850000, Status: "Aktif"},
		{EntitasID: yayasan.ID, NIS: "PONDOK-002", NISN: "9900001002", NamaLengkap: "Hafizhuddin Al-Mubarak", JenisKelamin: "L", TempatLahir: "Malang", TanggalLahir: "2009-08-15", Agama: "Islam", Alamat: "Asrama Putra Gedung A No. 2", NamaAyah: "Mukhlis", TeleponOrtu: "081122330002", Kelas: "Asrama Putra", Jenjang: "Pesantren", TarifSPP: 850000, Status: "Aktif"},
		{EntitasID: yayasan.ID, NIS: "PONDOK-003", NISN: "9900001003", NamaLengkap: "Ahmad Tsanaawi (Ma'had)", JenisKelamin: "L", TempatLahir: "Semarang", TanggalLahir: "2010-01-20", Agama: "Islam", Alamat: "Asrama Putra Gedung B No. 1", NamaAyah: "Darmawan", TeleponOrtu: "081122330003", Kelas: "Asrama Putra", Jenjang: "Pesantren", TarifSPP: 850000, Status: "Aktif"},
		{EntitasID: yayasan.ID, NIS: "PONDOK-004", NISN: "9900001004", NamaLengkap: "Khadijah Al-Kubra (Santri Putri)", JenisKelamin: "P", TempatLahir: "Jakarta", TanggalLahir: "2010-11-05", Agama: "Islam", Alamat: "Asrama Putri Gedung A No. 1", NamaAyah: "Usman", TeleponOrtu: "081122330004", Kelas: "Asrama Putri", Jenjang: "Pesantren", TarifSPP: 850000, Status: "Aktif"},
		{EntitasID: yayasan.ID, NIS: "PONDOK-005", NISN: "9900001005", NamaLengkap: "Sumayyah binti Khayyat", JenisKelamin: "P", TempatLahir: "Cirebon", TanggalLahir: "2009-12-12", Agama: "Islam", Alamat: "Asrama Putri Gedung A No. 2", NamaAyah: "Ammar", TeleponOrtu: "081122330005", Kelas: "Asrama Putri", Jenjang: "Pesantren", TarifSPP: 850000, Status: "Aktif"},
		{EntitasID: yayasan.ID, NIS: "PONDOK-006", NISN: "9900001006", NamaLengkap: "Zaynab binti Jahsy", JenisKelamin: "P", TempatLahir: "Tasikmalaya", TanggalLahir: "2010-06-18", Agama: "Islam", Alamat: "Asrama Putri Gedung B No. 1", NamaAyah: "Jalaluddin", TeleponOrtu: "081122330006", Kelas: "Asrama Putri", Jenjang: "Pesantren", TarifSPP: 850000, Status: "Aktif"},
		{EntitasID: yayasan.ID, NIS: "PONDOK-007", NISN: "9900001007", NamaLengkap: "Salman Al-Farisi (Santri Mandiri)", JenisKelamin: "L", TempatLahir: "Yogyakarta", TanggalLahir: "2008-05-11", Agama: "Islam", Alamat: "Asrama Putra Gedung B No. 2", NamaAyah: "Faridudin", TeleponOrtu: "081122330007", Kelas: "Asrama Putra", Jenjang: "Pesantren", TarifSPP: 850000, Status: "Aktif"},
		{EntitasID: yayasan.ID, NIS: "PONDOK-008", NISN: "9900001008", NamaLengkap: "Huwairah binti Abu Hurairah", JenisKelamin: "P", TempatLahir: "Bogor", TanggalLahir: "2009-04-14", Agama: "Islam", Alamat: "Asrama Putri Gedung B No. 2", NamaAyah: "Abdullah Syakhir", TeleponOrtu: "081122330008", Kelas: "Asrama Putri", Jenjang: "Pesantren", TarifSPP: 850000, Status: "Aktif"},
	}
	for _, st := range santrisYys {
		_, _ = a.CreateSiswa(st)
	}

	// Data Pegawai & Pengasuh Pondok
	pegawaisYys := []PegawaiInput{
		{EntitasID: yayasan.ID, NIP: "PONDOK-202001", Nama: "K.H. Muhammad Syadir, M.A.", Jabatan: "Pimpinan Pondok Pesantren", GajiPokok: 7500000, TanggalMulaiKerja: "2020-01-01", PeriodeKenaikan: 1, NominalKenaikan: 500000, Tunjangan: `[{"nama":"Tunjangan Pengabdian","nominal":2500000}]`, Potongan: `[{"nama":"BPJS","nominal":200000}]`},
		{EntitasID: yayasan.ID, NIP: "PONDOK-202105", Nama: "Abdurrahman Al-Hafidz, S.E.", Jabatan: "Manajer Keuangan & Dapur Asrama", GajiPokok: 5500000, TanggalMulaiKerja: "2021-05-15", PeriodeKenaikan: 1, NominalKenaikan: 300000, Tunjangan: `[{"nama":"Tunjangan Kinerja","nominal":1500000}]`, Potongan: `[{"nama":"BPJS","nominal":160000}]`},
	}
	for _, p := range pegawaisYys {
		_ = a.CreatePegawai(p)
	}

	// =========================================================================
	// SIMULASI TRANSAKSI 1 TAHUN AJARAN PENUH (JULI 2025 - JUNI 2026)
	// =========================================================================
	type BulanKegiatan struct {
		Bulan      int
		Tahun      int
		Nama       string
		HariAkhir  int
		EventSMP   string
		BiayaSMP   int64
		EventYys   string
		BiayaYys   int64
	}

	bulanTahunList := []BulanKegiatan{
		{7, 2025, "Juli 2025", 31, "Masa Pengenalan Lingkungan Sekolah (MPLS) & Orientasi Siswa Baru", 3500000, "Pengadaan Perabotan & Pembersihan Awal Tahun Asrama", 12500000},
		{8, 2025, "Agustus 2025", 31, "Semarak Lomba Peringatan Hari Kemerdekaan RI di Sekolah", 2800000, "Lomba Kebersihan Kamar Asrama & Perlombaan Diniyah", 1500000},
		{9, 2025, "September 2025", 30, "Pelaksanaan Ujian Tengah Semester (PTS) Ganjil", 3200000, "Kajian Akbar Wali Santri & Syukuran Hifdzil Qur'an", 4500000},
		{10, 2025, "Oktober 2025", 31, "Kegiatan Kemah Santri & Peringatan Hari Santri Nasional", 4200000, "Renovasi Ringan Sanitas & Kamar Mandi Asrama Putra", 8500000},
		{11, 2025, "November 2025", 30, "Tryout Olimpiade & Pembinaan Intensif Kelas Akhir", 2500000, "Pengadaan Alat Kebersihan Besar-besaran Menghadapi Musim Penghujan", 2800000},
		{12, 2025, "Desember 2025", 31, "Ujian Akhir Semester (PAS) & Pembagian Raport", 5500000, "Perawatan Total Gedung Asrama Saat Libur Semester Ganjil", 15000000},
		{1, 2026, "Januari 2026", 31, "Orientasi Semester Genap & Rapat Kerja Evaluasi Guru", 2200000, "Pengadaan Kasur & Sprei Baru Santri Asrama Putri", 7500000},
		{2, 2026, "Februari 2026", 28, "Kegiatan Rihlah Ilmiah / Study Tour Santri & Guru", 6800000, "Upgrade Jaringan Wifi, CCTV, & Keamanan Asrama", 5200000},
		{3, 2026, "Maret 2026", 31, "Ujian Tengah Semester (PTS) Genap & Tarhib Ramadhan Sekolah", 3400000, "Persiapan Dapur Umum & Logistik Khusus Bulan Ramadhan", 8500000},
		{4, 2026, "April 2026", 30, "Daurah Intensif Al-Qur'an 10 Hari Ramadhan & Buka Bersama", 7500000, "Santunan Yatim & Bingkisan Lebaran untuk Musrif serta Staf Asrama", 12000000},
		{5, 2026, "Mei 2026", 31, "Pelaksanaan Ujian Sekolah & Ujian Akhir Tahun Ajaran", 5800000, "Pengecatan Ulang Gedung Asrama Jelang Tahun Baru", 14000000},
		{6, 2026, "Juni 2026", 30, "Acara Tasyakuran, Khotmul Qur'an & Wisuda Purna Studi Santri", 12500000, "Perbaikan Instalasi Listrik, Genset & Pompa Air Pondok", 9500000},
	}

	// Ambil akun-akun penting SMP
	akunBOS_SMP, _ := getAkunByKode(smp.ID, "4190")
	akunBebanKonsumsiSMP, _ := getAkunByKode(smp.ID, "5120")
	akunBebanKegiatanSMP, _ := getAkunByKode(smp.ID, "5130")
	akunBebanKesehatanSMP, _ := getAkunByKode(smp.ID, "5140")
	akunBebanUtilitasSMP, _ := getAkunByKode(smp.ID, "5190")
	akunBebanKebersihanSMP, _ := getAkunByKode(smp.ID, "5200")
	akunBebanOperasionalSMP, _ := getAkunByKode(smp.ID, "5210")

	// Ambil akun-akun penting Yayasan
	akunInfaqYys, _ := getAkunByKode(yayasan.ID, "4170")
	akunKoperasiYys, _ := getAkunByKode(yayasan.ID, "4180")
	akunBebanKonsumsiYys, _ := getAkunByKode(yayasan.ID, "5120")
	akunBebanPembangunanYys, _ := getAkunByKode(yayasan.ID, "5170")
	akunBebanUtilitasYys, _ := getAkunByKode(yayasan.ID, "5190")
	akunBebanKebersihanYys, _ := getAkunByKode(yayasan.ID, "5200")
	akunBebanOperasionalYys, _ := getAkunByKode(yayasan.ID, "5210")

	for i, bk := range bulanTahunList {
		bulanStr := fmt.Sprintf("%02d", bk.Bulan)
		tahunStr := fmt.Sprintf("%04d", bk.Tahun)
		tglPrefix := fmt.Sprintf("%s-%s", tahunStr, bulanStr)

		log.Printf("-> Mensimulasikan bulan %s (%s)...", bk.Nama, tglPrefix)

		// ---------------------------------------------------------------------
		// A. SIMULASI UNIT SEKOLAH (SMP ISLAM IMAM NAWAWI)
		// ---------------------------------------------------------------------

		// 1. Tarik Tunai Bank ke Kas Sekolah di awal bulan (Tgl 1)
		if akunKasSMP != nil && akunBankSMP != nil {
			_ = a.CreateTransaksi(TransaksiInput{
				EntitasID: smp.ID, Tanggal: fmt.Sprintf("%s-01", tglPrefix), NoReferensi: fmt.Sprintf("TRK-SMP-%s-01", bulanStr),
				Keterangan: fmt.Sprintf("Tarik Tunai Bank BSI untuk Kas Operasional Sekolah Bulan %s", bk.Nama),
				TotalDebit: 4000000, TotalKredit: 4000000,
				Details: []models.JurnalDetail{
					{AkunID: akunKasSMP.ID, Debit: 4000000, Kredit: 0},
					{AkunID: akunBankSMP.ID, RekeningBankID: &rekBSISMP.ID, Debit: 0, Kredit: 4000000},
				},
			})
		}

		// 2. Transaksi Khusus Awal Semester (Juli & Januari) - Penerimaan BOS & Tagihan Tahunan
		if bk.Bulan == 7 && bk.Tahun == 2025 && akunBankSMP != nil && akunBOS_SMP != nil {
			_ = a.CreateTransaksi(TransaksiInput{
				EntitasID: smp.ID, Tanggal: "2025-07-03", NoReferensi: "BOS-2025-I",
				Keterangan: "Pencairan Dana Bantuan Operasional Sekolah (BOS) Tahap I Tahun 2025/2026",
				TotalDebit: 35000000, TotalKredit: 35000000,
				Details: []models.JurnalDetail{
					{AkunID: akunBankSMP.ID, RekeningBankID: &rekBSISMP.ID, Debit: 35000000, Kredit: 0},
					{AkunID: akunBOS_SMP.ID, Debit: 0, Kredit: 35000000},
				},
			})
			// Generate Tagihan Awal Tahun SMP
			_ = a.GenerateTagihanBebasMassal(smp.ID, "", "Uang Pangkal & Kegiatan Sekolah", 3500000)
			_ = a.GenerateTagihanBebasMassal(smp.ID, "", "Seragam Sekolah Lengkap", 850000)
			_ = a.GenerateTagihanBebasMassal(smp.ID, "", "Buku Paket & Modul Diniyah", 500000)

			// Bayarkan tagihan tahunan untuk sebagian besar siswa (sisakan 1 menunggak)
			var tagihanTahunan []models.TagihanSiswa
			database.DB.Where("entitas_id = ? AND jenis_tagihan IN (?, ?, ?)", smp.ID, "Uang Pangkal & Kegiatan Sekolah", "Seragam Sekolah Lengkap", "Buku Paket & Modul Diniyah").Limit(27).Find(&tagihanTahunan)
			for _, t := range tagihanTahunan {
				_ = a.BayarTagihanSiswa(t.ID, t.TotalTagihan, akunBankSMP.ID, "2025-07-06")
			}
		} else if bk.Bulan == 1 && bk.Tahun == 2026 && akunBankSMP != nil && akunBOS_SMP != nil {
			_ = a.CreateTransaksi(TransaksiInput{
				EntitasID: smp.ID, Tanggal: "2026-01-04", NoReferensi: "BOS-2026-II",
				Keterangan: "Pencairan Dana Bantuan Operasional Sekolah (BOS) Tahap II Tahun 2025/2026",
				TotalDebit: 35000000, TotalKredit: 35000000,
				Details: []models.JurnalDetail{
					{AkunID: akunBankSMP.ID, RekeningBankID: &rekBSISMP.ID, Debit: 35000000, Kredit: 0},
					{AkunID: akunBOS_SMP.ID, Debit: 0, Kredit: 35000000},
				},
			})
			_ = a.GenerateTagihanBebasMassal(smp.ID, "", "Buku Modul & LKS Semester Genap", 400000)
			var tagihanGenap []models.TagihanSiswa
			database.DB.Where("entitas_id = ? AND jenis_tagihan = ?", smp.ID, "Buku Modul & LKS Semester Genap").Limit(8).Find(&tagihanGenap)
			for _, t := range tagihanGenap {
				_ = a.BayarTagihanSiswa(t.ID, t.TotalTagihan, akunBankSMP.ID, "2026-01-08")
			}
		}

		// 3. Generate Tagihan SPP Bulanan (Sekolah)
		namaTagihanSPP := fmt.Sprintf("SPP Sekolah %s", bk.Nama)
		_ = a.GenerateTagihanBebasMassal(smp.ID, "", namaTagihanSPP, 450000)

		// Simulasi Pembayaran SPP (Untuk bulan ke 1 s/d 10 dibayar lunas 100%, bulan Mei-Juni disisakan menunggak)
		var tagihanSPPBulanIni []models.TagihanSiswa
		limitBayar := 10
		if i == 10 { // Mei 2026
			limitBayar = 8
		} else if i == 11 { // Juni 2026 (bulan terakhir)
			limitBayar = 6 // Sisakan 4 siswa menunggak untuk testing fitur bayar
		}
		database.DB.Where("entitas_id = ? AND jenis_tagihan = ? AND status != 'Lunas'", smp.ID, namaTagihanSPP).Limit(limitBayar).Find(&tagihanSPPBulanIni)
		for _, t := range tagihanSPPBulanIni {
			_ = a.BayarTagihanSiswa(t.ID, t.TotalTagihan, akunBankSMP.ID, fmt.Sprintf("%s-05", tglPrefix))
		}

		// 4. Transaksi Operasional Rutin Sekolah
		trxRutinSMP := []struct {
			Tgl        int
			NoRef      string
			Keterangan string
			AkunBeban  *models.AkunKeuangan
			Nominal    int64
			PakaiBank  bool
		}{
			{2, fmt.Sprintf("ATK-SMP-%02d", bk.Bulan), fmt.Sprintf("Belanja ATK, Kertas HVS & Tinta Kantor TU Bulan %s", bk.Nama), akunBebanOperasionalSMP, 420000, false},
			{8, fmt.Sprintf("KONS-SMP-%02d", bk.Bulan), fmt.Sprintf("Konsumsi Rapat Dewan Guru & Pembinaan Bulan %s", bk.Nama), akunBebanKonsumsiSMP, 350000, false},
			{12, fmt.Sprintf("UKS-SMP-%02d", bk.Bulan), fmt.Sprintf("Pengadaan Obat-obatan UKS & Kesehatan Siswa Bulan %s", bk.Nama), akunBebanKesehatanSMP, 250000, false},
			{15, fmt.Sprintf("PLN-SMP-%02d", bk.Bulan), fmt.Sprintf("Pembayaran Tagihan Listrik, Air PDAM & Internet Wifi Sekolah Bulan %s", bk.Nama), akunBebanUtilitasSMP, 2350000, true},
			{18, fmt.Sprintf("FOT-SMP-%02d", bk.Bulan), fmt.Sprintf("Pembayaran Jasa Fotokopi Ujian, Tugas & Modul Sekolah %s", bk.Nama), akunBebanOperasionalSMP, 390000, false},
			{22, fmt.Sprintf("EVT-SMP-%02d", bk.Bulan), bk.EventSMP, akunBebanKegiatanSMP, bk.BiayaSMP, true},
			{28, fmt.Sprintf("TRN-SMP-%02d", bk.Bulan), fmt.Sprintf("Biaya Transportasi Dinas & Pengiriman Dokumen Sekolah Bulan %s", bk.Nama), akunBebanOperasionalSMP, 240000, false},
			{bk.HariAkhir, fmt.Sprintf("KBR-SMP-%02d", bk.Bulan), fmt.Sprintf("Pembersihan Sarana & Pemeliharaan Alat Kebersihan Sekolah Bulan %s", bk.Nama), akunBebanKebersihanSMP, 330000, false},
		}
		for _, tx := range trxRutinSMP {
			if tx.AkunBeban != nil && akunKasSMP != nil && akunBankSMP != nil {
				akunKreditID := akunKasSMP.ID
				var rekID *int = nil
				if tx.PakaiBank {
					akunKreditID = akunBankSMP.ID
					rekID = &rekBSISMP.ID
				}
				_ = a.CreateTransaksi(TransaksiInput{
					EntitasID: smp.ID, Tanggal: fmt.Sprintf("%s-%02d", tglPrefix, tx.Tgl), NoReferensi: tx.NoRef,
					Keterangan: tx.Keterangan, TotalDebit: tx.Nominal, TotalKredit: tx.Nominal,
					Details: []models.JurnalDetail{
						{AkunID: tx.AkunBeban.ID, Debit: tx.Nominal, Kredit: 0},
						{AkunID: akunKreditID, RekeningBankID: rekID, Debit: 0, Kredit: tx.Nominal},
					},
				})
			}
		}

		// 5. Proses Penggajian Guru & Staf SMP
		_ = a.ProsesPenggajianBulanan(ProsesGajiInput{
			EntitasID:      smp.ID,
			Bulan:          bk.Bulan,
			Tahun:          bk.Tahun,
			RekeningBankID: rekBSISMP.ID,
			Tanggal:        fmt.Sprintf("%s-25", tglPrefix),
		})

		// ---------------------------------------------------------------------
		// B. SIMULASI UNIT PUSAT & ASRAMA (YAYASAN PONDOK PESANTREN)
		// ---------------------------------------------------------------------

		// 1. Tarik Tunai Bank ke Kas Asrama di awal bulan (Tgl 1)
		if akunKasYys != nil && akunBankYys != nil {
			_ = a.CreateTransaksi(TransaksiInput{
				EntitasID: yayasan.ID, Tanggal: fmt.Sprintf("%s-01", tglPrefix), NoReferensi: fmt.Sprintf("TRK-YYS-%s-01", bulanStr),
				Keterangan: fmt.Sprintf("Tarik Tunai Bank Mandiri untuk Kas Operasional Asrama Bulan %s", bk.Nama),
				TotalDebit: 5500000, TotalKredit: 5500000,
				Details: []models.JurnalDetail{
					{AkunID: akunKasYys.ID, Debit: 5500000, Kredit: 0},
					{AkunID: akunBankYys.ID, RekeningBankID: &rekMandiriYys.ID, Debit: 0, Kredit: 5500000},
				},
			})
		}

		// 2. Transaksi Khusus Awal Tahun (Juli 2025) - Tagihan Pembangunan & Kitab
		if bk.Bulan == 7 && bk.Tahun == 2025 && akunBankYys != nil {
			_ = a.GenerateTagihanBebasMassal(yayasan.ID, "Asrama Putra", "Uang Pembangunan & Wakaf Sarana", 5000000)
			_ = a.GenerateTagihanBebasMassal(yayasan.ID, "Asrama Putri", "Uang Pembangunan & Wakaf Sarana", 5000000)
			_ = a.GenerateTagihanBebasMassal(yayasan.ID, "", "Kitab Turats & Perlengkapan Asrama", 1200000)

			var tagihanAsramaAwal []models.TagihanSiswa
			database.DB.Where("entitas_id = ? AND jenis_tagihan IN (?, ?)", yayasan.ID, "Uang Pembangunan & Wakaf Sarana", "Kitab Turats & Perlengkapan Asrama").Limit(14).Find(&tagihanAsramaAwal)
			for _, t := range tagihanAsramaAwal {
				_ = a.BayarTagihanSiswa(t.ID, t.TotalTagihan, akunBankYys.ID, "2025-07-07")
			}
		} else if bk.Bulan == 1 && bk.Tahun == 2026 && akunBankYys != nil {
			_ = a.GenerateTagihanBebasMassal(yayasan.ID, "", "Kegiatan & Ekstra Asrama Semester Genap", 600000)
			var tagihanKegiatan []models.TagihanSiswa
			database.DB.Where("entitas_id = ? AND jenis_tagihan = ?", yayasan.ID, "Kegiatan & Ekstra Asrama Semester Genap").Limit(7).Find(&tagihanKegiatan)
			for _, t := range tagihanKegiatan {
				_ = a.BayarTagihanSiswa(t.ID, t.TotalTagihan, akunBankYys.ID, "2026-01-07")
			}
		}

		// 3. Generate Tagihan Syahriyah Bulanan (Asrama & Makan)
		namaSyahriyah := fmt.Sprintf("Syahriyah Asrama & Makan %s", bk.Nama)
		_ = a.GenerateTagihanBebasMassal(yayasan.ID, "", namaSyahriyah, 850000)

		// Simulasi Pembayaran Syahriyah (Sisakan menunggak di Mei dan Juni 2026)
		limitBayarYys := 8
		if i == 10 { // Mei 2026
			limitBayarYys = 6
		} else if i == 11 { // Juni 2026
			limitBayarYys = 5 // Sisakan 3 santri menunggak
		}
		var syahriyahBulanIni []models.TagihanSiswa
		database.DB.Where("entitas_id = ? AND jenis_tagihan = ? AND status != 'Lunas'", yayasan.ID, namaSyahriyah).Limit(limitBayarYys).Find(&syahriyahBulanIni)
		for _, t := range syahriyahBulanIni {
			_ = a.BayarTagihanSiswa(t.ID, t.TotalTagihan, akunBankYys.ID, fmt.Sprintf("%s-06", tglPrefix))
		}

		// 4. Penerimaan Infaq Wakaf & Koperasi Pondok
		if akunBankYys != nil && akunInfaqYys != nil {
			_ = a.CreateTransaksi(TransaksiInput{
				EntitasID: yayasan.ID, Tanggal: fmt.Sprintf("%s-14", tglPrefix), NoReferensi: fmt.Sprintf("WKF-YYS-%02d", bk.Bulan),
				Keterangan: fmt.Sprintf("Penerimaan Sumbangan Infaq & Wakaf Pembangunan Asrama Bulan %s", bk.Nama),
				TotalDebit: 25000000, TotalKredit: 25000000,
				Details: []models.JurnalDetail{
					{AkunID: akunBankYys.ID, RekeningBankID: &rekBSIYys.ID, Debit: 25000000, Kredit: 0},
					{AkunID: akunInfaqYys.ID, Debit: 0, Kredit: 25000000},
				},
			})
		}
		if akunBankYys != nil && akunKoperasiYys != nil {
			_ = a.CreateTransaksi(TransaksiInput{
				EntitasID: yayasan.ID, Tanggal: fmt.Sprintf("%s-25", tglPrefix), NoReferensi: fmt.Sprintf("KOP-YYS-%02d", bk.Bulan),
				Keterangan: fmt.Sprintf("Setor Bagi Hasil Usaha Koperasi Pondok & Laundry Bulan %s", bk.Nama),
				TotalDebit: 15500000, TotalKredit: 15500000,
				Details: []models.JurnalDetail{
					{AkunID: akunBankYys.ID, RekeningBankID: &rekMandiriYys.ID, Debit: 15500000, Kredit: 0},
					{AkunID: akunKoperasiYys.ID, Debit: 0, Kredit: 15500000},
				},
			})
		}

		// 5. Transaksi Operasional Rutin & Dapur Asrama
		trxRutinYys := []struct {
			Tgl        int
			NoRef      string
			Keterangan string
			AkunBeban  *models.AkunKeuangan
			Nominal    int64
			PakaiBank  bool
		}{
			{2, fmt.Sprintf("DPR-YYS-%02d-1", bk.Bulan), fmt.Sprintf("Belanja Bahan Makan & Dapur Asrama Minggu I Bulan %s", bk.Nama), akunBebanKonsumsiYys, 4500000, true},
			{8, fmt.Sprintf("DPR-YYS-%02d-2", bk.Bulan), fmt.Sprintf("Belanja Bahan Makan & Dapur Asrama Minggu II Bulan %s", bk.Nama), akunBebanKonsumsiYys, 4500000, true},
			{15, fmt.Sprintf("PLN-YYS-%02d", bk.Bulan), fmt.Sprintf("Tagihan Listrik, Air PDAM & Internet Asrama serta Masjid Bulan %s", bk.Nama), akunBebanUtilitasYys, 7500000, true},
			{16, fmt.Sprintf("DPR-YYS-%02d-3", bk.Bulan), fmt.Sprintf("Belanja Bahan Makan & Dapur Asrama Minggu III Bulan %s", bk.Nama), akunBebanKonsumsiYys, 4500000, true},
			{20, fmt.Sprintf("EVT-YYS-%02d", bk.Bulan), bk.EventYys, akunBebanPembangunanYys, bk.BiayaYys, true},
			{23, fmt.Sprintf("DPR-YYS-%02d-4", bk.Bulan), fmt.Sprintf("Belanja Bahan Makan & Dapur Asrama Minggu IV Bulan %s", bk.Nama), akunBebanKonsumsiYys, 4500000, true},
			{28, fmt.Sprintf("KBR-YYS-%02d", bk.Bulan), fmt.Sprintf("Pembelian Alat Kebersihan, Obat Pembersih & Sanitasi Asrama %s", bk.Nama), akunBebanKebersihanYys, 1250000, false},
			{bk.HariAkhir, fmt.Sprintf("OPR-YYS-%02d", bk.Bulan), fmt.Sprintf("Biaya Operasional Kantor Yayasan, ATK & Administrasi Asrama %s", bk.Nama), akunBebanOperasionalYys, 850000, false},
		}
		for _, ty := range trxRutinYys {
			if ty.AkunBeban != nil && akunKasYys != nil && akunBankYys != nil {
				akunKreditID := akunKasYys.ID
				var rekID *int = nil
				if ty.PakaiBank {
					akunKreditID = akunBankYys.ID
					rekID = &rekMandiriYys.ID
				}
				_ = a.CreateTransaksi(TransaksiInput{
					EntitasID: yayasan.ID, Tanggal: fmt.Sprintf("%s-%02d", tglPrefix, ty.Tgl), NoReferensi: ty.NoRef,
					Keterangan: ty.Keterangan, TotalDebit: ty.Nominal, TotalKredit: ty.Nominal,
					Details: []models.JurnalDetail{
						{AkunID: ty.AkunBeban.ID, Debit: ty.Nominal, Kredit: 0},
						{AkunID: akunKreditID, RekeningBankID: rekID, Debit: 0, Kredit: ty.Nominal},
					},
				})
			}
		}

		// 6. Proses Penggajian Pimpinan & Pengasuh Asrama
		_ = a.ProsesPenggajianBulanan(ProsesGajiInput{
			EntitasID:      yayasan.ID,
			Bulan:          bk.Bulan,
			Tahun:          bk.Tahun,
			RekeningBankID: rekMandiriYys.ID,
			Tanggal:        fmt.Sprintf("%s-26", tglPrefix),
		})
	}

	log.Println("Berhasil generate data dummy lengkap Sekolah Boarding School selama 1 tahun (Juli 2025 - Juni 2026)!")
	return nil
}
