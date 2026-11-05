<script>
  // Force HMR reload for Vite cache invalidation
  import { onMount } from 'svelte';
  import { GetAkunKeuangan, CreateAkunKeuangan, UpdateAkunKeuangan, DeleteAkunKeuangan, ExportAkunCSV, ImportAkunCSV, GenerateISAK35, GetEntitas, CreateEntitas, GetTransaksi, CreateTransaksi, UpdateTransaksi, DeleteTransaksi, GetJurnalDetail, GetPegawai, CreatePegawai, UpdatePegawai, DeletePegawai, GetRiwayatGajiBulanan, ProsesPenggajianBulanan, GetBulanGajiTerproses, GetRekeningBank, CreateRekeningBank, UpdateRekeningBank, DeleteRekeningBank, GetSiswa, CreateSiswa, UpdateSiswa, DeleteSiswa, GenerateTagihanSPP, GetTagihanSiswa, GetTunggakanSiswa, BayarTagihanSiswa, GenerateTagihanBebasMassal, BayarBanyakTagihan, GetLaporanKonsolidasi, GenerateDataDummy, BackupDatabase, RestoreDatabase, GetDatabaseInfo } from '../wailsjs/go/main/App.js';

  // State Manajemen Entitas
  let activeEntitas = null;
  
  // State untuk Laporan Konsolidasi
  let laporanData = null;
  let activeLaporanTab = 'neraca';
  let laporanTahun = 2026;
  let laporanMode = 'bulanan';
  let laporanPeriode = 6; // Juni (karena akhir simulasi data 1 tahun berada di Juni 2026)

  async function fetchLaporanKonsolidasi() {
    try {
        laporanData = await GetLaporanKonsolidasi(laporanMode, Number(laporanPeriode), Number(laporanTahun), activeEntitas ? activeEntitas.id : 0);
    } catch(e) {
        showToast("Gagal memuat laporan konsolidasi", "error");
    }
  }

  function onModeChange() {
    if (laporanMode === 'bulanan') {
      laporanPeriode = 6; // Juni 2026
    } else if (laporanMode === 'kuartal') {
      laporanPeriode = 2; // Kuartal II (Apr - Jun 2026)
    } else {
      laporanPeriode = 0;
    }
    fetchLaporanKonsolidasi();
  }
  let entitasList = [];
  let showCreateForm = false;
  let newNama = '';
  let newTipe = 'Sekolah';

  // --- STATE & FUNGSI BACKUP / RESTORE DATABASE (POIN 1) ---
  let dbInfo = { path: '', size: '-', lastModified: '-' };
  let isBackingUp = false;
  let isRestoring = false;

  async function loadDatabaseInfo() {
    try {
      dbInfo = await GetDatabaseInfo();
    } catch (e) {
      console.error("Gagal mengambil info database:", e);
    }
  }

  async function handleBackupDatabase() {
    try {
      isBackingUp = true;
      const savedPath = await BackupDatabase();
      if (savedPath) {
        showToast("Database berhasil dicadangkan ke: " + savedPath, "success");
        await loadDatabaseInfo();
      }
    } catch (err) {
      showToast("Gagal mencadangkan data: " + err, "error");
    } finally {
      isBackingUp = false;
    }
  }

  function handleRestoreDatabase() {
    confirmModal = {
      title: "Konfirmasi Pemulihan Database (Restore)",
      message: "PERINGATAN KRUSIAL: Memulihkan database akan menimpa seluruh data yang ada saat ini dengan data dari file cadangan. Pastikan Anda telah membuat cadangan data terkini sebelum melanjutkan. Lanjutkan pemulihan?",
      onConfirm: async () => {
        confirmModal = null;
        try {
          isRestoring = true;
          const restoredPath = await RestoreDatabase();
          if (restoredPath) {
            showToast("Database berhasil dipulihkan!", "success");
            await loadDatabaseInfo();
            await loadEntitas();
            activeEntitas = null;
            currentPage = 'dashboard';
          }
        } catch (err) {
          showToast("Gagal memulihkan database: " + err, "error");
        } finally {
          isRestoring = false;
        }
      },
      onCancel: () => confirmModal = null
    };
  }

  // State Aplikasi Utama
  let currentPage = 'dashboard';
  
  // Data Struktur Sidebar (Grup dan Item)
  let menuGroups = [
    {
      title: '',
      items: [
        { id: 'dashboard', label: 'Ringkasan / Dashboard' }
      ]
    },
    {
      title: 'Akuntansi',
      items: [
        { id: 'akun', label: 'Daftar Akun (COA)' },
        { id: 'rekening', label: 'Rekening Bank' },
        { id: 'jurnal', label: 'Jurnal Umum' }
      ]
    },
    {
      title: 'Operasional',
      items: [
        { id: 'siswa', label: 'Data Siswa' },
        { id: 'spp', label: 'Keuangan Siswa' },
        { id: 'pegawai', label: 'Data Pegawai' },
        { id: 'penggajian', label: 'Proses Penggajian' }
      ]
    },
    {
      title: 'Laporan ISAK 35',
      items: [
        { id: 'neraca', label: 'Posisi Keuangan' },
        { id: 'aktivitas', label: 'Penghasilan Komprehensif' },
        { id: 'aruskas', label: 'Laporan Arus Kas' }
      ]
    },
    {
      title: 'Sistem & Data',
      items: [
        { id: 'backup', label: 'Cadangkan & Pulihkan' }
      ]
    }
  ];

  // --- STATE AKUN ---
  let akunList = [];
  let editingAkun = null;
  let akunForm = { kode_akun: '', nama_akun: '', kategori: 'Aset', sifat: 'Debit', saldo: 0 };
  let fileInput;

  // --- STATE JURNAL UMUM ---
  let transaksiList = [];
  let searchJurnal = '';
  let editingJurnalId = null;
  let viewingJurnal = null;
  let viewingJurnalDetails = [];
  
  // --- STATE UMUM ---
  let todayDate = (() => {
    const d = new Date();
    return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
  })();

  // --- STATE REKENING BANK ---
  let rekeningList = [];
  let editingRekeningId = null;
  let rekeningForm = { nama_rekening: '', nomor_rekening: '', saldo: 0 };

  // --- STATE KEUANGAN SISWA (TAGIHAN) ---
  let selectedSiswaSPP = null;
  let tunggakanSPPList = [];
  let tunggakanSiswaGrouped = [];
  
  // State untuk Bayar Tagihan Baru
  let activeBayarSiswa = null;
  let selectedTagihanIDs = [];
  let bayarSPPForm = { akun_penerima_id: '', tanggal_bayar: todayDate };
  let showBuatTagihanModal = false;
  let generateSPPForm = { bulan: new Date().getMonth() + 1, tahun: new Date().getFullYear() };
  let buatTagihanForm = { target_kelas: '', items: [{ jenis_tagihan: '', nominal: 0 }] };
  let isGeneratingSPP = false;

  // --- STATE SISWA ---
  let siswaList = [];
  $: uniqueKelas = [...new Set(siswaList.map(s => s.kelas).filter(k => k))].sort();
  let editingSiswaId = null;
  let siswaFilter = { kelas: '', jenjang: '', status: '' };
  let searchSiswa = '';
  let siswaForm = { nis: '', nisn: '', nama_lengkap: '', jenis_kelamin: 'L', tempat_lahir: '', tanggal_lahir: '', agama: 'Islam', alamat: '', nama_ayah: '', nama_ibu: '', telepon_ortu: '', kelas: '', jenjang: 'SD', tarif_spp: 0, status: 'Aktif' };

  // --- STATE PENGGAJIAN ---
  let pegawaiList = [];
  let riwayatGaji = [];
  let pegawaiForm = { nip: '', nama: '', jabatan: '', gaji_pokok: 0, tanggal_mulai_kerja: todayDate, periode_kenaikan: 1, nominal_kenaikan: 0 };
  let showPegawaiForm = false;
  let editingPegawaiId = null;
  let prosesGajiForm = { bulan: new Date().getMonth() + 1, tahun: new Date().getFullYear(), rekening_bank_id: '' };
  let processedMonths = [];

  $: filteredTransaksi = transaksiList.filter(trx => 
    (trx.keterangan || '').toLowerCase().includes(searchJurnal.toLowerCase()) || 
    (trx.no_referensi || '').toLowerCase().includes(searchJurnal.toLowerCase())
  );
  
  let jurnalForm = {
    tanggal: todayDate,
    no_referensi: '',
    keterangan: '',
    details: [
      { akun_id: '', debit: 0, kredit: 0, rekening_bank_id: '' },
      { akun_id: '', debit: 0, kredit: 0, rekening_bank_id: '' }
    ]
  };

  $: totalDebit = jurnalForm.details.reduce((sum, row) => sum + (parseFloat(row.debit) || 0), 0);
  $: totalKredit = jurnalForm.details.reduce((sum, row) => sum + (parseFloat(row.kredit) || 0), 0);
  $: isBalanced = totalDebit === totalKredit && totalDebit > 0;

  onMount(async () => {
    loadEntitas();
  });

  async function loadEntitas() {
    try {
      entitasList = await GetEntitas();
    } catch (e) {
      console.error(e);
    }
  }

  // --- CUSTOM NOTIFICATIONS & MODALS ---
  let toastMsg = '';
  let toastType = 'success'; 
  let toastTimeout;

  function showToast(message, type = 'info') {
    toastMsg = message;
    toastType = type;
    clearTimeout(toastTimeout);
    toastTimeout = setTimeout(() => { toastMsg = ''; }, 3000);
  }

  let confirmModal = null; 

  function showConfirm(title, message, onConfirm) {
    confirmModal = { title, message, onConfirm, onCancel: () => confirmModal = null };
  }

  async function handleCreateEntitas() {
    if (!newNama) return showToast("Nama entitas tidak boleh kosong!", "error");
    try {
      const e = await CreateEntitas(newNama, newTipe);
      entitasList = [...entitasList, e];
      showCreateForm = false;
      newNama = '';
      showToast("Entitas berhasil dibuat", "success");
    } catch (err) {
      console.error(err);
      showToast("Gagal menyimpan entitas.", "error");
    }
  }

  async function handleGenerateDummy() {
    showConfirm("Generate Data Dummy", "Sistem akan menghapus data saat ini dan menggantinya dengan simulasi data dummy lengkap untuk semua modul (Sekolah & Yayasan). Lanjutkan?", async () => {
      try {
        await GenerateDataDummy();
        showToast("Data dummy lengkap berhasil digenerate!", "success");
        await loadEntitas();
        confirmModal = null;
      } catch (err) {
        showToast("Gagal generate data dummy: " + err, "error");
      }
    });
  }

  function selectEntitas(e) {
    activeEntitas = e;
    loadDataUtama(e.id);
  }

  async function loadDataUtama(entitasId) {
    await loadAkun(entitasId);
    await loadRekening(entitasId);
    await loadTransaksi(entitasId);
    await loadPegawai(entitasId);
    await loadSiswa(entitasId);
  }

  // --- LOGIKA SISWA ---
  async function loadSiswa(entitasId) {
    try {
      siswaList = await GetSiswa(entitasId);
    } catch(err) {
      console.error(err);
    }
  }

  $: filteredSiswa = siswaList.filter(s => {
    const matchSearch = !searchSiswa || s.nama_lengkap.toLowerCase().includes(searchSiswa.toLowerCase()) || s.nis.includes(searchSiswa);
    const matchStatus = !siswaFilter.status || s.status === siswaFilter.status;
    return matchSearch && matchStatus;
  });

  function openSiswaForm(s = null) {
    if (s) {
      editingSiswaId = s.id;
      siswaForm = { ...s };
    } else {
      editingSiswaId = null;
      siswaForm = { nis: '', nisn: '', nama_lengkap: '', jenis_kelamin: 'L', tempat_lahir: '', tanggal_lahir: '', agama: 'Islam', alamat: '', nama_ayah: '', nama_ibu: '', telepon_ortu: '', kelas: '', jenjang: 'SD', tarif_spp: 0, status: 'Aktif' };
    }
    currentPage = 'form_siswa';
  }

  async function saveSiswa() {
    if (!siswaForm.nama_lengkap || !siswaForm.nis) {
      showToast("Nama dan NIS wajib diisi!", "error");
      return;
    }
    try {
      if (editingSiswaId) {
        await UpdateSiswa({ id: editingSiswaId, entitas_id: activeEntitas.id, ...siswaForm });
        showToast("Data siswa berhasil diperbarui", "success");
      } else {
        await CreateSiswa({ entitas_id: activeEntitas.id, ...siswaForm });
        showToast("Siswa berhasil ditambahkan", "success");
      }
      currentPage = 'siswa';
      await loadSiswa(activeEntitas.id);
    } catch (err) {
      showToast(err, "error");
    }
  }

  function deleteSiswa(id) {
    showConfirm("Hapus Siswa", "Yakin ingin menghapus data siswa ini?", async () => {
      try {
        await DeleteSiswa(id);
        showToast("Data siswa berhasil dihapus", "success");
        await loadSiswa(activeEntitas.id);
      } catch (err) {
        showToast(err, "error");
      }
    });
  }

  // --- LOGIKA KEUANGAN SISWA (TAGIHAN) ---
  async function loadTunggakanSPP() {
    if (!activeEntitas || !activeEntitas.id) return;
    try {
      tunggakanSPPList = await GetTunggakanSiswa(activeEntitas.id) || [];
      const grouped = {};
      tunggakanSPPList.forEach(t => {
        if (!grouped[t.siswa_id]) {
          grouped[t.siswa_id] = {
            siswa_id: t.siswa_id,
            siswa: t.siswa,
            total_tunggakan: 0,
            tagihans: []
          };
        }
        grouped[t.siswa_id].total_tunggakan += t.sisa_tagihan;
        grouped[t.siswa_id].tagihans.push(t);
      });
      tunggakanSiswaGrouped = Object.values(grouped);
    } catch (err) {
      showToast("Gagal memuat tunggakan tagihan", "error");
    }
  }

  async function handleGenerateSPP() {
    if (!activeEntitas) return;
    isGeneratingSPP = true;
    try {
      await GenerateTagihanSPP(activeEntitas.id, Number(generateSPPForm.bulan), Number(generateSPPForm.tahun));
      showToast(`Tagihan SPP bulan ${generateSPPForm.bulan}/${generateSPPForm.tahun} berhasil di-generate!`, "success");
      await loadTunggakanSPP();
    } catch (err) {
      showToast(err, "error");
    }
    isGeneratingSPP = false;
  }

  async function openBuatTagihanModal() {
    if (activeEntitas) {
      await loadSiswa(activeEntitas.id);
    }
    showBuatTagihanModal = true;
  }

  async function handleBuatTagihanBebas() {
    let validItems = buatTagihanForm.items.filter(i => i.jenis_tagihan && i.nominal > 0);
    if (validItems.length === 0) {
      showToast("Lengkapi minimal satu Jenis Tagihan dan Nominal!", "error");
      return;
    }
    isGeneratingSPP = true;
    try {
      let successCount = 0;
      for (const item of validItems) {
        try {
          await GenerateTagihanBebasMassal(activeEntitas.id, buatTagihanForm.target_kelas, item.jenis_tagihan, Number(item.nominal));
          successCount++;
        } catch (itemErr) {
          showToast(`${item.jenis_tagihan}: ${itemErr}`, "error");
        }
      }
      if (successCount > 0) {
        let targetMsg = buatTagihanForm.target_kelas ? `untuk kelas ${buatTagihanForm.target_kelas}` : 'untuk semua siswa';
        showToast(`${successCount} tagihan massal berhasil dibuat ${targetMsg}!`, "success");
      }
      showBuatTagihanModal = false;
      buatTagihanForm = { target_kelas: '', items: [{ jenis_tagihan: '', nominal: 0 }] };
      await loadTunggakanSPP();
    } catch (err) {
      showToast(err, "error");
    }
    isGeneratingSPP = false;
  }
  
  function addTagihanItem() {
    buatTagihanForm.items = [...buatTagihanForm.items, { jenis_tagihan: '', nominal: 0 }];
  }
  
  function removeTagihanItem(idx) {
    buatTagihanForm.items = buatTagihanForm.items.filter((_, i) => i !== idx);
  }

  function gotoBayarTagihanSiswa(group) {
    activeBayarSiswa = group;
    selectedTagihanIDs = group.tagihans.map(t => t.id);
    bayarSPPForm = { akun_penerima_id: '', tanggal_bayar: todayDate };
    currentPage = 'bayar_tagihan';
  }

  function toggleTagihanSelection(id) {
    if (selectedTagihanIDs.includes(id)) {
      selectedTagihanIDs = selectedTagihanIDs.filter(tid => tid !== id);
    } else {
      selectedTagihanIDs = [...selectedTagihanIDs, id];
    }
  }

  function toggleSemuaTagihan() {
    const payableTags = activeBayarSiswa.tagihans.filter(t => t.status !== 'Lunas');
    if (selectedTagihanIDs.length === payableTags.length && payableTags.length > 0) {
      selectedTagihanIDs = [];
    } else {
      selectedTagihanIDs = payableTags.map(t => t.id);
    }
  }

  async function submitBayarMassal() {
    if (selectedTagihanIDs.length === 0) {
      showToast("Pilih minimal satu tagihan untuk dibayar", "error");
      return;
    }
    if (!bayarSPPForm.akun_penerima_id) {
      showToast("Pilih Rekening/Kas penerima", "error");
      return;
    }
    
    isGeneratingSPP = true;
    try {
      await BayarBanyakTagihan(selectedTagihanIDs, parseInt(bayarSPPForm.akun_penerima_id), bayarSPPForm.tanggal_bayar);
      showToast("Berhasil memproses pembayaran tagihan!", "success");
      
      currentPage = 'spp';
      activeBayarSiswa = null;
      await loadTunggakanSPP();
      await loadAkun(activeEntitas.id);
      await loadRekening(activeEntitas.id);
      await loadTransaksi(activeEntitas.id);
    } catch (err) {
      showToast("Gagal memproses pembayaran: " + err, "error");
    } finally {
      isGeneratingSPP = false;
    }
  }

  // --- LOGIKA REKENING BANK ---
  async function loadRekening(entitasId) {
    try {
      rekeningList = await GetRekeningBank(entitasId);
    } catch(err) {
      console.error(err);
    }
  }

  function openRekeningForm(rek = null) {
    if (rek) {
      editingRekeningId = rek.id;
      rekeningForm = { ...rek };
    } else {
      editingRekeningId = null;
      rekeningForm = { nama_rekening: '', nomor_rekening: '', saldo: 0 };
    }
    currentPage = 'form_rekening';
  }

  async function saveRekening() {
    if (!rekeningForm.nama_rekening) {
      showToast("Nama rekening harus diisi!", "error");
      return;
    }
    try {
      if (editingRekeningId) {
        await UpdateRekeningBank({ id: editingRekeningId, ...rekeningForm });
        showToast("Rekening berhasil diperbarui", "success");
      } else {
        await CreateRekeningBank({ entitas_id: activeEntitas.id, ...rekeningForm, saldo: Number(rekeningForm.saldo) });
        showToast("Rekening berhasil ditambahkan", "success");
      }
      currentPage = 'rekening';
      await loadRekening(activeEntitas.id);
      await loadAkun(activeEntitas.id); // sinkronkan saldo akun Bank
    } catch (err) {
      showToast(err, "error");
    }
  }

  function deleteRekening(id) {
    showConfirm("Hapus Rekening", "Yakin ingin menghapus rekening ini?", async () => {
      try {
        await DeleteRekeningBank(id);
        await loadRekening(activeEntitas.id);
        await loadAkun(activeEntitas.id); // sinkronkan saldo akun Bank
        showToast("Rekening berhasil dihapus", "success");
        confirmModal = null;
      } catch (err) {
        showToast(err, "error");
      }
    });
  }

  async function loadPegawai(entitasId) {
    try {
      pegawaiList = await GetPegawai(entitasId);
    } catch(err) {
      console.error(err);
    }
  }

  async function loadRiwayatGaji() {
    if(!activeEntitas) return;
    try {
      riwayatGaji = await GetRiwayatGajiBulanan(Number(prosesGajiForm.bulan), Number(prosesGajiForm.tahun), activeEntitas.id);
      await checkProcessedMonths();
    } catch(err) {
      console.error(err);
    }
  }

  async function checkProcessedMonths() {
    try {
      if (prosesGajiForm.tahun) {
        processedMonths = await GetBulanGajiTerproses(Number(prosesGajiForm.tahun), activeEntitas.id);
      } else {
        processedMonths = [];
      }
    } catch (e) {
      console.error(e);
    }
  }

  async function loadTransaksi(entitasID) {
    try {
      transaksiList = await GetTransaksi(entitasID);
    } catch (e) {
      console.error(e);
    }
  }

  async function loadAkun(entitasID) {
    try {
      akunList = await GetAkunKeuangan(entitasID);
    } catch (e) {
      console.error(e);
    }
  }

  function goBackToSelection() {
    activeEntitas = null;
    akunList = [];
    currentPage = 'dashboard';
  }

  // --- AKUN KEUANGAN (BUKU BESAR) FUNCTIONS ---
  function openAkunForm(akun = null) {
    if (akun) {
      editingAkun = akun;
      akunForm = { ...akun };
    } else {
      editingAkun = null;
      akunForm = { kode_akun: '', nama_akun: '', kategori: 'Aset', sifat: 'Debit', saldo: 0 };
    }
    currentPage = 'form_akun';
  }

  async function saveAkun() {
    try {
      akunForm.saldo = parseFloat(akunForm.saldo) || 0;
      akunForm.entitas_id = activeEntitas.id;
      
      if (editingAkun) {
        await UpdateAkunKeuangan(akunForm);
      } else {
        await CreateAkunKeuangan(akunForm);
      }
      currentPage = 'akun';
      showToast("Akun berhasil disimpan", "success");
      await loadAkun(activeEntitas.id);
    } catch (err) {
      showToast("Gagal menyimpan akun.", "error");
    }
  }

  function deleteAkun(id) {
    showConfirm("Hapus Akun", "Yakin ingin menghapus akun ini? Data jurnal terkait mungkin terpengaruh.", async () => {
      try {
        await DeleteAkunKeuangan(id);
        await loadAkun(activeEntitas.id);
        showToast("Akun berhasil dihapus", "success");
        confirmModal = null;
      } catch (err) {
        showToast(err, "error");
      }
    });
  }

  function handleGenerateISAK() {
    showConfirm("Template ISAK 35", "Sistem akan memasukkan 15+ akun standar secara otomatis. Lanjutkan?", async () => {
      await GenerateISAK35(activeEntitas.id);
      await loadAkun(activeEntitas.id);
      showToast("Template ISAK 35 berhasil dipasang", "success");
      confirmModal = null;
    });
  }

  async function handleExport() {
    try {
      const csvStr = await ExportAkunCSV(activeEntitas.id);
      if (!csvStr) return showToast("Tidak ada data untuk diekspor", "error");
      
      const blob = new Blob([csvStr], { type: 'text/csv' });
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.setAttribute('hidden', '');
      a.setAttribute('href', url);
      a.setAttribute('download', `akun_${activeEntitas.nama}.csv`);
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      showToast("Berhasil ekspor ke CSV", "success");
    } catch (err) {
      showToast("Gagal ekspor CSV", "error");
    }
  }

  function triggerImport() {
    fileInput.click();
  }

  async function handleImport(event) {
    const file = event.target.files[0];
    if (!file) return;

    // Batasi ukuran file 1MB untuk mencegah serangan upload file raksasa (OOM/Crash)
    if (file.size > 1024 * 1024) {
      showToast("File terlalu besar! Maksimal 1MB.", "error");
      fileInput.value = '';
      return;
    }
    
    const reader = new FileReader();
    reader.onload = async (e) => {
      const csvStr = e.target.result;
      try {
        await ImportAkunCSV(activeEntitas.id, csvStr);
        showToast("Import CSV berhasil!", "success");
        await loadAkun(activeEntitas.id);
      } catch (err) {
        showToast("Gagal import CSV", "error");
      }
    };
    reader.readAsText(file);
    fileInput.value = ''; // Reset input
  }

  function hitungGajiAktual(p) {
    let gaji = p.gaji_pokok;
    if (p.tanggal_mulai_kerja && p.periode_kenaikan > 0 && p.nominal_kenaikan > 0) {
      let mulai = new Date(p.tanggal_mulai_kerja);
      let sekarang = new Date();
      let tahunKerja = sekarang.getFullYear() - mulai.getFullYear();
      if (sekarang.getMonth() < mulai.getMonth() || (sekarang.getMonth() === mulai.getMonth() && sekarang.getDate() < mulai.getDate())) {
        tahunKerja--;
      }
      if (tahunKerja > 0) {
        let kenaikan = Math.floor(tahunKerja / p.periode_kenaikan);
        gaji += (kenaikan * p.nominal_kenaikan);
      }
    }
    return gaji;
  }

  // --- HELPER FORMAT TANGGAL ---
  function formatTanggalID(tgl) {
    if (!tgl || tgl.startsWith('0001')) return '-';
    const date = new Date(tgl);
    return date.toLocaleDateString('id-ID', {
      day: 'numeric',
      month: 'long',
      year: 'numeric'
    });
  }

  // --- FUNGSI JURNAL UMUM ---
  // Cek apakah akun yang dipilih adalah akun Bank (bawaan)
  function isBankAkun(akunId) {
    if (!akunId) return false;
    const akun = akunList.find(a => a.id === Number(akunId) || a.id === akunId);
    return akun && akun.is_bawaan && akun.nama_akun === 'Bank';
  }

  function addJurnalRow() {
    jurnalForm.details = [...jurnalForm.details, { akun_id: '', debit: 0, kredit: 0, rekening_bank_id: '' }];
  }

  function removeJurnalRow(index) {
    if (jurnalForm.details.length <= 2) return showToast("Minimal harus ada 2 baris jurnal", "error");
    jurnalForm.details = jurnalForm.details.filter((_, i) => i !== index);
  }

  async function saveJurnal() {
    if (!isBalanced) return showToast("Debit dan Kredit harus seimbang (Balance)", "error");
    if (!jurnalForm.tanggal || !jurnalForm.keterangan) return showToast("Tanggal dan Keterangan wajib diisi", "error");
    
    for (let row of jurnalForm.details) {
      if (!row.akun_id) return showToast("Pilih akun untuk setiap baris jurnal", "error");
      if (row.debit === 0 && row.kredit === 0) return showToast("Nominal Debit/Kredit harus diisi", "error");
      // Validasi: jika akun Bank, wajib pilih rekening
      if (isBankAkun(row.akun_id) && !row.rekening_bank_id) {
        return showToast("Pilih rekening bank yang terlibat (BSI/Mandiri/dll)", "error");
      }
    }

    try {
      let payload = {
        entitas_id: activeEntitas.id,
        tanggal: jurnalForm.tanggal,
        no_referensi: jurnalForm.no_referensi,
        keterangan: jurnalForm.keterangan,
        total_debit: totalDebit,
        total_kredit: totalKredit,
        details: jurnalForm.details.map(d => ({
          akun_id: parseInt(d.akun_id),
          debit: parseFloat(d.debit) || 0,
          kredit: parseFloat(d.kredit) || 0,
          rekening_bank_id: d.rekening_bank_id ? parseInt(d.rekening_bank_id) : null
        }))
      };
      if (editingJurnalId) {
        await UpdateTransaksi(editingJurnalId, payload);
        showToast("Jurnal berhasil diperbarui!", "success");
      } else {
        await CreateTransaksi(payload);
        showToast("Jurnal berhasil disimpan!", "success");
      }
      
      currentPage = 'jurnal';
      editingJurnalId = null;
      await loadTransaksi(activeEntitas.id);
      await loadAkun(activeEntitas.id);
      await loadRekening(activeEntitas.id); // sinkronkan saldo rekening bank
    } catch (err) {
      console.error(err);
      showToast("Gagal menyimpan jurnal", "error");
    }
  }

  async function editJurnal(trx) {
    try {
      const details = await GetJurnalDetail(trx.id);
      jurnalForm = {
        tanggal: trx.tanggal ? trx.tanggal.split('T')[0] : todayDate,
        no_referensi: trx.no_referensi || '',
        keterangan: trx.keterangan,
        details: details.map(d => ({
          akun_id: d.akun_id,
          debit: d.debit,
          kredit: d.kredit,
          rekening_bank_id: d.rekening_bank_id || ''
        }))
      };
      // Ensure at least 2 rows
      while (jurnalForm.details.length < 2) {
        jurnalForm.details.push({ akun_id: '', debit: 0, kredit: 0 });
      }
      editingJurnalId = trx.id;
      currentPage = 'form_jurnal';
    } catch (err) {
      showToast("Gagal memuat data untuk diedit", "error");
    }
  }

  async function hapusJurnal(trx) {
    showConfirm("Hapus Transaksi", `Yakin ingin menghapus jurnal "${trx.keterangan}"? Saldo akun akan dikembalikan seperti semula.`, async () => {
      try {
        await DeleteTransaksi(trx.id);
        showToast("Transaksi berhasil dihapus & Saldo di-rollback!", "success");
        await loadTransaksi(activeEntitas.id);
        await loadAkun(activeEntitas.id);
        confirmModal = null;
      } catch (err) {
        showToast("Gagal menghapus transaksi", "error");
      }
    });
  }

  async function lihatDetailJurnal(trx) {
    try {
      const details = await GetJurnalDetail(trx.id);
      viewingJurnalDetails = details || [];
      viewingJurnal = trx;
      currentPage = 'view_jurnal';
    } catch (err) {
      console.error(err);
      showToast("Gagal memuat detail jurnal: " + err, "error");
    }
  }

  function openPegawaiForm() {
    pegawaiForm = { nip: '', nama: '', jabatan: '', gaji_pokok: 0, tanggal_mulai_kerja: todayDate, periode_kenaikan: 1, nominal_kenaikan: 0, tunjangan: [], potongan: [] };
    editingPegawaiId = null;
    currentPage = 'form_pegawai';
  }

  function editPegawai(p) {
    editingPegawaiId = p.id;
    let tJson = [];
    let pJson = [];
    try { if (p.tunjangan) tJson = JSON.parse(p.tunjangan); } catch (e) {}
    try { if (p.potongan) pJson = JSON.parse(p.potongan); } catch (e) {}
    
    pegawaiForm = { 
      nip: p.nip, nama: p.nama, jabatan: p.jabatan, gaji_pokok: p.gaji_pokok, 
      tanggal_mulai_kerja: p.tanggal_mulai_kerja ? p.tanggal_mulai_kerja.split('T')[0] : '',
      periode_kenaikan: p.periode_kenaikan,
      nominal_kenaikan: p.nominal_kenaikan,
      tunjangan: tJson,
      potongan: pJson
    };
    currentPage = 'form_pegawai';
  }

  function addKomponenGaji(type) {
    if (type === 'tunjangan') pegawaiForm.tunjangan = [...pegawaiForm.tunjangan, { nama: '', nominal: 0 }];
    if (type === 'potongan') pegawaiForm.potongan = [...pegawaiForm.potongan, { nama: '', nominal: 0 }];
  }

  function removeKomponenGaji(type, idx) {
    if (type === 'tunjangan') pegawaiForm.tunjangan = pegawaiForm.tunjangan.filter((_, i) => i !== idx);
    if (type === 'potongan') pegawaiForm.potongan = pegawaiForm.potongan.filter((_, i) => i !== idx);
  }

  async function savePegawai() {
    if (!pegawaiForm.nama || pegawaiForm.nama.trim() === '') {
      showToast("Nama pegawai tidak boleh kosong!", "error");
      return;
    }
    
    try {
      const payload = { 
        ...pegawaiForm, 
        entitas_id: activeEntitas.id,
        gaji_pokok: Number(pegawaiForm.gaji_pokok),
        periode_kenaikan: Number(pegawaiForm.periode_kenaikan),
        nominal_kenaikan: Number(pegawaiForm.nominal_kenaikan),
        tunjangan: JSON.stringify(pegawaiForm.tunjangan.map(t => ({ nama: t.nama, nominal: Number(t.nominal) }))),
        potongan: JSON.stringify(pegawaiForm.potongan.map(p => ({ nama: p.nama, nominal: Number(p.nominal) })))
      };

      if (editingPegawaiId) {
        await UpdatePegawai(editingPegawaiId, payload);
        showToast("Pegawai berhasil diperbarui", "success");
      } else {
        await CreatePegawai(payload);
        showToast("Pegawai berhasil ditambah", "success");
      }
      
      editingPegawaiId = null;
      pegawaiForm = { nip: '', nama: '', jabatan: '', gaji_pokok: 0, tanggal_mulai_kerja: todayDate, periode_kenaikan: 1, nominal_kenaikan: 0, tunjangan: [], potongan: [] };
      await loadPegawai(activeEntitas.id);
    } catch (err) {
      showToast(typeof err === 'string' ? err : "Gagal menyimpan pegawai: " + err, "error");
    }
  }

  async function hapusPegawai(p) {
    showConfirm("Hapus Pegawai", "Yakin ingin menghapus data pegawai ini?", async () => {
      await DeletePegawai(p.id);
      await loadPegawai(activeEntitas.id);
      showToast("Pegawai berhasil dihapus", "success");
      confirmModal = null;
    });
  }

  async function prosesGaji() {
    if (!prosesGajiForm.rekening_bank_id) {
      showToast("Silakan pilih Rekening Sumber Dana terlebih dahulu!", "error");
      return;
    }
    
    try {
      const payload = {
        bulan: Number(prosesGajiForm.bulan),
        tahun: Number(prosesGajiForm.tahun),
        rekening_bank_id: Number(prosesGajiForm.rekening_bank_id),
        entitas_id: activeEntitas.id
      };
      
      await ProsesPenggajianBulanan(payload);
      showToast("Proses penggajian berhasil!", "success");
      await loadRiwayatGaji();
      await checkProcessedMonths();
      await loadAkun(activeEntitas.id);
    } catch (err) {
      showToast("Gagal memproses gaji: " + err, "error");
    }
  }
</script>

{#if !activeEntitas}
  <!-- LAYAR PEMILIHAN ENTITAS (SETUP SCREEN) - SPLIT-SCREEN DENGAN ILUSTRASI -->
  <div class="setup-layout">
    <div class="setup-left">
      <div class="brand-huge">nawawi</div>
      <h2>Sistem Akuntansi Terpadu</h2>
      <p>Kelola pencatatan keuangan untuk seluruh unit pendidikan dan yayasan Anda secara profesional.</p>
    </div>

    <div class="setup-right">
      <div class="setup-content">
        {#if showCreateForm}
          <div class="kledo-form-container">
            <h2>Buat Unit/Perusahaan Baru</h2>
            <p>Masukkan data dasar untuk memisahkan pencatatan keuangan.</p>
            
            <input type="text" bind:value={newNama} placeholder="Nama Entitas (Misal: SD Nawawi)" class="kledo-input" />
            
            <select bind:value={newTipe} class="kledo-input">
              <option value="Sekolah">Sekolah / Unit Pendidikan</option>
              <option value="Perusahaan">Yayasan / Perusahaan Pusat</option>
            </select>
            
            <div class="form-actions-kledo">
              <button class="kledo-btn-primary" on:click={handleCreateEntitas}>Simpan & Buat</button>
              <button class="kledo-btn-outline" on:click={() => showCreateForm = false}>Batal</button>
            </div>
          </div>
        {:else}
          <h2 class="section-title">Pilih Unit / Entitas</h2>
          <div class="kledo-grid">
            {#each entitasList as e}
              <button class="kledo-card" on:click={() => selectEntitas(e)}>
                <div class="card-illustration">
                  {#if e.tipe === 'Sekolah'}
                    <div class="css-icon-wrapper green-theme">
                      <div class="css-sekolah">
                        <div class="roof"></div>
                        <div class="building">
                          <div class="windows">
                            <div class="win"></div><div class="win"></div>
                          </div>
                          <div class="door"></div>
                        </div>
                      </div>
                    </div>
                  {:else}
                    <div class="css-icon-wrapper blue-theme">
                      <div class="css-perusahaan">
                        <div class="tower">
                          <div class="win-row"><div class="win"></div><div class="win"></div></div>
                          <div class="win-row"><div class="win"></div><div class="win"></div></div>
                          <div class="win-row"><div class="win"></div><div class="win"></div></div>
                          <div class="door"></div>
                        </div>
                      </div>
                    </div>
                  {/if}
                </div>
                <h3>{e.nama}</h3>
                <p>{e.tipe === 'Sekolah' ? 'Unit Pendidikan' : 'Pusat Yayasan/Perusahaan'}</p>
              </button>
            {/each}
            
            <button class="kledo-card new-card" on:click={() => showCreateForm = true}>
              <div class="card-illustration">
                <div class="css-icon-wrapper dashed-theme">
                  <div class="css-plus">
                    <div class="v-line"></div>
                    <div class="h-line"></div>
                  </div>
                </div>
              </div>
              <h3>Tambah Unit Baru</h3>
              <p>Buat entitas akuntansi baru</p>
            </button>

            <button class="kledo-card new-card" style="border-color: #f59e0b; color: #f59e0b;" on:click={handleGenerateDummy}>
              <div class="card-illustration">
                <div class="css-icon-wrapper dashed-theme" style="border-color: #f59e0b; color: #f59e0b;">
                  <div style="font-size: 28px;">⚡</div>
                </div>
              </div>
              <h3 style="color: #f59e0b;">Generate Data Dummy</h3>
              <p>Isi data simulasi lengkap</p>
            </button>

            <!-- TOOLBAR CADANGKAN & PULIHKAN (LAYAR AWAL) -->
            <div style="grid-column: 1 / -1; display: flex; justify-content: space-between; align-items: center; margin-top: 1.5rem; padding: 1rem 1.25rem; background: #f8fafc; border: 1px dashed #cbd5e1; border-radius: 10px;">
              <div style="text-align: left;">
                <span style="font-weight: 700; color: #334155; font-size: 0.95rem;">💾 Manajemen Cadangan Data</span>
                <p style="margin: 2px 0 0 0; font-size: 0.8rem; color: #64748b;">Cadangkan data ke Flashdisk atau pulihkan database dari komputer lama</p>
              </div>
              <div style="display: flex; gap: 0.75rem;">
                <button class="kledo-btn-outline" style="padding: 0.4rem 0.9rem; font-size: 0.85rem;" on:click={handleBackupDatabase} disabled={isBackingUp}>
                  {isBackingUp ? 'Menyimpan...' : '📥 Cadangkan (.db)'}
                </button>
                <button class="kledo-btn-outline" style="padding: 0.4rem 0.9rem; font-size: 0.85rem; border-color: #f59e0b; color: #b45309;" on:click={handleRestoreDatabase} disabled={isRestoring}>
                  {isRestoring ? 'Memulihkan...' : '🔄 Pulihkan (.db)'}
                </button>
              </div>
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>

{:else}
  <!-- LAYAR APLIKASI UTAMA (DASHBOARD) -->
  <div class="layout">
    <!-- Sidebar Green Premium -->
    <aside class="sidebar-green">
      <div class="sidebar-header">
        <div class="brand-text">nawawi</div>
      </div>
      
      <!-- Pemilih Unit (Accounting Selector box) -->
      <button class="unit-selector" on:click={goBackToSelection} title="Ganti Unit/Perusahaan">
        <span class="unit-name">{activeEntitas.nama}</span>
        <svg class="chevron-right" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="9 18 15 12 9 6"></polyline></svg>
      </button>

      <!-- Navigasi Menu Bergrup -->
      <nav class="sidebar-nav">
        {#each menuGroups as group}
          <div class="nav-group">
            {#if group.title}
              <div class="nav-group-title">{group.title}</div>
            {/if}
            {#each group.items as item}
              <button 
                class="nav-item {currentPage === item.id ? 'active' : ''}" 
                on:click={() => {
                  currentPage = item.id;
                  if (currentPage === 'pegawai' || currentPage === 'penggajian') {
                    loadPegawai(activeEntitas.id);
                    if (currentPage === 'penggajian') {
                      loadRiwayatGaji();
                      checkProcessedMonths();
                    }
                  } else if (currentPage === 'spp') {
                    loadTunggakanSPP();
                    loadRekening(activeEntitas.id);
                    loadSiswa(activeEntitas.id);
                  } else if (['laporan', 'neraca', 'aktivitas', 'aruskas'].includes(currentPage)) {
                    activeLaporanTab = currentPage === 'laporan' ? 'neraca' : currentPage;
                    fetchLaporanKonsolidasi();
                  } else if (currentPage === 'backup') {
                    loadDatabaseInfo();
                  }
                }}>
                <span class="nav-icon">
                  <!-- Ikon SVG Manual: Lingkaran Kosong -->
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                    <circle cx="12" cy="12" r="8"></circle>
                  </svg>
                </span>
                <span class="nav-label">{item.label}</span>
              </button>
            {/each}
          </div>
        {/each}
      </nav>
    </aside>

    <!-- Main Content -->
    <main class="main-content">
      {#if currentPage === 'dashboard'}
        <div class="page-header">
          <h1>Dashboard {activeEntitas.nama}</h1>
          <p>Ringkasan posisi keuangan untuk unit ini.</p>
        </div>
        <div class="card-grid">
          <div class="card summary-card">
            <h3>Total Aset</h3>
            <div class="amount">Rp 0</div>
          </div>
          <div class="card summary-card">
            <h3>Total Liabilitas</h3>
            <div class="amount">Rp 0</div>
          </div>
          <div class="card summary-card">
            <h3>Surplus (Defisit)</h3>
            <div class="amount">Rp 0</div>
          </div>
        </div>

      {:else if currentPage === 'akun'}
        <div class="page-header">
          <div>
            <h1>Bagan Akun (Chart of Accounts)</h1>
            <p>Kelola daftar akun untuk {activeEntitas.nama}.</p>
          </div>
          <div class="header-actions">
            <button class="btn btn-outline" on:click={handleGenerateISAK}>
              <div class="css-icon-star"></div> Template ISAK
            </button>
            <button class="btn btn-outline" on:click={handleExport}>
              <div class="css-icon-arrow-up"></div> Export
            </button>
            <input type="file" bind:this={fileInput} on:change={handleImport} accept=".csv" style="display: none;" />
            <button class="btn btn-outline" on:click={triggerImport}>
              <div class="css-icon-arrow-down"></div> Import
            </button>
            <button class="btn btn-primary" on:click={() => openAkunForm()}>
              <div class="css-icon-plus-small"></div> Tambah Akun
            </button>
          </div>
        </div>
        <div class="card">
          <table class="data-table">
            <thead>
              <tr>
                <th>Kode</th>
                <th>Nama Akun</th>
                <th>Kategori</th>
                <th>Saldo Normal</th>
                <th>Saldo Akhir</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {#if akunList.length === 0}
                <tr>
                  <td colspan="6" class="empty-state">
                    Belum ada akun keuangan terdaftar.<br><br>
                    <button class="btn btn-outline" on:click={handleGenerateISAK}>Buat Otomatis (Standar ISAK 35)</button>
                  </td>
                </tr>
              {:else}
                {#each akunList as akun}
                  <tr>
                    <td>{akun.kode_akun}</td>
                    <td>{akun.nama_akun}</td>
                    <td>{akun.kategori}</td>
                    <td>{akun.sifat}</td>
                    <td class="text-right">Rp {akun.saldo.toLocaleString('id-ID')}</td>
                    <td>
                      <button class="btn-icon" on:click={() => openAkunForm(akun)} title="Edit">
                        <div class="css-icon-pencil"></div>
                      </button>
                      {#if !akun.is_bawaan}
                        <button class="btn-icon" on:click={() => deleteAkun(akun.id)} title="Hapus">
                          <div class="css-icon-trash"></div>
                        </button>
                      {:else}
                        <span style="color: #94a3b8; font-size: 0.8rem; margin-left: 0.5rem;" title="Akun sistem tidak bisa dihapus">🔒 Sistem</span>
                      {/if}
                    </td>
                  </tr>
                {/each}
              {/if}
            </tbody>
          </table>
        </div>

        <!-- MODAL FORM AKUN -->
      {:else if currentPage === 'form_akun'}
        <div class="page-header" style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
          <div>
            <h1 style="margin:0;">{editingAkun ? 'Edit Akun' : 'Tambah Akun Baru'}</h1>
            <p style="margin: 0.2rem 0 0; color:#64748b; font-size:0.9rem;">{editingAkun ? 'Ubah informasi akun keuangan' : 'Buat akun keuangan baru untuk entitas ini'}</p>
          </div>
          <button class="kledo-btn-outline" on:click={() => currentPage = 'akun'}>&larr; Kembali</button>
        </div>

        <div class="card" style="max-width: 600px; margin: 0 auto; padding: 2rem;">
          <div class="form-group">
            <label>Kode Akun</label>
            <input type="text" bind:value={akunForm.kode_akun} class="kledo-input" placeholder="Misal: 1110" />
          </div>
          
          <div class="form-group">
            <label>Nama Akun</label>
            <input type="text" bind:value={akunForm.nama_akun} class="kledo-input" placeholder="Misal: Kas Tunai" />
          </div>
          
          <div class="form-row">
            <div class="form-group">
              <label>Kategori</label>
              <select bind:value={akunForm.kategori} class="kledo-input">
                <option value="Aset">Aset</option>
                <option value="Liabilitas">Liabilitas</option>
                <option value="Aset Neto">Aset Neto (Ekuitas)</option>
                <option value="Pendapatan">Pendapatan</option>
                <option value="Beban">Beban</option>
              </select>
            </div>
            
            <div class="form-group">
              <label>Saldo Normal</label>
              <select bind:value={akunForm.sifat} class="kledo-input">
                <option value="Debit">Debit</option>
                <option value="Kredit">Kredit</option>
              </select>
            </div>
          </div>
          
          {#if !editingAkun?.is_bawaan}
          <div class="form-group">
            <label>Saldo Awal (Rp)</label>
            <input type="number" bind:value={akunForm.saldo} class="kledo-input" />
          </div>
          {:else}
          <div class="form-group">
            <label>Saldo (Rp)</label>
            <input type="number" value={akunForm.saldo} class="kledo-input" disabled style="background:#f8fafc; color:#64748b;" />
            <small style="color:#64748b; margin-top:0.3rem; display:block;">🔒 Saldo dikelola otomatis oleh sistem.</small>
          </div>
          {/if}
          
          <div class="form-actions-kledo" style="margin-top: 2rem; display: flex; gap: 1rem; justify-content: flex-start;">
            <button class="kledo-btn-primary" on:click={saveAkun}>Simpan Akun</button>
            <button class="kledo-btn-outline" on:click={() => currentPage = 'akun'}>Batal</button>
          </div>
        </div>

      {:else if currentPage === 'rekening'}
        <div class="page-header" style="display: flex; justify-content: space-between; align-items: flex-end;">
          <div>
            <h1>Rekening Bank</h1>
            <p>Kelola rekening bank dan buku pembantu kas untuk {activeEntitas.nama}.</p>
          </div>
          <button class="kledo-btn-primary" on:click={() => openRekeningForm()}>+ Tambah Rekening</button>
        </div>

        <div class="card" style="padding: 0;">
          <table class="data-table">
            <thead>
              <tr>
                <th>Nama Rekening</th>
                <th>Nomor Rekening</th>
                <th class="text-right">Saldo Saat Ini</th>
                <th style="width: 100px;">Aksi</th>
              </tr>
            </thead>
            <tbody>
              {#if rekeningList.length === 0}
                <tr><td colspan="4" class="text-center" style="padding: 2rem; color: #64748b;">Belum ada rekening.</td></tr>
              {:else}
                {#each rekeningList as rek}
                  <tr>
                    <td><strong>{rek.nama_rekening}</strong></td>
                    <td>{rek.nomor_rekening || '-'}</td>
                    <td class="text-right">Rp {rek.saldo.toLocaleString('id-ID')}</td>
                    <td>
                      <button class="btn-icon" on:click={() => openRekeningForm(rek)} title="Edit">
                        <div class="css-icon-pencil"></div>
                      </button>
                      <button class="btn-icon action-delete" on:click={() => deleteRekening(rek.id)} title="Hapus">
                        <div class="css-icon-trash"></div>
                      </button>
                    </td>
                  </tr>
                {/each}
              {/if}
            </tbody>
          </table>
        </div>

      {:else if currentPage === 'form_rekening'}
        <div class="page-header" style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
          <div>
            <h1 style="margin:0;">{editingRekeningId ? 'Edit Rekening' : 'Tambah Rekening'}</h1>
            <p style="margin: 0.2rem 0 0; color:#64748b; font-size:0.9rem;">{editingRekeningId ? 'Perbarui informasi rekening' : 'Tambahkan rekening kas atau bank baru'}</p>
          </div>
          <button class="kledo-btn-outline" on:click={() => currentPage = 'rekening'}>&larr; Kembali</button>
        </div>

        <div class="card" style="max-width: 600px; margin: 0 auto; padding: 2rem;">
          <div class="form-group">
            <label>Nama Rekening (Kas/Bank) *</label>
            <input type="text" class="kledo-input" bind:value={rekeningForm.nama_rekening} placeholder="Contoh: Bank BSI, Kas Tunai" />
          </div>
          <div class="form-group">
            <label>Nomor Rekening (Opsional)</label>
            <input type="text" class="kledo-input" bind:value={rekeningForm.nomor_rekening} placeholder="Contoh: 1234567890" />
          </div>
          <div class="form-group">
            <label>Saldo Awal (Rp)</label>
            <input type="number" class="kledo-input" bind:value={rekeningForm.saldo} disabled={editingRekeningId !== null} />
            {#if editingRekeningId}
              <small style="color: #64748b; margin-top: 0.2rem; display: block;">Saldo tidak bisa diubah langsung setelah pembuatan.</small>
            {/if}
          </div>
          
          <div class="form-actions-kledo" style="margin-top: 2rem; display: flex; gap: 1rem; justify-content: flex-start;">
            <button class="kledo-btn-primary" on:click={saveRekening}>Simpan Rekening</button>
            <button class="kledo-btn-outline" on:click={() => currentPage = 'rekening'}>Batal</button>
          </div>
        </div>

      {:else if currentPage === 'jurnal'}
        <div class="page-header" style="display: flex; justify-content: space-between; align-items: flex-end;">
          <div>
            <h1>Jurnal Umum</h1>
            <p>Catat transaksi harian untuk {activeEntitas.nama}.</p>
          </div>
          <div style="display: flex; gap: 1rem; align-items: center;">
            <input type="text" class="kledo-input" placeholder="Cari keterangan/referensi..." bind:value={searchJurnal} style="margin: 0; min-width: 250px;" />
            <button class="btn btn-primary" on:click={() => {
              editingJurnalId = null;
              jurnalForm = {
                tanggal: todayDate,
                no_referensi: '',
                keterangan: '',
                details: [ { akun_id: '', debit: 0, kredit: 0 }, { akun_id: '', debit: 0, kredit: 0 } ]
              };
              currentPage = 'form_jurnal';
            }}>
              <div class="css-icon-plus-small"></div> Catat Transaksi
            </button>
          </div>
        </div>
        <div class="card">
          <table class="data-table">
            <thead>
              <tr>
                <th>Tanggal</th>
                <th>No Referensi</th>
                <th>Keterangan</th>
                <th class="text-right">Debit</th>
                <th class="text-right">Kredit</th>
                <th>Aksi</th>
              </tr>
            </thead>
            <tbody>
              {#if filteredTransaksi.length === 0}
                <tr>
                  <td colspan="6" class="empty-state">
                    {searchJurnal ? 'Tidak ada transaksi yang cocok dengan pencarian.' : 'Belum ada transaksi jurnal terdaftar.'}
                  </td>
                </tr>
              {:else}
                {#each filteredTransaksi as trx}
                  <tr>
                    <td>{formatTanggalID(trx.tanggal)}</td>
                    <td>{trx.no_referensi || '-'}</td>
                    <td>{trx.keterangan}</td>
                    <td class="text-right">Rp {trx.total_debit.toLocaleString('id-ID')}</td>
                    <td class="text-right">Rp {trx.total_kredit.toLocaleString('id-ID')}</td>
                    <td>
                      <button class="btn-icon action-view" on:click={() => lihatDetailJurnal(trx)} title="Lihat Detail">
                        <div class="css-icon-eye"></div>
                      </button>
                      {#if !trx.no_referensi || !trx.no_referensi.startsWith('GAJI-')}
                        <button class="btn-icon action-edit" on:click={() => editJurnal(trx)} title="Edit">
                          <div class="css-icon-pencil"></div>
                        </button>
                      {/if}
                      <button class="btn-icon action-delete" on:click={() => hapusJurnal(trx)} title="Hapus">
                        <div class="css-icon-trash"></div>
                      </button>
                    </td>
                  </tr>
                {/each}
              {/if}
            </tbody>
          </table>
        </div>

        {:else if currentPage === 'view_jurnal' && viewingJurnal}
          <div class="page-header" style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <h1 style="margin-top: 0;">Detail Transaksi</h1>
              <p style="margin-bottom: 0;">Melihat rincian transaksi jurnal.</p>
            </div>
            <button class="kledo-btn-outline" on:click={() => { viewingJurnal = null; currentPage = 'jurnal'; }}>&larr; Kembali</button>
          </div>

          <div class="card" style="padding: 2rem;">
            <div style="margin-bottom: 2rem; color: #64748b; font-size: 1.05rem;">
              <p style="margin-bottom: 0.5rem;"><strong>Tanggal:</strong> <span style="color:#1e293b;">{formatTanggalID(viewingJurnal.tanggal)}</span></p>
              <p style="margin-bottom: 0.5rem;"><strong>Referensi:</strong> <span style="color:#1e293b;">{viewingJurnal.no_referensi || '-'}</span></p>
              <p style="margin-bottom: 0;"><strong>Keterangan:</strong> <span style="color:#1e293b;">{viewingJurnal.keterangan}</span></p>
            </div>

            <table class="jurnal-table" style="margin-bottom: 2rem;">
              <thead>
                <tr>
                  <th>Kode</th>
                  <th>Nama Akun</th>
                  <th class="text-right">Debit (Rp)</th>
                  <th class="text-right">Kredit (Rp)</th>
                </tr>
              </thead>
              <tbody>
                {#each viewingJurnalDetails as det}
                  <tr>
                    <td>{det.akun ? det.akun.kode_akun : '-'}</td>
                    <td>{det.akun ? det.akun.nama_akun : 'Akun Terhapus'}</td>
                    <td class="text-right">{det.debit > 0 ? det.debit.toLocaleString('id-ID') : '-'}</td>
                    <td class="text-right">{det.kredit > 0 ? det.kredit.toLocaleString('id-ID') : '-'}</td>
                  </tr>
                {/each}
              </tbody>
            </table>

            <div class="jurnal-footer balanced">
              <div class="jurnal-total">
                <span>Total Debit:</span>
                <strong>Rp {viewingJurnal.total_debit.toLocaleString('id-ID')}</strong>
              </div>
              <div class="jurnal-total">
                <span>Total Kredit:</span>
                <strong>Rp {viewingJurnal.total_kredit.toLocaleString('id-ID')}</strong>
              </div>
            </div>
          </div>

        {:else if currentPage === 'form_jurnal'}
          <div class="page-header" style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <h1 style="margin-top: 0;">{editingJurnalId ? 'Edit Transaksi Jurnal' : 'Catat Transaksi Jurnal'}</h1>
              <p style="margin-bottom: 0;">Masukkan rincian jurnal Anda. Pastikan total debit dan kredit seimbang.</p>
            </div>
            <button class="kledo-btn-outline" on:click={() => { currentPage = 'jurnal'; editingJurnalId = null; }}>
              &larr; Kembali
            </button>
          </div>

          <div class="kledo-card">
            <div class="form-row">
              <div class="form-group">
                <label>Tanggal Transaksi</label>
                <input type="date" class="kledo-input" bind:value={jurnalForm.tanggal} />
              </div>
              <div class="form-group">
                <label>No. Referensi (Opsional)</label>
                <input type="text" class="kledo-input" bind:value={jurnalForm.no_referensi} placeholder="Misal: INV-001" />
              </div>
            </div>
            <div class="form-group">
              <label>Keterangan / Deskripsi Singkat</label>
              <input type="text" class="kledo-input" bind:value={jurnalForm.keterangan} placeholder="Misal: Pembayaran SPP Bulan Juli" />
            </div>

            <table class="jurnal-table">
              <thead>
                <tr>
                  <th style="width: 35%;">Akun Keuangan</th>
                  <th style="width: 20%; font-size: 0.8rem; color: #64748b;">Rekening (jika Bank)</th>
                  <th class="text-right">Debit (Rp)</th>
                  <th class="text-right">Kredit (Rp)</th>
                  <th style="width: 40px;"></th>
                </tr>
              </thead>
              <tbody>
                {#each jurnalForm.details as row, i}
                  <tr>
                    <td>
                      <select bind:value={row.akun_id} class="jurnal-input" on:change={() => { if (!isBankAkun(row.akun_id)) row.rekening_bank_id = ''; }}>
                        <option value="">-- Pilih Akun --</option>
                        {#each akunList as ak}
                          <option value={ak.id}>{ak.kode_akun} - {ak.nama_akun}</option>
                        {/each}
                      </select>
                    </td>
                    <td>
                      {#if isBankAkun(row.akun_id)}
                        <select bind:value={row.rekening_bank_id} class="jurnal-input" style="background: #f0fdf4; border-color: #22c55e;">
                          <option value="">-- Pilih Rekening --</option>
                          {#each rekeningList as rek}
                            <option value={rek.id}>{rek.nama_rekening}</option>
                          {/each}
                        </select>
                      {:else}
                        <span style="color: #cbd5e1; font-size: 0.8rem; padding-left: 0.5rem;">-</span>
                      {/if}
                    </td>
                    <td>
                      <input type="number" bind:value={row.debit} class="jurnal-input text-right" min="0" on:input={() => { if(row.debit > 0) row.kredit = 0; }} />
                    </td>
                    <td>
                      <input type="number" bind:value={row.kredit} class="jurnal-input text-right" min="0" on:input={() => { if(row.kredit > 0) row.debit = 0; }} />
                    </td>
                    <td class="text-center">
                      <button class="btn-icon" on:click={() => removeJurnalRow(i)} disabled={jurnalForm.details.length <= 2} title="Hapus Baris">
                        <div class="css-icon-trash"></div>
                      </button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>

            <button class="kledo-btn-outline" style="margin-bottom: 2rem;" on:click={addJurnalRow}>+ Tambah Baris Akun</button>

            <div class="jurnal-footer {isBalanced ? 'balanced' : 'unbalanced'}">
              <div class="jurnal-total">
                <span>Total Debit:</span>
                <strong>Rp {totalDebit.toLocaleString('id-ID')}</strong>
              </div>
              <div class="jurnal-total">
                <span>Total Kredit:</span>
                <strong>Rp {totalKredit.toLocaleString('id-ID')}</strong>
              </div>
              {#if !isBalanced}
                <div style="width: 100%; text-align: right; color: #ef4444; font-size: 0.85rem; margin-top: 0.5rem;">
                  Debit dan Kredit harus seimbang! (Selisih: Rp {Math.abs(totalDebit - totalKredit).toLocaleString('id-ID')})
                </div>
              {/if}
            </div>
            
            <div class="form-actions-kledo">
              <button class="kledo-btn-primary" on:click={saveJurnal} disabled={!isBalanced}>{editingJurnalId ? 'Update Jurnal' : 'Simpan Jurnal'}</button>
              <button class="kledo-btn-outline" on:click={() => { currentPage = 'jurnal'; editingJurnalId = null; }}>Batal</button>
            </div>
          </div>

        {:else if currentPage === 'siswa'}
          <div class="page-header" style="display: flex; justify-content: space-between; align-items: flex-end;">
            <div>
              <h1>Data Siswa</h1>
              <p>Kelola data siswa {activeEntitas.nama} — Total: {siswaList.length} siswa</p>
            </div>
            <button class="kledo-btn-primary" on:click={() => openSiswaForm()}>+ Tambah Siswa</button>
          </div>

          <!-- Filter & Search -->
          <div class="card" style="padding: 1rem 1.5rem; margin-bottom: 1rem; display: flex; gap: 1rem; flex-wrap: wrap; align-items: center;">
            <input type="text" placeholder="Cari nama / NIS..." bind:value={searchSiswa} class="kledo-input" style="flex: 1; min-width: 200px;" />

            <select bind:value={siswaFilter.status} class="kledo-input" style="width: auto; min-width: 150px;">
              <option value="">Semua Status</option>
              <option value="Aktif">Aktif</option>
              <option value="Keluar">Keluar</option>
            </select>
            <span style="color: #64748b; font-size: 0.85rem;">Menampilkan {filteredSiswa.length} siswa</span>
          </div>

          <div class="card" style="padding: 0;">
            <table class="data-table">
              <thead>
                <tr>
                  <th style="width: 40px;">No</th>
                  <th>NIS</th>
                  <th>Nama Lengkap</th>
                  <th>Kelas</th>
                  <th>Jenis Kelamin</th>
                  <th>Status</th>
                  <th style="width: 80px;">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {#if filteredSiswa.length === 0}
                  <tr><td colspan="8" class="text-center" style="padding: 2rem; color: #64748b;">Belum ada data siswa.</td></tr>
                {:else}
                  {#each filteredSiswa as s, idx}
                    <tr>
                      <td class="text-center">{idx + 1}</td>
                      <td><code style="font-size:0.85rem;">{s.nis}</code></td>
                      <td><strong>{s.nama_lengkap}</strong></td>
                      <td>{s.kelas || '-'}</td>
                      <td>{s.jenis_kelamin === 'L' ? 'Laki-laki' : 'Perempuan'}</td>
                      <td>
                        <span style="padding: 0.2rem 0.6rem; border-radius: 99px; font-size: 0.75rem; font-weight: 600; background: {s.status === 'Aktif' ? '#dcfce7' : '#fee2e2'}; color: {s.status === 'Aktif' ? '#16a34a' : '#dc2626'};">
                          {s.status}
                        </span>
                      </td>
                      <td>
                        <button class="btn-icon" on:click={() => openSiswaForm(s)} title="Edit"><div class="css-icon-pencil"></div></button>
                        <button class="btn-icon action-delete" on:click={() => deleteSiswa(s.id)} title="Hapus"><div class="css-icon-trash"></div></button>
                      </td>
                    </tr>
                  {/each}
                {/if}
              </tbody>
            </table>
          </div>

        {:else if currentPage === 'form_siswa'}
          <div class="page-header" style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
            <div>
              <h1 style="margin:0;">{editingSiswaId ? 'Edit Data Siswa' : 'Tambah Siswa Baru'}</h1>
              <p style="margin: 0.2rem 0 0; color:#64748b; font-size:0.9rem;">Kelola data induk siswa</p>
            </div>
            <button class="kledo-btn-outline" on:click={() => currentPage = 'siswa'}>&larr; Kembali</button>
          </div>

          <div class="card" style="max-width: 800px; margin: 0 auto; padding: 2rem;">
            <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem 1.5rem;">
              <div class="form-group" style="margin: 0;">
                <label>NIS *</label>
                <input type="text" class="kledo-input" bind:value={siswaForm.nis} placeholder="Nomor Induk Siswa" />
              </div>
              <div class="form-group" style="margin: 0;">
                <label>NISN</label>
                <input type="text" class="kledo-input" bind:value={siswaForm.nisn} placeholder="Nomor Induk Siswa Nasional" />
              </div>
              <div class="form-group" style="margin: 0; grid-column: span 2;">
                <label>Nama Lengkap *</label>
                <input type="text" class="kledo-input" bind:value={siswaForm.nama_lengkap} placeholder="Nama lengkap siswa" />
              </div>
              <div class="form-group" style="margin: 0;">
                <label>Jenis Kelamin</label>
                <select class="kledo-input" bind:value={siswaForm.jenis_kelamin}>
                  <option value="L">Laki-laki</option>
                  <option value="P">Perempuan</option>
                </select>
              </div>
              <div class="form-group" style="margin: 0;">
                <label>Agama</label>
                <select class="kledo-input" bind:value={siswaForm.agama}>
                  <option>Islam</option>
                  <option>Kristen</option>
                  <option>Katolik</option>
                  <option>Hindu</option>
                  <option>Buddha</option>
                  <option>Konghucu</option>
                </select>
              </div>

              <div class="form-group" style="margin: 0;">
                <label>Kelas</label>
                <input type="text" class="kledo-input" bind:value={siswaForm.kelas} placeholder="Contoh: 7A, 1B" />
              </div>
              <div class="form-group" style="margin: 0;">
                <label>Tempat Lahir</label>
                <input type="text" class="kledo-input" bind:value={siswaForm.tempat_lahir} />
              </div>
              <div class="form-group" style="margin: 0;">
                <label>Tanggal Lahir</label>
                <input type="date" class="kledo-input" bind:value={siswaForm.tanggal_lahir} />
              </div>

              <div class="form-group" style="margin: 0;">
                <label>Alamat</label>
                <input type="text" class="kledo-input" bind:value={siswaForm.alamat} />
              </div>
              <div class="form-group" style="margin: 0;">
                <label>Nama Ayah</label>
                <input type="text" class="kledo-input" bind:value={siswaForm.nama_ayah} />
              </div>
              <div class="form-group" style="margin: 0;">
                <label>Nama Ibu</label>
                <input type="text" class="kledo-input" bind:value={siswaForm.nama_ibu} />
              </div>
              <div class="form-group" style="margin: 0;">
                <label>Telepon Orang Tua</label>
                <input type="text" class="kledo-input" bind:value={siswaForm.telepon_ortu} />
              </div>
              <div class="form-group" style="margin: 0;">
                <label>Status</label>
                <select class="kledo-input" bind:value={siswaForm.status}>
                  <option value="Aktif">Aktif</option>
                  <option value="Keluar">Keluar</option>
                </select>
              </div>
            </div>

            <div class="form-actions-kledo" style="margin-top: 2rem; display: flex; gap: 1rem; justify-content: flex-start;">
              <button class="kledo-btn-primary" on:click={saveSiswa}>Simpan Data Siswa</button>
              <button class="kledo-btn-outline" on:click={() => currentPage = 'siswa'}>Batal</button>
            </div>
          </div>

        {:else if currentPage === 'spp'}
          <div class="page-header" style="display: flex; justify-content: space-between; align-items: flex-end;">
            <div>
              <h1>Keuangan & Tagihan Siswa</h1>
              <p>Kelola tagihan (SPP, Uang Pangkal, Seragam) dan terima pembayaran secara fleksibel (Lunas/Cicil).</p>
            </div>
            
            <div style="display: flex; gap: 0.5rem; align-items: center;">
              <button class="kledo-btn-primary" style="height: 35px; padding: 0 1.2rem;" on:click={openBuatTagihanModal}>
                + Buat Tagihan
              </button>
            </div>
          </div>

          <div class="card">
            <h3 style="margin-top:0; margin-bottom: 1rem; color: #dc2626;">Daftar Tunggakan (Belum Lunas)</h3>
            <table class="data-table">
              <thead>
                <tr>
                  <th>NIS</th>
                  <th>Nama Siswa</th>
                  <th class="text-right">Total Tagihan Belum Lunas</th>
                  <th class="text-center">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {#if tunggakanSiswaGrouped.length === 0}
                  <tr>
                    <td colspan="4" class="text-center" style="color: #64748b; padding: 2rem 0;">Tidak ada siswa yang menunggak.</td>
                  </tr>
                {:else}
                  {#each tunggakanSiswaGrouped as group}
                    <tr>
                      <td>{group.siswa?.nis || '-'}</td>
                      <td>
                        <strong>{group.siswa?.nama_lengkap || '-'}</strong><br/>
                        <small style="color: #64748b;">Kelas {group.siswa?.kelas || '-'} - {group.siswa?.jenjang || '-'}</small>
                      </td>
                      <td class="text-right" style="font-weight: 500; color: #dc2626;">Rp {group.total_tunggakan.toLocaleString('id-ID')}</td>
                      <td class="text-center">
                        <button class="kledo-btn-primary" style="padding: 0.25rem 0.75rem; font-size: 0.85rem;" on:click={() => gotoBayarTagihanSiswa(group)}>
                          Lihat & Bayar
                        </button>
                      </td>
                    </tr>
                  {/each}
                {/if}
              </tbody>
            </table>
          </div>

          <!-- MODAL BUAT TAGIHAN BEBAS -->
          {#if showBuatTagihanModal}
            <div class="modal-overlay" on:click|self={() => showBuatTagihanModal = false}>
              <div class="modal-content" style="max-width: 500px;">
                <h2 style="margin-top:0;">Buat Tagihan</h2>
                
                {#each buatTagihanForm.items as item, i}
                  <div style="display: flex; gap: 1rem; align-items: flex-end; margin-bottom: 1rem;">
                    <div class="form-group" style="flex: 2; margin: 0;">
                      <label>Jenis Tagihan *</label>
                      <input type="text" list="jenis-tagihan-list" class="kledo-input" bind:value={item.jenis_tagihan} placeholder="Pilih dari daftar ISAK..." />
                      <datalist id="jenis-tagihan-list">
                        {#each akunList.filter(a => a.kategori === 'Pendapatan' && a.nama_akun.startsWith('Pendapatan ')) as akun}
                          <option value={akun.nama_akun.replace('Pendapatan ', '')} />
                        {/each}
                      </datalist>
                    </div>
                    <div class="form-group" style="flex: 1.5; margin: 0;">
                      <label>Nominal (Rp) *</label>
                      <input type="number" class="kledo-input" bind:value={item.nominal} />
                    </div>
                    {#if buatTagihanForm.items.length > 1}
                      <button class="kledo-btn-danger" style="margin-bottom: 0.2rem;" on:click={() => removeTagihanItem(i)}>Hapus</button>
                    {/if}
                  </div>
                {/each}
                <button class="kledo-btn-outline" style="margin-bottom: 1.5rem;" on:click={addTagihanItem}>+ Tambah Tagihan</button>
                
                <div class="form-group">
                  <label>Target Kelas (Opsional)</label>
                  <select class="kledo-input" bind:value={buatTagihanForm.target_kelas}>
                    <option value="">-- Semua Kelas (Semua Siswa) --</option>
                    {#each uniqueKelas as kelas}
                      <option value={kelas}>Kelas {kelas}</option>
                    {/each}
                  </select>
                  <small style="display:block; margin-top:0.3rem; color:#64748b;">Hanya siswa Aktif yang akan ditagih.</small>
                </div>

                <div class="form-actions-kledo" style="margin-top: 2rem;">
                  <button class="kledo-btn-primary" on:click={handleBuatTagihanBebas} disabled={isGeneratingSPP}>
                    {isGeneratingSPP ? 'Menyimpan...' : 'Simpan Tagihan'}
                  </button>
                  <button class="kledo-btn-outline" on:click={() => showBuatTagihanModal = false}>Batal</button>
                </div>
              </div>
            </div>
          {/if}

        {:else if currentPage === 'bayar_tagihan' && activeBayarSiswa}
          <div class="page-header" style="display: flex; justify-content: space-between; align-items: center;">
            <div>
              <h1 style="margin-top: 0;">Pembayaran Tagihan</h1>
              <p style="margin-bottom: 0;">Pilih tagihan yang akan dilunasi untuk siswa <strong>{activeBayarSiswa.siswa?.nama_lengkap}</strong>.</p>
            </div>
            <button class="kledo-btn-outline" on:click={() => currentPage = 'spp'}>&larr; Kembali</button>
          </div>

          <div class="card" style="padding: 2rem;">
            <div style="margin-bottom: 1.5rem; display: flex; justify-content: space-between; align-items: center;">
              <h3 style="margin: 0;">Rincian Tagihan Belum Lunas</h3>
              <button class="kledo-btn-outline" style="padding: 0.2rem 0.5rem; font-size: 0.85rem;" on:click={toggleSemuaTagihan}>
                {activeBayarSiswa.tagihans.filter(t => t.status !== 'Lunas').length > 0 && selectedTagihanIDs.length === activeBayarSiswa.tagihans.filter(t => t.status !== 'Lunas').length ? 'Batal Pilih Semua' : 'Pilih Semua'}
              </button>
            </div>

            <table class="data-table" style="margin-bottom: 2rem;">
              <thead>
                <tr>
                  <th style="width: 50px;" class="text-center">Pilih</th>
                  <th>Jenis Tagihan</th>
                  <th class="text-right">Nominal (Rp)</th>
                </tr>
              </thead>
              <tbody>
                {#each activeBayarSiswa.tagihans as t}
                  <tr style={t.status === 'Lunas' ? 'background-color: #f1f5f9; color: #94a3b8;' : (selectedTagihanIDs.includes(t.id) ? 'background-color: #f0fdf4;' : '')}>
                    <td class="text-center">
                      {#if t.status === 'Lunas'}
                        <span style="font-size: 0.75rem; background: #10b981; color: white; padding: 2px 6px; border-radius: 4px;">LUNAS</span>
                      {:else}
                        <input type="checkbox" checked={selectedTagihanIDs.includes(t.id)} on:change={() => toggleTagihanSelection(t.id)} style="width:16px; height:16px; cursor:pointer;" />
                      {/if}
                    </td>
                    <td style={t.status === 'Lunas' ? 'text-decoration: line-through;' : ''}>
                      {t.jenis_tagihan}
                      {#if t.bulan && t.tahun}
                        <small style="display: block; opacity: 0.8;">({t.bulan}/{t.tahun})</small>
                      {/if}
                    </td>
                    <td class="text-right" style="font-weight: 500; color: {t.status === 'Lunas' ? '#94a3b8' : '#dc2626'};">
                      Rp {t.sisa_tagihan.toLocaleString('id-ID')}
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>

            <div style="background: #f8fafc; padding: 1.5rem; border-radius: 6px; margin-bottom: 2rem; border: 1px solid #e2e8f0;">
              <h4 style="margin-top: 0; margin-bottom: 1rem;">Total yang akan dibayar: <span style="color: #dc2626; font-size: 1.25rem;">Rp {activeBayarSiswa.tagihans.filter(t => selectedTagihanIDs.includes(t.id)).reduce((acc, curr) => acc + curr.sisa_tagihan, 0).toLocaleString('id-ID')}</span></h4>
              
              <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1rem;">
                <div class="form-group" style="margin: 0;">
                  <label>Tanggal Pembayaran</label>
                  <input type="date" class="kledo-input" bind:value={bayarSPPForm.tanggal_bayar} />
                </div>
                <div class="form-group" style="margin: 0;">
                  <label>Terima Ke Rekening/Kas *</label>
                  <select class="kledo-input" bind:value={bayarSPPForm.akun_penerima_id}>
                    <option value="" disabled selected>-- Pilih Kas/Bank Tujuan --</option>
                    {#each rekeningList as rek}
                      <option value={rek.id}>{rek.nama_rekening}</option>
                    {/each}
                  </select>
                </div>
              </div>
            </div>

            <div class="form-actions-kledo">
              <button class="kledo-btn-primary" on:click={submitBayarMassal} disabled={isGeneratingSPP || selectedTagihanIDs.length === 0}>
                {isGeneratingSPP ? 'Memproses...' : 'Proses Pembayaran'}
              </button>
            </div>
          </div>

        {:else if currentPage === 'pegawai'}
          <div class="page-header" style="display: flex; justify-content: space-between; align-items: flex-end;">
            <div>
              <h1>Data Pegawai / Guru</h1>
              <p>Kelola data pegawai untuk {activeEntitas.nama}.</p>
            </div>
            <button class="kledo-btn-primary" on:click={openPegawaiForm}>
              <div class="css-plus-white" style="margin-right: 0.5rem;">
                <div class="v-line" style="background:white;"></div>
                <div class="h-line" style="background:white;"></div>
              </div> 
              Tambah Pegawai
            </button>
          </div>
          <div class="card">
            <table class="data-table">
              <thead>
                <tr>
                  <th>NIP/NIK</th>
                  <th>Nama Pegawai</th>
                  <th>Jabatan</th>
                  <th class="text-right">Gaji Pokok Awal</th>
                  <th class="text-right">Gaji Aktual (Saat Ini)</th>
                  <th class="text-center">Aksi</th>
                </tr>
              </thead>
              <tbody>
                {#if pegawaiList.length === 0}
                  <tr><td colspan="6" class="empty-state">Belum ada data pegawai.</td></tr>
                {:else}
                  {#each pegawaiList as p}
                    <tr>
                      <td>{p.nip || '-'}</td>
                      <td><strong>{p.nama}</strong><br><small style="color: #64748b;">Mulai: {formatTanggalID(p.tanggal_mulai_kerja)}</small></td>
                      <td>{p.jabatan}</td>
                      <td class="text-right">Rp {p.gaji_pokok.toLocaleString('id-ID')}</td>
                      <td class="text-right" style="color: #16a34a; font-weight: bold;">Rp {hitungGajiAktual(p).toLocaleString('id-ID')}</td>
                      <td class="text-center">
                        <button class="btn-icon" on:click={() => editPegawai(p)} title="Edit" style="margin-right: 5px;">
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#2563eb" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
                        </button>
                        <button class="btn-icon action-delete" on:click={() => hapusPegawai(p)} title="Hapus">
                          <div class="css-icon-trash"></div>
                        </button>
                      </td>
                    </tr>
                  {/each}
                {/if}
              </tbody>
            </table>
          </div>

          {:else if currentPage === 'form_pegawai'}
            <div class="page-header" style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem;">
              <div>
                <h1 style="margin:0;">{editingPegawaiId ? 'Edit Pegawai' : 'Tambah Pegawai Baru'}</h1>
                <p style="margin: 0.2rem 0 0; color:#64748b; font-size:0.9rem;">Kelola data pegawai/guru</p>
              </div>
              <button class="kledo-btn-outline" on:click={() => currentPage = 'pegawai'}>&larr; Kembali</button>
            </div>

            <div class="card" style="max-width: 600px; margin: 0 auto; padding: 2rem;">
              <div class="form-group">
                <label>NIP/NIK (Opsional)</label>
                <input type="text" class="kledo-input" bind:value={pegawaiForm.nip} />
              </div>
              <div class="form-group">
                <label>Nama Pegawai</label>
                <input type="text" class="kledo-input" bind:value={pegawaiForm.nama} />
              </div>
              <div class="form-group">
                <label>Jabatan</label>
                <input type="text" class="kledo-input" bind:value={pegawaiForm.jabatan} />
              </div>
              <div class="form-group">
                <label>Tanggal Mulai Kerja</label>
                <input type="date" class="kledo-input" bind:value={pegawaiForm.tanggal_mulai_kerja} />
              </div>
              <div style="display: flex; gap: 1rem;">
                <div class="form-group" style="flex: 1;">
                  <label>Gaji Pokok Awal (Rp)</label>
                  <input type="number" class="kledo-input" bind:value={pegawaiForm.gaji_pokok} />
                </div>
                <div class="form-group" style="flex: 1;">
                  <label>Nominal Kenaikan (Rp)</label>
                  <input type="number" class="kledo-input" bind:value={pegawaiForm.nominal_kenaikan} placeholder="Msl: 100000" />
                </div>
                <div class="form-group" style="flex: 1;">
                  <label>Per Setiap (Tahun)</label>
                  <input type="number" class="kledo-input" bind:value={pegawaiForm.periode_kenaikan} min="1" />
                </div>
              </div>
              
              <hr style="border:0; border-top:1px solid #e2e8f0; margin: 2rem 0;" />
              <h3 style="margin-top:0; font-size:1rem;">Tambahan / Tunjangan Lainnya</h3>
              {#each pegawaiForm.tunjangan as t, i}
                <div style="display: flex; gap: 1rem; align-items: flex-end; margin-bottom: 1rem;">
                  <div class="form-group" style="flex: 2; margin: 0;">
                    <label>Nama Tambahan</label>
                    <input type="text" class="kledo-input" bind:value={t.nama} placeholder="Misal: Tunjangan Transport" />
                  </div>
                  <div class="form-group" style="flex: 1.5; margin: 0;">
                    <label>Nominal (Rp)</label>
                    <input type="number" class="kledo-input" bind:value={t.nominal} />
                  </div>
                  <button class="kledo-btn-danger" style="margin-bottom: 0.2rem;" on:click={() => removeKomponenGaji('tunjangan', i)}>Hapus</button>
                </div>
              {/each}
              <button class="kledo-btn-outline" style="margin-bottom: 1.5rem;" on:click={() => addKomponenGaji('tunjangan')}>+ Tambah Tunjangan</button>

              <hr style="border:0; border-top:1px solid #e2e8f0; margin: 2rem 0;" />
              <h3 style="margin-top:0; font-size:1rem;">Pengurangan / Potongan</h3>
              {#each pegawaiForm.potongan as p, i}
                <div style="display: flex; gap: 1rem; align-items: flex-end; margin-bottom: 1rem;">
                  <div class="form-group" style="flex: 2; margin: 0;">
                    <label>Nama Potongan</label>
                    <input type="text" class="kledo-input" bind:value={p.nama} placeholder="Misal: Cicilan Koperasi" />
                  </div>
                  <div class="form-group" style="flex: 1.5; margin: 0;">
                    <label>Nominal (Rp)</label>
                    <input type="number" class="kledo-input" bind:value={p.nominal} />
                  </div>
                  <button class="kledo-btn-danger" style="margin-bottom: 0.2rem;" on:click={() => removeKomponenGaji('potongan', i)}>Hapus</button>
                </div>
              {/each}
              <button class="kledo-btn-outline" style="margin-bottom: 1.5rem;" on:click={() => addKomponenGaji('potongan')}>+ Tambah Potongan</button>

              <hr style="border:0; border-top:1px solid #e2e8f0; margin: 2rem 0;" />
              <div class="form-actions-kledo" style="margin-top: 2rem; display: flex; gap: 1rem; justify-content: flex-start;">
                <button class="kledo-btn-primary" on:click={savePegawai}>Simpan Pegawai</button>
                <button class="kledo-btn-outline" on:click={() => currentPage = 'pegawai'}>Batal</button>
              </div>
            </div>

        {:else if currentPage === 'penggajian'}
          <div class="page-header">
            <h1>Proses Penggajian</h1>
            <p>Otomatisasi pembayaran gaji guru dan staf {activeEntitas.nama}.</p>
          </div>

          <div class="kledo-card" style="margin-bottom: 2rem;">
            <div style="display:flex; gap: 1rem; align-items: flex-end; margin-bottom: 1.5rem;">
              <div class="form-group" style="margin: 0;">
                <label>Pilih Bulan</label>
                <select class="kledo-input" bind:value={prosesGajiForm.bulan} on:change={loadRiwayatGaji}>
                  {#each Array(12) as _, i}
                    <option value={i+1}>Bulan {i+1} {processedMonths.includes(i+1) ? '✓ (Selesai)' : ''}</option>
                  {/each}
                </select>
              </div>
              <div class="form-group" style="margin: 0;">
                <label>Tahun</label>
                <input type="number" class="kledo-input" bind:value={prosesGajiForm.tahun} on:change={loadRiwayatGaji} />
              </div>
              <div class="form-group" style="margin: 0; flex: 1;">
                <label>Akun Sumber Dana (Kas/Bank)</label>
                <select class="kledo-input" bind:value={prosesGajiForm.rekening_bank_id}>
                  <option value="">-- Pilih Rekening Pembayaran --</option>
                  {#each rekeningList as rek}
                    <option value={rek.id}>{rek.nama_rekening} {rek.nomor_rekening ? '('+rek.nomor_rekening+')' : ''} - Rp {rek.saldo.toLocaleString('id-ID')}</option>
                  {/each}
                </select>
              </div>
            </div>

            <div style="background: #f8fafc; padding: 1.5rem; border-radius: 8px; border: 1px dashed #cbd5e1; display:flex; justify-content:space-between; align-items:center; flex-wrap: wrap; gap: 1rem;">
              <div style="flex: 1; min-width: 250px;">
                <h3 style="margin-top:0; color:#0f172a; margin-bottom: 0.5rem;">Rekapitulasi Gaji Bulan {prosesGajiForm.bulan} / {prosesGajiForm.tahun}</h3>
                <p style="margin:0 0 0.25rem 0; color:#64748b;">Total Pegawai: <strong>{pegawaiList.length} orang</strong></p>
                <p style="margin:0; color:#64748b;">Total Dana Dibutuhkan: <strong>Rp {pegawaiList.reduce((acc, curr) => acc + hitungGajiAktual(curr), 0).toLocaleString('id-ID')}</strong></p>
              </div>
              <div>
                {#if riwayatGaji.length > 0}
                  <button class="kledo-btn-outline" disabled style="color: #16a34a; border-color: #16a34a;">
                    <div class="css-check"></div> Sudah Diproses
                  </button>
                {:else}
                  <button class="kledo-btn-primary" on:click={prosesGaji} disabled={pegawaiList.length === 0} style="white-space: nowrap;">
                    Proses & Bayarkan Gaji
                  </button>
                {/if}
              </div>
            </div>
          </div>

          {#if riwayatGaji.length > 0}
            <div class="card">
              <h3 style="margin: 0 0 1rem 0; padding: 1rem 1rem 0 1rem;">Riwayat Pembayaran Bulan Ini</h3>
              <table class="data-table">
                <thead>
                  <tr>
                    <th>Tanggal Bayar</th>
                    <th>Nama Pegawai</th>
                    <th class="text-right">Total Dibayarkan</th>
                  </tr>
                </thead>
                <tbody>
                  {#each riwayatGaji as r}
                    <tr>
                      <td>{formatTanggalID(r.tanggal_bayar)}</td>
                      <td>{r.pegawai.nama}</td>
                      <td class="text-right" style="color: #16a34a; font-weight: 500;">Rp {r.total_gaji.toLocaleString('id-ID')}</td>
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          {/if}

        {:else if ['laporan', 'neraca', 'aktivitas', 'aruskas'].includes(currentPage)}
          <div class="app-header" style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; background: white; padding: 1.5rem 2rem; border-radius: 16px; box-shadow: 0 4px 20px rgba(0, 0, 0, 0.03);">
              <div class="app-title">
                  <h1 style="font-size: 1.5rem; font-weight: 700; margin-bottom: 0.25rem; margin-top: 0;">Laporan Keuangan Konsolidasi</h1>
                  <p style="font-size: 0.9rem; color: #64748b; margin: 0;">Berdasarkan Standar ISAK 35</p>
              </div>
              <div class="header-actions" style="display: flex; gap: 0.75rem; align-items: center; flex-wrap: wrap;">
                  <select class="kledo-input" bind:value={laporanMode} on:change={onModeChange} style="width: 130px; margin-bottom: 0; font-weight: 600; background-color: #f8fafc;">
                      <option value="bulanan">📅 Bulanan</option>
                      <option value="kuartal">📊 Kuartal</option>
                      <option value="tahunan">📁 Tahunan</option>
                  </select>
                  {#if laporanMode === 'bulanan'}
                      <select class="kledo-input" bind:value={laporanPeriode} on:change={fetchLaporanKonsolidasi} style="width: 135px; margin-bottom: 0;">
                          <option value={1}>Januari</option>
                          <option value={2}>Februari</option>
                          <option value={3}>Maret</option>
                          <option value={4}>April</option>
                          <option value={5}>Mei</option>
                          <option value={6}>Juni</option>
                          <option value={7}>Juli</option>
                          <option value={8}>Agustus</option>
                          <option value={9}>September</option>
                          <option value={10}>Oktober</option>
                          <option value={11}>November</option>
                          <option value={12}>Desember</option>
                      </select>
                  {:else if laporanMode === 'kuartal'}
                      <select class="kledo-input" bind:value={laporanPeriode} on:change={fetchLaporanKonsolidasi} style="width: 170px; margin-bottom: 0;">
                          <option value={1}>Kuartal I (Jan-Mar)</option>
                          <option value={2}>Kuartal II (Apr-Jun)</option>
                          <option value={3}>Kuartal III (Jul-Sep)</option>
                          <option value={4}>Kuartal IV (Okt-Des)</option>
                      </select>
                  {/if}
                  <select class="kledo-input" bind:value={laporanTahun} on:change={fetchLaporanKonsolidasi} style="width: 95px; margin-bottom: 0;">
                      <option value={2026}>2026</option>
                      <option value={2025}>2025</option>
                      <option value={2024}>2024</option>
                  </select>
                  <button class="kledo-btn-primary" on:click={() => currentPage = 'print_laporan'} title="Cetak Laporan">Cetak Laporan</button>
              </div>
          </div>

          <div class="tabs" style="display: flex; gap: 0.5rem; margin-bottom: 1.5rem; overflow-x: auto; padding-bottom: 0.5rem;">
              <button class="tab-btn {activeLaporanTab === 'neraca' ? 'active' : ''}" style="background: {activeLaporanTab === 'neraca' ? '#3b82f6' : 'white'}; color: {activeLaporanTab === 'neraca' ? 'white' : '#64748b'}; padding: 0.75rem 1.5rem; border: 1px solid #e2e8f0; border-radius: 10px; cursor: pointer; font-weight: 500;" on:click={() => { activeLaporanTab = 'neraca'; currentPage = 'neraca'; }}>Posisi Keuangan</button>
              <button class="tab-btn {activeLaporanTab === 'aktivitas' ? 'active' : ''}" style="background: {activeLaporanTab === 'aktivitas' ? '#3b82f6' : 'white'}; color: {activeLaporanTab === 'aktivitas' ? 'white' : '#64748b'}; padding: 0.75rem 1.5rem; border: 1px solid #e2e8f0; border-radius: 10px; cursor: pointer; font-weight: 500;" on:click={() => { activeLaporanTab = 'aktivitas'; currentPage = 'aktivitas'; }}>Penghasilan Komprehensif</button>
              <button class="tab-btn {activeLaporanTab === 'aruskas' ? 'active' : ''}" style="background: {activeLaporanTab === 'aruskas' ? '#3b82f6' : 'white'}; color: {activeLaporanTab === 'aruskas' ? 'white' : '#64748b'}; padding: 0.75rem 1.5rem; border: 1px solid #e2e8f0; border-radius: 10px; cursor: pointer; font-weight: 500;" on:click={() => { activeLaporanTab = 'aruskas'; currentPage = 'aruskas'; }}>Arus Kas</button>
          </div>

          {#if !laporanData}
             <div class="card"><p class="text-center" style="color: #64748b;">Memuat data laporan konsolidasi...</p></div>
          {:else}
             {#if activeLaporanTab === 'neraca'}
                <div class="report-section active card" style="padding: 2.5rem; background: white; border-radius: 16px; box-shadow: 0 10px 30px rgba(0,0,0,0.04);">
                    <div class="report-title" style="text-align: center; margin-bottom: 2.5rem; padding-bottom: 1.5rem; border-bottom: 2px dashed #f1f5f9;">
                        <h2 style="font-size: 1.35rem; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 0.5rem; margin-top: 0;">Laporan Posisi Keuangan - {activeEntitas ? activeEntitas.nama : 'Konsolidasi'}</h2>
                        <p style="font-size: 0.95rem; color: #475569; font-weight: 600; margin: 0;">
                            {laporanMode === 'bulanan' ? `Per Akhir Bulan ${['','Januari','Februari','Maret','April','Mei','Juni','Juli','Agustus','September','Oktober','November','Desember'][laporanPeriode] || ''} ${laporanTahun}` : laporanMode === 'kuartal' ? `Per Akhir Kuartal ${['','I (Jan-Mar)','II (Apr-Jun)','III (Jul-Sep)','IV (Okt-Des)'][laporanPeriode] || ''} ${laporanTahun}` : `Per Akhir Tahun Buku ${laporanTahun}`}
                        </p>
                    </div>
                    <div style="overflow-x: auto;">
                        <table style="width: 100%; border-collapse: separate; border-spacing: 0; font-size: 0.85rem;">
                            <thead>
                                <tr>
                                    <th style="padding: 1rem; text-align: left; background-color: #f8fafc; font-weight: 600; color: #64748b; text-transform: uppercase; font-size: 0.75rem; border-bottom: 1px solid #f1f5f9; border-top-left-radius: 8px;">Uraian</th>
                                    {#each laporanData.entitas_names as entitasName}
                                        <th style="padding: 1rem; text-align: right; background-color: #f8fafc; font-weight: 600; color: #64748b; text-transform: uppercase; font-size: 0.75rem; border-bottom: 1px solid #f1f5f9;">{entitasName}</th>
                                    {/each}
                                </tr>
                            </thead>
                            <tbody>
                                {#each laporanData.neraca as grup}
                                    <tr><td colspan="{laporanData.entitas_names.length + 1}" style="padding: 1rem; font-weight: 700; color: #2563eb; background-color: #f8fafc; border-bottom: 1px solid #f1f5f9;">{grup.nama_grup}</td></tr>
                                    {#each grup.items as item}
                                        <tr class="laporan-row">
                                            <td style="padding: 1rem; border-bottom: 1px solid #f1f5f9; padding-left: 2.5rem; position: relative;">
                                                <span style="position: absolute; left: 1.25rem; color: #64748b;">•</span>
                                                {item.uraian}
                                            </td>
                                            {#each laporanData.entitas_names as entitasName}
                                                <td style="padding: 1rem; text-align: right; border-bottom: 1px solid #f1f5f9; {item.is_red ? 'color: #ef4444;' : ''}">
                                                    {(item.values[entitasName] || 0) !== 0 ? (item.is_red ? `(Rp ${(item.values[entitasName] || 0).toLocaleString('id-ID')})` : `Rp ${(item.values[entitasName] || 0).toLocaleString('id-ID')}`) : '-'}
                                                </td>
                                            {/each}
                                        </tr>
                                    {/each}
                                    <tr style="background-color: {grup.total.is_bold ? 'rgba(59, 130, 246, 0.05)' : '#f1f5f9'}; font-weight: 700;">
                                        <td style="padding: 1rem; border-top: 2px solid {grup.total.is_bold ? '#3b82f6' : '#e2e8f0'}; {grup.total.is_bold ? 'color: #1e293b;' : ''}">{grup.total.uraian}</td>
                                        {#each laporanData.entitas_names as entitasName}
                                            <td style="padding: 1rem; text-align: right; border-top: 2px solid {grup.total.is_bold ? '#3b82f6' : '#e2e8f0'};">
                                                {(grup.total.values[entitasName] || 0) !== 0 ? `Rp ${(grup.total.values[entitasName] || 0).toLocaleString('id-ID')}` : '-'}
                                            </td>
                                        {/each}
                                    </tr>
                                    <tr><td colspan="{laporanData.entitas_names.length + 1}" style="border: none; padding: 0.5rem;"></td></tr>
                                {/each}
                                
                                <!-- TOTAL PASIVA (LIABILITAS + ASET NETO) & CEK BALANCE -->
                                <tr style="background-color: #f0fdf4; font-weight: 800; font-size: 0.95rem;">
                                    <td style="padding: 1.2rem 1rem; border-top: 3px double #16a34a; color: #15803d;">
                                        TOTAL LIABILITAS & ASET NETO (MODAL + HUTANG)
                                    </td>
                                    {#each laporanData.entitas_names as entitasName}
                                        <td style="padding: 1.2rem 1rem; text-align: right; border-top: 3px double #16a34a; color: #15803d;">
                                            {(() => {
                                                const aset = laporanData.neraca[0]?.total?.values[entitasName] || 0;
                                                const pasiva = (laporanData.neraca[1]?.total?.values[entitasName] || 0) + (laporanData.neraca[2]?.total?.values[entitasName] || 0);
                                                const isBalance = Math.abs(aset - pasiva) < 1;
                                                return `Rp ${pasiva.toLocaleString('id-ID')} ${isBalance ? '✅ (BALANCE)' : '⚠️ (TIDAK BALANCE)'}`;
                                            })()}
                                        </td>
                                    {/each}
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
             {:else if activeLaporanTab === 'aktivitas'}
                <div class="report-section active card" style="padding: 2.5rem; background: white; border-radius: 16px; box-shadow: 0 10px 30px rgba(0,0,0,0.04);">
                    <div class="report-title" style="text-align: center; margin-bottom: 2.5rem; padding-bottom: 1.5rem; border-bottom: 2px dashed #f1f5f9;">
                        <h2 style="font-size: 1.35rem; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 0.5rem; margin-top: 0;">Laporan Penghasilan Komprehensif - {activeEntitas ? activeEntitas.nama : 'Konsolidasi'}</h2>
                        <p style="font-size: 0.95rem; color: #475569; font-weight: 600; margin: 0;">
                            {laporanMode === 'bulanan' ? `Periode Bulan ${['','Januari','Februari','Maret','April','Mei','Juni','Juli','Agustus','September','Oktober','November','Desember'][laporanPeriode] || ''} ${laporanTahun}` : laporanMode === 'kuartal' ? `Periode Kuartal ${['','I (Jan - Mar)','II (Apr - Jun)','III (Jul - Sep)','IV (Okt - Des)'][laporanPeriode] || ''} ${laporanTahun}` : `Periode Tahun Buku ${laporanTahun}`}
                        </p>
                    </div>
                    <div style="overflow-x: auto;">
                        <table style="width: 100%; border-collapse: separate; border-spacing: 0; font-size: 0.85rem;">
                            <thead>
                                <tr>
                                    <th style="padding: 1rem; text-align: left; background-color: #f8fafc; font-weight: 600; color: #64748b; text-transform: uppercase; font-size: 0.75rem; border-bottom: 1px solid #f1f5f9; border-top-left-radius: 8px;">Uraian</th>
                                    {#each laporanData.entitas_names as entitasName}
                                        <th style="padding: 1rem; text-align: right; background-color: #f8fafc; font-weight: 600; color: #64748b; text-transform: uppercase; font-size: 0.75rem; border-bottom: 1px solid #f1f5f9;">{entitasName}</th>
                                    {/each}
                                </tr>
                            </thead>
                            <tbody>
                                {#each laporanData.aktivitas as grup}
                                    <tr><td colspan="{laporanData.entitas_names.length + 1}" style="padding: 1rem; font-weight: 700; color: #2563eb; background-color: #f8fafc; border-bottom: 1px solid #f1f5f9;">{grup.nama_grup}</td></tr>
                                    {#each grup.items as item}
                                        <tr class="laporan-row">
                                            <td style="padding: 1rem; border-bottom: 1px solid #f1f5f9; padding-left: 2.5rem; position: relative;">
                                                <span style="position: absolute; left: 1.25rem; color: #64748b;">•</span>
                                                {item.uraian}
                                            </td>
                                            {#each laporanData.entitas_names as entitasName}
                                                <td style="padding: 1rem; text-align: right; border-bottom: 1px solid #f1f5f9;">
                                                    {(item.values[entitasName] || 0) !== 0 ? `Rp ${(item.values[entitasName] || 0).toLocaleString('id-ID')}` : '-'}
                                                </td>
                                            {/each}
                                        </tr>
                                    {/each}
                                    <tr style="background-color: {grup.total.is_bold ? 'rgba(59, 130, 246, 0.05)' : '#f1f5f9'}; font-weight: 700;">
                                        <td style="padding: 1rem; border-top: 2px solid {grup.total.is_bold ? '#3b82f6' : '#e2e8f0'}; {grup.total.is_bold ? 'color: #1e293b;' : ''}">{grup.total.uraian}</td>
                                        {#each laporanData.entitas_names as entitasName}
                                            <td style="padding: 1rem; text-align: right; border-top: 2px solid {grup.total.is_bold ? '#3b82f6' : '#e2e8f0'};">
                                                {(grup.total.values[entitasName] || 0) !== 0 ? `Rp ${(grup.total.values[entitasName] || 0).toLocaleString('id-ID')}` : '-'}
                                            </td>
                                        {/each}
                                    </tr>
                                    <tr><td colspan="{laporanData.entitas_names.length + 1}" style="border: none; padding: 0.5rem;"></td></tr>
                                {/each}
                                
                                <!-- SURPLUS (DEFISIT) -->
                                <tr style="background-color: rgba(59, 130, 246, 0.05); font-weight: 700;">
                                    <td style="padding: 1rem; border-top: 2px solid #3b82f6; color: #1e293b;">SURPLUS (DEFISIT) PERIODE BERJALAN</td>
                                    {#each laporanData.entitas_names as entitasName}
                                        <td style="padding: 1rem; text-align: right; border-top: 2px solid #3b82f6;">
                                            {(() => {
                                                const p = laporanData.aktivitas[0].total.values[entitasName] || 0;
                                                const b = laporanData.aktivitas[1].total.values[entitasName] || 0;
                                                const surplus = p - b;
                                                return surplus !== 0 ? (surplus < 0 ? `(Rp ${Math.abs(surplus).toLocaleString('id-ID')})` : `Rp ${surplus.toLocaleString('id-ID')}`) : '-';
                                            })()}
                                        </td>
                                    {/each}
                                </tr>
                            </tbody>
                        </table>
                    </div>
                </div>
             {:else if activeLaporanTab === 'aruskas'}
                <div class="report-section active card" style="padding: 2.5rem; background: white; border-radius: 16px; box-shadow: 0 10px 30px rgba(0,0,0,0.04);">
                    <div class="report-title" style="text-align: center; margin-bottom: 2.5rem; padding-bottom: 1.5rem; border-bottom: 2px dashed #f1f5f9;">
                        <h2 style="font-size: 1.35rem; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 0.5rem; margin-top: 0;">Laporan Arus Kas - {activeEntitas ? activeEntitas.nama : 'Konsolidasi'}</h2>
                        <p style="font-size: 0.95rem; color: #475569; font-weight: 600; margin: 0;">
                            {laporanMode === 'bulanan' ? `Periode Bulan ${['','Januari','Februari','Maret','April','Mei','Juni','Juli','Agustus','September','Oktober','November','Desember'][laporanPeriode] || ''} ${laporanTahun}` : laporanMode === 'kuartal' ? `Periode Kuartal ${['','I (Jan - Mar)','II (Apr - Jun)','III (Jul - Sep)','IV (Okt - Des)'][laporanPeriode] || ''} ${laporanTahun}` : `Periode Tahun Buku ${laporanTahun}`}
                        </p>
                    </div>
                    <div style="overflow-x: auto;">
                        <table style="width: 100%; border-collapse: separate; border-spacing: 0; font-size: 0.85rem;">
                            <thead>
                                <tr>
                                    <th style="padding: 1rem; text-align: left; background-color: #f8fafc; font-weight: 600; color: #64748b; text-transform: uppercase; font-size: 0.75rem; border-bottom: 1px solid #f1f5f9; border-top-left-radius: 8px;">Uraian</th>
                                    {#each laporanData.entitas_names as entitasName}
                                        <th style="padding: 1rem; text-align: right; background-color: #f8fafc; font-weight: 600; color: #64748b; text-transform: uppercase; font-size: 0.75rem; border-bottom: 1px solid #f1f5f9;">{entitasName}</th>
                                    {/each}
                                </tr>
                            </thead>
                            <tbody>
                                {#each (laporanData.arus_kas || []) as grup}
                                    <tr><td colspan="{laporanData.entitas_names.length + 1}" style="padding: 1rem; font-weight: 700; color: #2563eb; background-color: #f8fafc; border-bottom: 1px solid #f1f5f9;">{grup.nama_grup}</td></tr>
                                    {#each grup.items as item}
                                        <tr class="laporan-row">
                                            <td style="padding: 1rem; border-bottom: 1px solid #f1f5f9; padding-left: 2.5rem; position: relative;">
                                                <span style="position: absolute; left: 1.25rem; color: #64748b;">•</span>
                                                {item.uraian}
                                            </td>
                                            {#each laporanData.entitas_names as entitasName}
                                                <td style="padding: 1rem; text-align: right; border-bottom: 1px solid #f1f5f9; {item.is_red || (item.values[entitasName] || 0) < 0 ? 'color: #ef4444;' : ''}">
                                                    {(() => {
                                                        const val = item.values[entitasName] || 0;
                                                        if (val === 0) return '-';
                                                        if (item.is_red || val < 0) return `(Rp ${Math.abs(val).toLocaleString('id-ID')})`;
                                                        return `Rp ${val.toLocaleString('id-ID')}`;
                                                    })()}
                                                </td>
                                            {/each}
                                        </tr>
                                    {/each}
                                    <tr style="background-color: {grup.total.is_bold ? 'rgba(59, 130, 246, 0.05)' : '#f1f5f9'}; font-weight: 700;">
                                        <td style="padding: 1rem; border-top: 2px solid {grup.total.is_bold ? '#3b82f6' : '#e2e8f0'}; {grup.total.is_bold ? 'color: #1e293b;' : ''}">{grup.total.uraian}</td>
                                        {#each laporanData.entitas_names as entitasName}
                                            <td style="padding: 1rem; text-align: right; border-top: 2px solid {grup.total.is_bold ? '#3b82f6' : '#e2e8f0'}; {((grup.total.values[entitasName] || 0) < 0) ? 'color: #ef4444;' : ''}">
                                                {(() => {
                                                    const val = grup.total.values[entitasName] || 0;
                                                    if (val === 0) return '-';
                                                    return val < 0 ? `(Rp ${Math.abs(val).toLocaleString('id-ID')})` : `Rp ${val.toLocaleString('id-ID')}`;
                                                })()}
                                            </td>
                                        {/each}
                                    </tr>
                                    <tr><td colspan="{laporanData.entitas_names.length + 1}" style="border: none; padding: 0.5rem;"></td></tr>
                                {/each}
                            </tbody>
                        </table>
                    </div>
                </div>
             {/if}
          {/if}

        {:else if currentPage === 'print_laporan'}
          <div class="print-preview-container" style="max-width: 950px; margin: 0 auto; text-align: left;">
              <!-- ACTION BAR (NO-PRINT) -->
              <div class="no-print" style="display: flex; justify-content: space-between; align-items: center; background: #1e293b; color: white; padding: 1rem 1.5rem; border-radius: 12px; margin-bottom: 2rem; box-shadow: 0 4px 15px rgba(0,0,0,0.1);">
                  <div style="display: flex; align-items: center; gap: 1rem;">
                      <button class="kledo-btn-outline" style="background: rgba(255,255,255,0.1); color: white; border: 1px solid rgba(255,255,255,0.2); padding: 0.6rem 1.2rem;" on:click={() => currentPage = activeLaporanTab || 'neraca'}>
                          ⬅ Kembali ke Laporan
                      </button>
                      <div>
                          <h3 style="margin: 0; font-size: 1.1rem; color: white;">Pratayang Resmi Siap Cetak Kertas</h3>
                          <p style="margin: 0; font-size: 0.85rem; color: #94a3b8;">Format Standar A4/F4 dengan Kop Yayasan & Kolom Tanda Tangan</p>
                      </div>
                  </div>
                  <button class="btn btn-primary" style="font-size: 1rem; padding: 0.75rem 1.5rem; background: #10b981; box-shadow: 0 4px 12px rgba(16,185,129,0.3);" on:click={() => window.print()}>
                      🖨️ Cetak Sekarang ke Kertas (Ctrl+P)
                  </button>
              </div>

              <!-- PRINT PAPER AREA -->
              <div class="print-paper" style="background: white; color: black; padding: 3rem 4rem; border-radius: 4px; box-shadow: 0 10px 35px rgba(0,0,0,0.1); font-family: 'Times New Roman', Times, serif; min-height: 1100px;">
                  <!-- KOP SURAT YAYASAN / PESANTREN -->
                  <div style="text-align: center; border-bottom: 3px double #000; padding-bottom: 15px; margin-bottom: 25px;">
                      <h1 style="font-size: 18pt; font-weight: bold; margin: 0; text-transform: uppercase; letter-spacing: 1px;">YAYASAN PONDOK PESANTREN IMAM NAWAWI</h1>
                      <h2 style="font-size: 15pt; font-weight: bold; margin: 6px 0 0 0; text-transform: uppercase;">
                          {activeEntitas ? activeEntitas.nama : 'KONSOLIDASI GABUNGAN (YAYASAN & SMP)'}
                      </h2>
                      <p style="font-size: 10pt; margin: 6px 0 0 0; font-style: italic; font-family: Arial, sans-serif;">
                          Alamat: Jl. Raya Pesantren No. 99, Kabupaten Bogor, Jawa Barat | Email: keuangan@nawawi.or.id | Telp: (021) 8901234
                      </p>
                  </div>

                  <!-- JUDUL LAPORAN -->
                  <div style="text-align: center; margin-bottom: 25px;">
                      <h3 style="font-size: 14pt; font-weight: bold; margin: 0; text-transform: uppercase; text-decoration: underline; font-family: Arial, sans-serif;">
                          {activeLaporanTab === 'neraca' ? 'LAPORAN POSISI KEUANGAN (NERACA)' : activeLaporanTab === 'aktivitas' ? 'LAPORAN PENGHASILAN KOMPREHENSIF' : 'LAPORAN ARUS KAS'}
                      </h3>
                      <p style="font-size: 11pt; margin: 6px 0 0 0; font-weight: bold; font-family: Arial, sans-serif;">
                          {laporanMode === 'bulanan' ? `Periode Bulan ${['','Januari','Februari','Maret','April','Mei','Juni','Juli','Agustus','September','Oktober','November','Desember'][laporanPeriode] || ''} ${laporanTahun}` : laporanMode === 'kuartal' ? `Periode Kuartal ${['','I (Jan-Mar)','II (Apr-Jun)','III (Jul-Sep)','IV (Okt-Des)'][laporanPeriode] || ''} ${laporanTahun}` : `Periode Tahun Buku ${laporanTahun}`}
                      </p>
                  </div>

                  <!-- TABEL LAPORAN (NERACA / AKTIVITAS / ARUS KAS) -->
                  <table style="width: 100%; border-collapse: collapse; font-family: Arial, sans-serif; font-size: 10pt; margin-bottom: 40px;">
                      <thead>
                          <tr style="background-color: #f2f2f2;">
                              <th style="padding: 8px 10px; border: 1px solid #000; text-align: center; font-weight: bold; text-transform: uppercase;">Uraian</th>
                              {#each (laporanData?.entitas_names || []) as entitasName}
                                  <th style="padding: 8px 10px; border: 1px solid #000; text-align: center; font-weight: bold; text-transform: uppercase;">{entitasName}</th>
                              {/each}
                          </tr>
                      </thead>
                      <tbody>
                          {#if activeLaporanTab === 'neraca' && laporanData?.neraca}
                              {#each laporanData.neraca as grup}
                                  <tr>
                                      <td colspan="{(laporanData.entitas_names?.length || 0) + 1}" style="padding: 8px 10px; font-weight: bold; background-color: #f9f9f9; border: 1px solid #000;">
                                          {grup.nama_grup}
                                      </td>
                                  </tr>
                                  {#each grup.items as item}
                                      <tr>
                                          <td style="padding: 6px 10px 6px 25px; border: 1px solid #000;">
                                              • {item.uraian}
                                          </td>
                                          {#each laporanData.entitas_names as entitasName}
                                              <td style="padding: 6px 10px; border: 1px solid #000; text-align: right;">
                                                  {(item.values[entitasName] || 0) !== 0 ? (item.is_red ? `(Rp ${(item.values[entitasName] || 0).toLocaleString('id-ID')})` : `Rp ${(item.values[entitasName] || 0).toLocaleString('id-ID')}`) : '-'}
                                              </td>
                                          {/each}
                                      </tr>
                                  {/each}
                                  <tr style="background-color: #eaeaea; font-weight: bold;">
                                      <td style="padding: 8px 10px; border: 1px solid #000;">{grup.total.uraian}</td>
                                      {#each laporanData.entitas_names as entitasName}
                                          <td style="padding: 8px 10px; border: 1px solid #000; text-align: right;">
                                              {(grup.total.values[entitasName] || 0) !== 0 ? `Rp ${(grup.total.values[entitasName] || 0).toLocaleString('id-ID')}` : '-'}
                                          </td>
                                      {/each}
                                  </tr>
                              {/each}
                              <!-- TOTAL PASIVA (LIABILITAS + ASET NETO) -->
                              <tr style="background-color: #e2e8f0; font-weight: bold; font-size: 10.5pt;">
                                  <td style="padding: 10px; border: 2px solid #000;">
                                      TOTAL LIABILITAS & ASET NETO
                                  </td>
                                  {#each laporanData.entitas_names as entitasName}
                                      <td style="padding: 10px; text-align: right; border: 2px solid #000;">
                                          {(() => {
                                              const pasiva = (laporanData.neraca[1]?.total?.values[entitasName] || 0) + (laporanData.neraca[2]?.total?.values[entitasName] || 0);
                                              return `Rp ${pasiva.toLocaleString('id-ID')}`;
                                          })()}
                                      </td>
                                  {/each}
                              </tr>

                          {:else if activeLaporanTab === 'aktivitas' && laporanData?.aktivitas}
                              {#each laporanData.aktivitas as grup}
                                  <tr>
                                      <td colspan="{(laporanData.entitas_names?.length || 0) + 1}" style="padding: 8px 10px; font-weight: bold; background-color: #f9f9f9; border: 1px solid #000;">
                                          {grup.nama_grup}
                                      </td>
                                  </tr>
                                  {#each grup.items as item}
                                      <tr>
                                          <td style="padding: 6px 10px 6px 25px; border: 1px solid #000;">
                                              • {item.uraian}
                                          </td>
                                          {#each laporanData.entitas_names as entitasName}
                                              <td style="padding: 6px 10px; border: 1px solid #000; text-align: right;">
                                                  {(item.values[entitasName] || 0) !== 0 ? `Rp ${(item.values[entitasName] || 0).toLocaleString('id-ID')}` : '-'}
                                              </td>
                                          {/each}
                                      </tr>
                                  {/each}
                                  <tr style="background-color: #eaeaea; font-weight: bold;">
                                      <td style="padding: 8px 10px; border: 1px solid #000;">{grup.total.uraian}</td>
                                      {#each laporanData.entitas_names as entitasName}
                                          <td style="padding: 8px 10px; border: 1px solid #000; text-align: right;">
                                              {(grup.total.values[entitasName] || 0) !== 0 ? `Rp ${(grup.total.values[entitasName] || 0).toLocaleString('id-ID')}` : '-'}
                                          </td>
                                      {/each}
                                  </tr>
                              {/each}
                              <!-- SURPLUS / DEFISIT -->
                              <tr style="background-color: #e2e8f0; font-weight: bold; font-size: 10.5pt;">
                                  <td style="padding: 10px; border: 2px solid #000;">SURPLUS (DEFISIT) PERIODE BERJALAN</td>
                                  {#each laporanData.entitas_names as entitasName}
                                      <td style="padding: 10px; text-align: right; border: 2px solid #000;">
                                          {(() => {
                                              const p = laporanData.aktivitas[0]?.total?.values[entitasName] || 0;
                                              const b = laporanData.aktivitas[1]?.total?.values[entitasName] || 0;
                                              const surplus = p - b;
                                              if (surplus === 0) return '-';
                                              return surplus < 0 ? `(Rp ${Math.abs(surplus).toLocaleString('id-ID')})` : `Rp ${surplus.toLocaleString('id-ID')}`;
                                          })()}
                                      </td>
                                  {/each}
                              </tr>

                          {:else if activeLaporanTab === 'aruskas' && laporanData?.arus_kas}
                              {#each laporanData.arus_kas as grup}
                                  <tr>
                                      <td colspan="{(laporanData.entitas_names?.length || 0) + 1}" style="padding: 8px 10px; font-weight: bold; background-color: #f9f9f9; border: 1px solid #000;">
                                          {grup.nama_grup}
                                      </td>
                                  </tr>
                                  {#each grup.items as item}
                                      <tr>
                                          <td style="padding: 6px 10px 6px 25px; border: 1px solid #000;">
                                              • {item.uraian}
                                          </td>
                                          {#each laporanData.entitas_names as entitasName}
                                              <td style="padding: 6px 10px; border: 1px solid #000; text-align: right;">
                                                  {(() => {
                                                      const val = item.values[entitasName] || 0;
                                                      if (val === 0) return '-';
                                                      if (item.is_red || val < 0) return `(Rp ${Math.abs(val).toLocaleString('id-ID')})`;
                                                      return `Rp ${val.toLocaleString('id-ID')}`;
                                                  })()}
                                              </td>
                                          {/each}
                                      </tr>
                                  {/each}
                                  <tr style="background-color: #eaeaea; font-weight: bold;">
                                      <td style="padding: 8px 10px; border: 1px solid #000;">{grup.total.uraian}</td>
                                      {#each laporanData.entitas_names as entitasName}
                                          <td style="padding: 8px 10px; border: 1px solid #000; text-align: right;">
                                              {(() => {
                                                  const val = grup.total.values[entitasName] || 0;
                                                  if (val === 0) return '-';
                                                  return val < 0 ? `(Rp ${Math.abs(val).toLocaleString('id-ID')})` : `Rp ${val.toLocaleString('id-ID')}`;
                                              })()}
                                          </td>
                                      {/each}
                                  </tr>
                              {/each}
                          {/if}
                      </tbody>
                  </table>

                  <!-- KOLOM TANDA TANGAN -->
                  <div class="print-signature-section" style="margin-top: 50px; width: 100%; page-break-inside: avoid; font-family: Arial, sans-serif; font-size: 11pt;">
                      <div style="display: flex; justify-content: space-between; padding: 0 40px;">
                          <div style="text-align: center; width: 40%;">
                              <p style="margin: 0 0 70px 0;">Mengetahui,<br><b>Ketua Yayasan / Kepala Sekolah</b></p>
                              <p style="text-decoration: underline; font-weight: bold; margin: 0;">KH. Ahmad Zainuddin, Lc., M.A.</p>
                              <p style="margin: 3px 0 0 0; font-size: 10pt;">NIP. 19780512 200501 1 003</p>
                          </div>
                          <div style="text-align: center; width: 45%;">
                              <p style="margin: 0 0 70px 0;">Bogor, {new Date().toLocaleDateString('id-ID', {day: '2-digit', month: 'long', year: 'numeric'})}<br><b>Dibuat oleh,<br>Kepala Bagian Keuangan</b></p>
                              <p style="text-decoration: underline; font-weight: bold; margin: 0;">Ustadz Muhammad Faisal, S.E.</p>
                              <p style="margin: 3px 0 0 0; font-size: 10pt;">NIP. 19890115 201207 1 001</p>
                          </div>
                      </div>
                  </div>
              </div>
          </div>

        {:else if currentPage === 'backup'}
          <div class="page-header" style="text-align: left;">
            <h1>Cadangkan & Pulihkan Database</h1>
            <p>Amankan data akuntansi Anda secara berkala ke Flashdisk, harddisk eksternal, atau cloud storage.</p>
          </div>

          <!-- KARTU INFORMASI DATABASE AKTIF -->
          <div style="background: white; border-radius: 12px; padding: 1.5rem; margin-bottom: 1.5rem; box-shadow: 0 1px 3px rgba(0,0,0,0.05); border: 1px solid #e2e8f0; text-align: left;">
            <div style="display: flex; align-items: center; gap: 0.75rem; margin-bottom: 1rem;">
              <div style="width: 40px; height: 40px; border-radius: 8px; background: #ecfdf5; color: #059669; display: flex; align-items: center; justify-content: center; font-size: 1.25rem; font-weight: bold;">
                🗄️
              </div>
              <div>
                <h3 style="margin: 0; font-size: 1.05rem; color: #1e293b;">Database Aktif Sistem (SQLite WAL)</h3>
                <span style="font-size: 0.8rem; color: #10b981; font-weight: 600;">● Terhubung & Terenkapsulasi Aman di AppData</span>
              </div>
            </div>

            <div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 1rem; background: #f8fafc; padding: 1rem; border-radius: 8px; font-size: 0.88rem;">
              <div>
                <span style="color: #64748b; font-size: 0.75rem; text-transform: uppercase; font-weight: 700;">Lokasi Penyimpanan:</span>
                <p style="margin: 4px 0 0 0; font-family: monospace; font-size: 0.8rem; color: #334155; word-break: break-all;">{dbInfo.path || 'Memuat lokasi...'}</p>
              </div>
              <div>
                <span style="color: #64748b; font-size: 0.75rem; text-transform: uppercase; font-weight: 700;">Ukuran Database:</span>
                <p style="margin: 4px 0 0 0; font-weight: 700; color: #0f172a;">{dbInfo.size}</p>
              </div>
              <div>
                <span style="color: #64748b; font-size: 0.75rem; text-transform: uppercase; font-weight: 700;">Terakhir Diperbarui:</span>
                <p style="margin: 4px 0 0 0; color: #334155;">{dbInfo.lastModified}</p>
              </div>
            </div>
          </div>

          <!-- DUA KARTU AKSI UTAMA -->
          <div style="display: grid; grid-template-columns: 1fr 1fr; gap: 1.5rem; margin-bottom: 2rem;">
            <!-- KARTU CADANGKAN -->
            <div style="background: white; border-radius: 12px; padding: 2rem; box-shadow: 0 1px 3px rgba(0,0,0,0.05); border: 1px solid #e2e8f0; display: flex; flex-direction: column; justify-content: space-between; text-align: left;">
              <div>
                <div style="width: 48px; height: 48px; border-radius: 10px; background: #eff6ff; color: #2563eb; display: flex; align-items: center; justify-content: center; font-size: 1.5rem; margin-bottom: 1rem;">
                  📥
                </div>
                <h2 style="font-size: 1.25rem; color: #1e293b; margin: 0 0 0.5rem 0;">Cadangkan Data (Backup)</h2>
                <p style="color: #64748b; line-height: 1.6; font-size: 0.9rem; margin-bottom: 1.5rem;">
                  Salin seluruh catatan akuntansi, jurnal, daftar akun, data siswa, rekening, dan gaji ke dalam satu file berkas database (<code style="background: #f1f5f9; padding: 2px 6px; border-radius: 4px;">.db</code>). Simpan file ini di tempat yang aman sebagai arsip.
                </p>
              </div>
              <button class="kledo-btn-primary" style="width: 100%; justify-content: center; padding: 0.85rem;" on:click={handleBackupDatabase} disabled={isBackingUp}>
                {isBackingUp ? 'Sedang Memproses...' : '📥 Mulai Cadangkan Database (.db)'}
              </button>
            </div>

            <!-- KARTU PULIHKAN -->
            <div style="background: white; border-radius: 12px; padding: 2rem; box-shadow: 0 1px 3px rgba(0,0,0,0.05); border: 1px solid #fed7aa; display: flex; flex-direction: column; justify-content: space-between; text-align: left;">
              <div>
                <div style="width: 48px; height: 48px; border-radius: 10px; background: #fff7ed; color: #ea580c; display: flex; align-items: center; justify-content: center; font-size: 1.5rem; margin-bottom: 1rem;">
                  🔄
                </div>
                <h2 style="font-size: 1.25rem; color: #1e293b; margin: 0 0 0.5rem 0;">Pulihkan Data (Restore)</h2>
                <p style="color: #64748b; line-height: 1.6; font-size: 0.9rem; margin-bottom: 1.5rem;">
                  Pulihkan pembukuan dari file cadangan yang pernah dibuat sebelumnya. <b style="color: #dc2626;">Perhatian:</b> Seluruh data saat ini akan digantikan oleh isi file cadangan. Sebelum pemulihan, sistem akan otomatis membuat salinan darurat.
                </p>
              </div>
              <button class="kledo-btn-primary" style="width: 100%; justify-content: center; padding: 0.85rem; background-color: #ea580c;" on:click={handleRestoreDatabase} disabled={isRestoring}>
                {isRestoring ? 'Sedang Memulihkan...' : '🔄 Pilih & Pulihkan File (.db)'}
              </button>
            </div>
          </div>

          <!-- PANDUAN PRAKTIS & KEAMANAN -->
          <div style="background: #f8fafc; border-left: 4px solid #0e935a; border-radius: 0 8px 8px 0; padding: 1rem 1.5rem; text-align: left;">
            <h4 style="margin: 0 0 0.5rem 0; color: #0f172a; font-size: 0.95rem;">💡 Tips Keamanan Data Keuangan:</h4>
            <ul style="margin: 0; padding-left: 1.2rem; color: #475569; font-size: 0.85rem; line-height: 1.6;">
              <li>Lakukan pencadangan (backup) secara rutin, minimal seminggu sekali atau setiap akhir bulan setelah tutup buku.</li>
              <li>Simpan salinan file cadangan di luar laptop (seperti Google Drive, Flashdisk khusus bendahara, atau cloud storage yayasan).</li>
              <li>Jika berganti laptop baru, cukup install aplikasi di laptop baru lalu gunakan menu <b>Pulihkan Data (Restore)</b> dengan memilih file cadangan terakhir Anda.</li>
            </ul>
          </div>
        {/if}
    </main>
  </div>
{/if}

<!-- GLOBAL TOAST NOTIFICATION -->
{#if toastMsg}
  <div class="toast-container">
    <div class="toast {toastType}">
      <div class="toast-icon">
        {#if toastType === 'success'}
          <div class="css-check"></div>
        {:else if toastType === 'error'}
          <div class="css-cross"></div>
        {:else}
          <div style="font-family: serif; font-weight: bold; margin-top: -2px;">i</div>
        {/if}
      </div>
      <div class="toast-message">{toastMsg}</div>
    </div>
  </div>
{/if}

<!-- GLOBAL CONFIRM MODAL -->
{#if confirmModal}
  <div class="modal-overlay">
    <div class="modal-content confirm-box">
      <div class="confirm-icon-wrap">
        <div style="font-weight: bold; margin-top: -4px;">!</div>
      </div>
      <h2>{confirmModal.title}</h2>
      <p>{confirmModal.message}</p>
      <div class="form-actions-kledo confirm-actions">
        <button class="kledo-btn-primary" style="background-color: #ef4444;" on:click={confirmModal.onConfirm}>Ya, Lanjutkan</button>
        <button class="kledo-btn-outline" on:click={confirmModal.onCancel}>Batal</button>
      </div>
    </div>
  </div>
{/if}

<style>
  :global(body) {
    margin: 0;
    padding: 0;
    font-family: 'Inter', system-ui, -apple-system, sans-serif;
    background-color: #f8fafc;
    color: #334155;
  }

  .layout {
    display: flex;
    height: 100vh;
    width: 100vw;
    overflow: hidden;
  }

  /* --- SIDEBAR GREEN PREMIUM --- */
  .sidebar-green {
    width: 250px;
    background-color: #0e935a; /* Emerald green ala Odoo/Equip */
    color: #ffffff;
    display: flex;
    flex-direction: column;
    box-shadow: 2px 0 10px rgba(0,0,0,0.1);
    overflow-y: auto;
  }

  .sidebar-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1.5rem 1.25rem 1rem 1.25rem;
  }

  .brand-text {
    font-size: 1.7rem;
    font-weight: 800;
    letter-spacing: -1px;
    color: white;
  }



  /* Unit Selector Box (White Dropdown) */
  .unit-selector {
    margin: 0 1rem 1.5rem 1rem;
    background-color: white;
    color: #0e935a;
    border: none;
    border-radius: 4px;
    padding: 0.6rem 1rem;
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-weight: 600;
    font-size: 0.95rem;
    cursor: pointer;
    box-shadow: 0 2px 5px rgba(0,0,0,0.1);
    transition: background-color 0.2s;
  }

  .unit-selector:hover {
    background-color: #f1f5f9;
  }

  .unit-selector .chevron-right {
    opacity: 0.6;
  }

  /* Navigation Sidebar */
  .sidebar-nav {
    display: flex;
    flex-direction: column;
  }

  .nav-group {
    border-bottom: 1px solid rgba(255,255,255,0.2);
    padding: 0.75rem 0;
  }
  .nav-group:last-child {
    border-bottom: none;
  }

  .nav-group-title {
    padding: 0.5rem 1.25rem;
    font-size: 0.9rem;
    font-weight: 700;
    opacity: 0.95;
    letter-spacing: 0.2px;
  }

  .nav-item {
    display: flex;
    align-items: center;
    gap: 12px;
    padding: 0.6rem 1.25rem 0.6rem 1.8rem;
    background: transparent;
    border: none;
    color: rgba(255,255,255,0.8);
    cursor: pointer;
    font-size: 0.9rem;
    width: 100%;
    text-align: left;
    transition: all 0.2s;
    font-family: inherit;
  }

  .nav-item:hover {
    color: white;
    background: rgba(255,255,255,0.1);
  }

  .nav-item.active {
    color: white;
    font-weight: 600;
    background: rgba(0,0,0,0.15);
  }

  .nav-icon {
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0.7;
    margin-top: 1px;
  }
  .nav-item.active .nav-icon { opacity: 1; }

  /* --- SETUP SCREEN STYLES (SPLIT-SCREEN + ILUSTRASI) --- */
  .setup-layout { display: flex; min-height: 100vh; width: 100vw; }
  
  .setup-left {
    flex: 0 0 35%;
    background-color: #0e935a;
    color: white;
    padding: 4rem;
    display: flex;
    flex-direction: column;
    justify-content: center;
  }
  
  .brand-huge { font-size: 4rem; font-weight: 800; letter-spacing: -2px; margin-bottom: 1.5rem; }
  .setup-left h2 { font-size: 1.8rem; margin: 0 0 1rem 0; font-weight: 600; line-height: 1.3; }
  .setup-left p { font-size: 1.1rem; opacity: 0.8; line-height: 1.6; }

  .setup-right {
    flex: 1;
    background-color: #f4f6f9; /* Kledo gray */
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem;
  }

  .setup-content {
    width: 100%;
    max-width: 900px; /* Made wider for cards */
  }

  .section-title {
    margin-top: 0;
    margin-bottom: 2rem;
    font-size: 1.8rem;
    color: #111827;
    font-weight: 700;
  }

  .kledo-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
    gap: 1.5rem;
  }

  .kledo-card {
    background-color: #ffffff;
    border: 1px solid #ffffff;
    border-radius: 12px;
    padding: 2.5rem 1.5rem;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    cursor: pointer;
    box-shadow: 0 4px 10px rgba(0,0,0,0.03);
    transition: all 0.2s ease;
  }

  .kledo-card:hover {
    box-shadow: 0 10px 25px rgba(0,0,0,0.08);
    transform: translateY(-4px);
    border-color: #0e935a;
  }

  .card-illustration {
    height: 100px;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-bottom: 1.5rem;
  }

  /* CUSTOM ICONS */
  .css-icon-wrapper { width: 60px; height: 60px; border-radius: 50%; display: flex; justify-content: center; align-items: center; margin-bottom: 1rem; }
  .green-theme { background: #dcfce7; }
  .blue-theme { background: #e0e7ff; }
  .dashed-theme { border: 2px dashed #94a3b8; background: transparent; transition: all 0.2s; }

  /* CSS Eye Icon for Detail */
  .css-icon-eye { position: relative; width: 16px; height: 10px; border-radius: 50%; border: 2px solid currentColor; margin-top: 2px; }
  .css-icon-eye::after { content: ''; position: absolute; top: 50%; left: 50%; width: 4px; height: 4px; background: currentColor; border-radius: 50%; transform: translate(-50%, -50%); }

  /* CSS Pencil Icon for Edit */
  .css-icon-pencil { position: relative; width: 14px; height: 14px; transform: rotate(45deg); border: 2px solid currentColor; border-radius: 2px; margin-top: 4px; }
  .css-icon-pencil::before { content: ''; position: absolute; top: -6px; left: -2px; width: 0; height: 0; border-left: 6px solid transparent; border-right: 6px solid transparent; border-bottom: 6px solid currentColor; }

  /* CSS Trash Icon */
  
  .kledo-card:hover .dashed-theme {
    background-color: #dcfce7;
    border-color: #16a34a;
  }

  /* CSS Manual: Sekolah */
  .css-sekolah { display: flex; flex-direction: column; align-items: center; margin-top: 10px; }
  .css-sekolah .roof { 
    width: 0; height: 0; 
    border-left: 25px solid transparent; 
    border-right: 25px solid transparent; 
    border-bottom: 20px solid #16a34a; 
  }
  .css-sekolah .building { 
    width: 40px; height: 35px; 
    background-color: #22c55e; 
    position: relative; 
    display: flex; flex-direction: column; align-items: center; 
  }
  .css-sekolah .windows { display: flex; gap: 8px; margin-top: 6px; }
  .css-sekolah .win { width: 8px; height: 8px; background-color: #dcfce7; border-radius: 1px; }
  .css-sekolah .door { width: 12px; height: 14px; background-color: #15803d; position: absolute; bottom: 0; border-top-left-radius: 4px; border-top-right-radius: 4px; }

  /* CSS Manual: Perusahaan */
  .css-perusahaan { display: flex; align-items: flex-end; justify-content: center; height: 50px; }
  .css-perusahaan .tower { 
    width: 34px; height: 45px; 
    background-color: #0284c7; 
    border-top-left-radius: 4px; border-top-right-radius: 4px; 
    display: flex; flex-direction: column; align-items: center; padding-top: 6px; 
    position: relative;
  }
  .css-perusahaan .win-row { display: flex; gap: 6px; margin-bottom: 5px; }
  .css-perusahaan .win { width: 6px; height: 6px; background-color: #e0f2fe; }
  .css-perusahaan .door { width: 10px; height: 12px; background-color: #075985; position: absolute; bottom: 0; }

  /* CSS Manual: Plus */
  .css-plus { position: relative; width: 30px; height: 30px; }
  .css-plus .v-line { position: absolute; left: 13px; top: 0; width: 4px; height: 30px; background-color: #94a3b8; border-radius: 2px; }
  .css-plus .h-line { position: absolute; left: 0; top: 13px; width: 30px; height: 4px; background-color: #94a3b8; border-radius: 2px; }
  
  .kledo-card:hover .css-plus .v-line, .kledo-card:hover .css-plus .h-line {
    background-color: #16a34a;
  }

  .kledo-card h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1.15rem;
    font-weight: 700;
    color: #1f2937;
  }

  .kledo-card p {
    margin: 0;
    font-size: 0.85rem;
    color: #6b7280;
  }

  .kledo-card.new-card {
    background-color: transparent;
    border: 2px dashed #d1d5db;
    box-shadow: none;
  }

  .kledo-card.new-card:hover {
    background-color: #ffffff;
    border-color: #0e935a;
  }

  .kledo-form-container {
    background: #ffffff;
    padding: 3rem;
    border-radius: 16px;
    box-shadow: 0 10px 30px rgba(0,0,0,0.03);
    max-width: 500px;
    margin: 0 auto;
  }

  .kledo-form-container h2 {
    margin-top: 0;
    font-size: 1.5rem;
    color: #111827;
  }

  .kledo-form-container p {
    color: #6b7280;
    margin-bottom: 2rem;
  }

  .kledo-input {
    width: 100%;
    padding: 1rem 1.25rem;
    border: 1px solid #d1d5db;
    border-radius: 8px;
    font-size: 1rem;
    margin-bottom: 1.25rem;
    outline: none;
    transition: border-color 0.2s;
    font-family: inherit;
    box-sizing: border-box;
  }

  .kledo-input:focus {
    border-color: #0e935a;
  }

  .form-actions-kledo {
    display: flex;
    gap: 1rem;
    margin-top: 1rem;
  }

  .kledo-btn-primary {
    background-color: #0e935a;
    color: white;
    border: none;
    padding: 0.85rem 1.5rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .kledo-btn-primary:hover {
    background-color: #0b784a;
  }

  .btn-icon:hover {
    background: #f1f5f9;
    color: #334155;
  }
  
  .action-view { color: #64748b; }
  .action-view:hover { background-color: #f1f5f9; color: #334155; }
  
  .action-edit { color: #64748b; }
  .action-edit:hover { background-color: #fef3c7; color: #92400e; }
  
  .action-delete { color: #64748b; }
  .action-delete:hover { background-color: #fef2f2; color: #991b1b; }

  .kledo-btn-outline {
    background-color: transparent;
    color: #4b5563;
    border: 1px solid #d1d5db;
    padding: 0.85rem 1.5rem;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s;
  }

  .kledo-btn-outline:hover {
    background-color: #f3f4f6;
  }

  /* MAIN CONTENT STYLES */
  .main-content { 
    flex: 1; 
    padding: 2rem 3rem; 
    overflow-y: auto; 
    transform: translateZ(0); /* Force hardware acceleration */
    will-change: scroll-position;
  }
  .page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 2rem; }
  .page-header h1 { font-size: 1.75rem; font-weight: 700; color: #0f172a; margin: 0 0 0.5rem 0; }
  .page-header p { margin: 0; color: #64748b; }
  .card-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 1.5rem; margin-bottom: 2rem; }
  .card { 
    background: white; 
    border-radius: 16px; 
    padding: 1.5rem; 
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.02); /* Dikurangi untuk performa */
    border: 1px solid #e2e8f0; 
    transform: translateZ(0); 
  }
  .summary-card h3 { margin: 0 0 1rem 0; font-size: 0.95rem; color: #64748b; }
  .summary-card .amount { font-size: 2rem; font-weight: 700; color: #0f172a; }
  .report-card { display: flex; flex-direction: column; align-items: flex-start; }
  .report-card h3 { margin: 0 0 0.5rem 0; color: #1e293b; }
  .report-card p { color: #64748b; font-size: 0.9rem; margin: 0 0 1.5rem 0; line-height: 1.5; flex: 1; }
  
  .laporan-row { transition: background-color 0.2s; }
  .laporan-row:hover { background-color: #f8fafc; }
  
  .btn { display: flex; align-items: center; justify-content: center; gap: 8px; padding: 0.6rem 1.25rem; border-radius: 8px; font-weight: 600; font-size: 0.9rem; cursor: pointer; transition: all 0.2s ease; border: none; }
  .btn-primary { background: #0e935a; color: white; box-shadow: 0 4px 10px rgba(14, 147, 90, 0.2); }
  .btn-primary:hover { background: #0a7346; transform: translateY(-1px); }
  .btn-outline { background: transparent; border: 1px solid #cbd5e1; color: #475569; }
  .btn-outline:hover { border-color: #94a3b8; background: #f8fafc; }

  .data-table { width: 100%; border-collapse: collapse; }
  .data-table th, .data-table td { padding: 1rem; text-align: left; border-bottom: 1px solid #e2e8f0; }
  .data-table th { font-weight: 600; color: #64748b; font-size: 0.85rem; text-transform: uppercase; }
  .data-table td { font-size: 0.95rem; color: #1e293b; }
  .empty-state { text-align: center; padding: 3rem; color: #94a3b8; font-style: italic; }
  .text-right { text-align: right !important; }

  /* ACCOUNT UI EXTRAS */
  .header-actions { display: flex; gap: 0.5rem; align-items: center; }
  .btn-icon { background: none; border: none; cursor: pointer; opacity: 0.7; transition: opacity 0.2s; padding: 8px; display: flex; align-items: center; justify-content: center; }
  .btn-icon:hover { opacity: 1; transform: scale(1.1); }
  
  /* MINI CSS ICONS FOR BUTTONS */
  .css-icon-star { width: 10px; height: 10px; background: currentColor; clip-path: polygon(50% 0%, 61% 35%, 98% 35%, 68% 57%, 79% 91%, 50% 70%, 21% 91%, 32% 57%, 2% 35%, 39% 35%); }
  .css-icon-arrow-up { width: 0; height: 0; border-left: 5px solid transparent; border-right: 5px solid transparent; border-bottom: 7px solid currentColor; margin-top: -2px; }
  .css-icon-arrow-down { width: 0; height: 0; border-left: 5px solid transparent; border-right: 5px solid transparent; border-top: 7px solid currentColor; margin-top: 2px; }
  .css-icon-plus-small { width: 12px; height: 12px; position: relative; }
  .css-icon-plus-small::before { content: ''; position: absolute; left: 5px; top: 0; width: 2px; height: 12px; background: currentColor; }
  .css-icon-plus-small::after { content: ''; position: absolute; left: 0; top: 5px; width: 12px; height: 2px; background: currentColor; }
  
  .css-icon-pencil { width: 12px; height: 4px; border: 1px solid currentColor; background: currentColor; transform: rotate(-45deg); position: relative; border-radius: 1px; }
  .css-icon-pencil::before { content: ''; position: absolute; left: -5px; top: -1px; width: 0; height: 0; border-top: 2px solid transparent; border-bottom: 2px solid transparent; border-right: 4px solid currentColor; }
  
  .css-icon-trash { width: 10px; height: 12px; border: 1.5px solid currentColor; border-top: none; position: relative; border-bottom-left-radius: 2px; border-bottom-right-radius: 2px; margin-top: 2px; }
  .css-icon-trash::before { content: ''; position: absolute; top: -3px; left: -2px; width: 11px; height: 1.5px; background: currentColor; }
  .css-icon-trash::after { content: ''; position: absolute; top: -5px; left: 2px; width: 3px; height: 2px; border: 1.5px solid currentColor; border-bottom: none; }
  
  /* MODAL */
  .modal-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; background: rgba(0,0,0,0.5); display: flex; align-items: center; justify-content: center; z-index: 1000; }
  .modal-content { background: white; padding: 2rem; border-radius: 12px; width: 100%; max-width: 500px; box-shadow: 0 10px 25px rgba(0,0,0,0.1); }
  .modal-content h2 { margin-top: 0; margin-bottom: 1.5rem; }
  .form-group { margin-bottom: 1rem; }
  .form-group label { display: block; margin-bottom: 0.5rem; font-weight: 600; font-size: 0.9rem; color: #4b5563; }
  .form-row { display: flex; gap: 1rem; }
  .form-row .form-group { flex: 1; }

  /* --- CUSTOM NOTIFICATIONS --- */
  .toast-container { position: fixed; bottom: 2rem; right: 2rem; z-index: 9999; animation: slideUp 0.3s ease-out; }
  @keyframes slideUp { from { transform: translateY(20px); opacity: 0; } to { transform: translateY(0); opacity: 1; } }
  .toast { display: flex; align-items: center; gap: 12px; background: white; padding: 1rem 1.5rem; border-radius: 12px; box-shadow: 0 10px 30px rgba(0,0,0,0.1); border-left: 4px solid #3b82f6; }
  .toast.success { border-left-color: #10b981; }
  .toast.error { border-left-color: #ef4444; }
  .toast-icon { width: 24px; height: 24px; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: white; font-size: 14px; }
  .toast.success .toast-icon { background: #10b981; }
  .toast.error .toast-icon { background: #ef4444; }
  .toast.info .toast-icon { background: #3b82f6; }
  .toast-message { font-weight: 600; color: #1e293b; font-size: 0.95rem; }
  
  .css-check { width: 6px; height: 10px; border-bottom: 2px solid white; border-right: 2px solid white; transform: rotate(45deg); margin-bottom: 2px; }
  .css-cross { width: 12px; height: 12px; position: relative; }
  .css-cross::before, .css-cross::after { content: ''; position: absolute; left: 5px; top: 0; width: 2px; height: 12px; background: white; }
  .css-cross::before { transform: rotate(45deg); }
  .css-cross::after { transform: rotate(-45deg); }
  
  /* CONFIRM BOX */
  .confirm-box { text-align: center; max-width: 400px; padding: 2.5rem; }
  .confirm-icon-wrap { width: 60px; height: 60px; background: #fef2f2; color: #ef4444; border-radius: 50%; display: flex; align-items: center; justify-content: center; margin: 0 auto 1.5rem auto; font-size: 2rem; font-weight: bold; }
  .confirm-box h2 { font-size: 1.3rem; margin-bottom: 0.5rem; }
  .confirm-box p { color: #64748b; line-height: 1.5; margin-bottom: 2rem; }
  .confirm-actions { justify-content: center; }

  /* JURNAL UI STYLES */
  .jurnal-modal { max-width: 800px; }
  .jurnal-grid-container { border: 1px solid #e2e8f0; border-radius: 8px; margin-bottom: 1rem; overflow: hidden; }
  .jurnal-table { width: 100%; border-collapse: collapse; }
  .jurnal-table th { background: #f8fafc; padding: 0.75rem 1rem; text-align: left; font-size: 0.85rem; color: #475569; border-bottom: 1px solid #e2e8f0; }
  .jurnal-table td { padding: 0.5rem; border-bottom: 1px solid #e2e8f0; }
  .jurnal-input { width: 100%; padding: 0.6rem; border: 1px solid transparent; border-radius: 4px; font-family: inherit; font-size: 0.9rem; outline: none; background: #f8fafc; transition: all 0.2s; box-sizing: border-box; }
  .jurnal-input:focus { border-color: #3b82f6; background: white; box-shadow: 0 0 0 2px rgba(59,130,246,0.1); }
  .btn-add-row { width: 100%; padding: 0.75rem; background: #f8fafc; border: none; color: #3b82f6; font-weight: 600; cursor: pointer; transition: background 0.2s; }
  .btn-add-row:hover { background: #f1f5f9; }
  
  .jurnal-footer { display: flex; align-items: center; justify-content: space-between; padding: 1.5rem; background: #f8fafc; border-radius: 8px; margin-bottom: 1.5rem; border: 1px solid #e2e8f0; transition: all 0.3s; }
  .jurnal-footer.balanced { border-color: #10b981; background: #ecfdf5; }
  .jurnal-footer.unbalanced { border-color: #ef4444; background: #fef2f2; }
  .jurnal-total { display: flex; flex-direction: column; }
  .jurnal-total span { font-size: 0.8rem; color: #64748b; margin-bottom: 4px; }
  .jurnal-total strong { font-size: 1.2rem; color: #1e293b; }
  
  .status-badge { padding: 6px 12px; border-radius: 20px; font-weight: 600; font-size: 0.85rem; display: flex; align-items: center; gap: 6px; }
  .status-badge.success { background: #dcfce7; color: #16a34a; }
  .status-badge.error { background: #fee2e2; color: #ef4444; }

  .kledo-btn-primary:disabled { opacity: 0.5; cursor: not-allowed; }
</style>
