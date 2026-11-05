package main

import (
	"keuangan_nawawi/models"
	"testing"
)

func TestGenerateDummy(t *testing.T) {
	app := NewApp()
	app.startup(nil)

	err := app.GenerateDataDummy()
	if err != nil {
		t.Fatalf("Gagal generate data dummy: %v", err)
	}
	t.Log("Sukses generate data dummy lengkap!")
}

func TestGetLaporanKonsolidasi(t *testing.T) {
	app := NewApp()
	app.startup(nil)

	res, err := app.GetLaporanKonsolidasi("bulanan", 6, 2026, 0)
	if err != nil {
		t.Fatalf("Gagal get laporan: %v", err)
	}
	t.Logf("Entitas: %v", res.EntitasNames)
	for _, g := range res.ArusKas {
		t.Logf("Group: %s (Total Global: %d)", g.NamaGrup, g.Total.Global)
		for _, item := range g.Items {
			if item.Global != 0 {
				t.Logf("  - %s: %v (Red: %v)", item.Uraian, item.Values, item.IsRed)
			}
		}
	}
}

func TestLaporanJHKAndTransaksiSimpel(t *testing.T) {
	app := NewApp()
	app.startup(nil)

	// Pastikan akun ISAK 35 & JHK di-seed
	err := app.GenerateISAK35(1)
	if err != nil {
		t.Fatalf("Gagal GenerateISAK35: %v", err)
	}

	// Ambil akun 801 (Alat & Perlengkapan Belajar) atau 100 (SPP)
	akuns, err := app.GetAkunKeuangan(1)
	if err != nil {
		t.Fatalf("Gagal GetAkunKeuangan: %v", err)
	}
	var akunBebanID, akunSPPID int
	for _, a := range akuns {
		if a.KodeAkun == "801" {
			akunBebanID = a.ID
		}
		if a.KodeAkun == "100" {
			akunSPPID = a.ID
		}
	}

	if akunBebanID > 0 {
		// Test Transaksi Simpel Pengeluaran Kas Tunai
		err = app.CreateTransaksiSimpel(models.TransaksiSimpelInput{
			EntitasID:   1,
			Jenis:       "Pengeluaran",
			Tanggal:     "2026-07-15",
			NoReferensi: "BKK-TEST-01",
			Keterangan:  "Beli Spidol dan Penghapus Papan",
			AkunID:      akunBebanID,
			IsBank:      false,
			Nominal:     75000,
		})
		if err != nil {
			t.Fatalf("Gagal CreateTransaksiSimpel (Tunai): %v", err)
		}
		t.Log("Sukses catat transaksi pengeluaran tunai!")
	}

	if akunSPPID > 0 {
		// Test Transaksi Simpel Pemasukan Tunai
		err = app.CreateTransaksiSimpel(models.TransaksiSimpelInput{
			EntitasID:   1,
			Jenis:       "Pemasukan",
			Tanggal:     "2026-07-10",
			NoReferensi: "BKM-TEST-01",
			Keterangan:  "Penerimaan SPP Tunai",
			AkunID:      akunSPPID,
			IsBank:      false,
			Nominal:     500000,
		})
		if err != nil {
			t.Fatalf("Gagal CreateTransaksiSimpel (Pemasukan): %v", err)
		}
		t.Log("Sukses catat transaksi pemasukan tunai!")
	}

	// Test Laporan JHK
	jhk, err := app.GetLaporanJHK(1, 7, 2026)
	if err != nil {
		t.Fatalf("Gagal GetLaporanJHK: %v", err)
	}
	t.Logf("Laporan JHK Periode: %d/%d, Items: %d", jhk.Bulan, jhk.Tahun, len(jhk.Items))
	t.Logf("Total Debet: %v, Total Kredit: %v, Total Balance: %v, Total Saldo Akhir: %v",
		jhk.TotalDebet, jhk.TotalKredit, jhk.TotalBalance, jhk.TotalSaldoAkhir)
	for _, item := range jhk.Items {
		t.Logf("Pos %s (%s): Debet=%v, Kredit=%v, Balance=%v, SaldoAkhir=%v",
			item.Kode, item.NamaPos, item.Debet, item.Kredit, item.Balance, item.SaldoAkhir)
	}
}
