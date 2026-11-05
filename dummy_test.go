package main

import (
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
