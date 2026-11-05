export namespace main {
	
	export class PegawaiInput {
	    entitas_id: number;
	    nip: string;
	    nama: string;
	    jabatan: string;
	    gaji_pokok: number;
	    tanggal_mulai_kerja: string;
	    periode_kenaikan: number;
	    nominal_kenaikan: number;
	    tunjangan: string;
	    potongan: string;
	
	    static createFrom(source: any = {}) {
	        return new PegawaiInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entitas_id = source["entitas_id"];
	        this.nip = source["nip"];
	        this.nama = source["nama"];
	        this.jabatan = source["jabatan"];
	        this.gaji_pokok = source["gaji_pokok"];
	        this.tanggal_mulai_kerja = source["tanggal_mulai_kerja"];
	        this.periode_kenaikan = source["periode_kenaikan"];
	        this.nominal_kenaikan = source["nominal_kenaikan"];
	        this.tunjangan = source["tunjangan"];
	        this.potongan = source["potongan"];
	    }
	}
	export class ProsesGajiInput {
	    entitas_id: number;
	    bulan: number;
	    tahun: number;
	    rekening_bank_id: number;
	    tanggal: string;
	
	    static createFrom(source: any = {}) {
	        return new ProsesGajiInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entitas_id = source["entitas_id"];
	        this.bulan = source["bulan"];
	        this.tahun = source["tahun"];
	        this.rekening_bank_id = source["rekening_bank_id"];
	        this.tanggal = source["tanggal"];
	    }
	}
	export class TransaksiInput {
	    entitas_id: number;
	    tanggal: string;
	    no_referensi: string;
	    keterangan: string;
	    total_debit: number;
	    total_kredit: number;
	    details: models.JurnalDetail[];
	
	    static createFrom(source: any = {}) {
	        return new TransaksiInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entitas_id = source["entitas_id"];
	        this.tanggal = source["tanggal"];
	        this.no_referensi = source["no_referensi"];
	        this.keterangan = source["keterangan"];
	        this.total_debit = source["total_debit"];
	        this.total_kredit = source["total_kredit"];
	        this.details = this.convertValues(source["details"], models.JurnalDetail);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace models {
	
	export class AkunKeuangan {
	    id: number;
	    entitas_id: number;
	    kode_akun: string;
	    nama_akun: string;
	    kategori: string;
	    sifat: string;
	    is_bawaan: boolean;
	    saldo: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new AkunKeuangan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.entitas_id = source["entitas_id"];
	        this.kode_akun = source["kode_akun"];
	        this.nama_akun = source["nama_akun"];
	        this.kategori = source["kategori"];
	        this.sifat = source["sifat"];
	        this.is_bawaan = source["is_bawaan"];
	        this.saldo = source["saldo"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Entitas {
	    id: number;
	    nama: string;
	    tipe: string;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Entitas(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nama = source["nama"];
	        this.tipe = source["tipe"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RekeningBank {
	    id: number;
	    entitas_id: number;
	    nama_rekening: string;
	    nomor_rekening: string;
	    saldo: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new RekeningBank(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.entitas_id = source["entitas_id"];
	        this.nama_rekening = source["nama_rekening"];
	        this.nomor_rekening = source["nomor_rekening"];
	        this.saldo = source["saldo"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Transaksi {
	    id: number;
	    entitas_id: number;
	    tanggal: string;
	    no_referensi: string;
	    keterangan: string;
	    total_debit: number;
	    total_kredit: number;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Transaksi(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.entitas_id = source["entitas_id"];
	        this.tanggal = source["tanggal"];
	        this.no_referensi = source["no_referensi"];
	        this.keterangan = source["keterangan"];
	        this.total_debit = source["total_debit"];
	        this.total_kredit = source["total_kredit"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class JurnalDetail {
	    id: number;
	    transaksi_id: number;
	    transaksi?: Transaksi;
	    akun_id: number;
	    akun?: AkunKeuangan;
	    rekening_bank_id?: number;
	    rekening_bank?: RekeningBank;
	    debit: number;
	    kredit: number;
	
	    static createFrom(source: any = {}) {
	        return new JurnalDetail(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.transaksi_id = source["transaksi_id"];
	        this.transaksi = this.convertValues(source["transaksi"], Transaksi);
	        this.akun_id = source["akun_id"];
	        this.akun = this.convertValues(source["akun"], AkunKeuangan);
	        this.rekening_bank_id = source["rekening_bank_id"];
	        this.rekening_bank = this.convertValues(source["rekening_bank"], RekeningBank);
	        this.debit = source["debit"];
	        this.kredit = source["kredit"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LaporanKonsolidasiItem {
	    uraian: string;
	    values: Record<string, number>;
	    global: number;
	    is_bold: boolean;
	    is_red: boolean;
	
	    static createFrom(source: any = {}) {
	        return new LaporanKonsolidasiItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.uraian = source["uraian"];
	        this.values = source["values"];
	        this.global = source["global"];
	        this.is_bold = source["is_bold"];
	        this.is_red = source["is_red"];
	    }
	}
	export class LaporanKonsolidasiGroup {
	    nama_grup: string;
	    items: LaporanKonsolidasiItem[];
	    total: LaporanKonsolidasiItem;
	
	    static createFrom(source: any = {}) {
	        return new LaporanKonsolidasiGroup(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nama_grup = source["nama_grup"];
	        this.items = this.convertValues(source["items"], LaporanKonsolidasiItem);
	        this.total = this.convertValues(source["total"], LaporanKonsolidasiItem);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class LaporanKonsolidasiResponse {
	    entitas_names: string[];
	    neraca: LaporanKonsolidasiGroup[];
	    aktivitas: LaporanKonsolidasiGroup[];
	    arus_kas: LaporanKonsolidasiGroup[];
	
	    static createFrom(source: any = {}) {
	        return new LaporanKonsolidasiResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entitas_names = source["entitas_names"];
	        this.neraca = this.convertValues(source["neraca"], LaporanKonsolidasiGroup);
	        this.aktivitas = this.convertValues(source["aktivitas"], LaporanKonsolidasiGroup);
	        this.arus_kas = this.convertValues(source["arus_kas"], LaporanKonsolidasiGroup);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Pegawai {
	    id: number;
	    entitas_id: number;
	    nip: string;
	    nama: string;
	    jabatan: string;
	    gaji_pokok: number;
	    tanggal_mulai_kerja: string;
	    periode_kenaikan: number;
	    nominal_kenaikan: number;
	    tunjangan: string;
	    potongan: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Pegawai(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.entitas_id = source["entitas_id"];
	        this.nip = source["nip"];
	        this.nama = source["nama"];
	        this.jabatan = source["jabatan"];
	        this.gaji_pokok = source["gaji_pokok"];
	        this.tanggal_mulai_kerja = source["tanggal_mulai_kerja"];
	        this.periode_kenaikan = source["periode_kenaikan"];
	        this.nominal_kenaikan = source["nominal_kenaikan"];
	        this.tunjangan = source["tunjangan"];
	        this.potongan = source["potongan"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class RiwayatGaji {
	    id: number;
	    pegawai_id: number;
	    pegawai?: Pegawai;
	    bulan: number;
	    tahun: number;
	    total_gaji: number;
	    detail_tunjangan: string;
	    detail_potongan: string;
	    tanggal_bayar: string;
	    transaksi_id: number;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new RiwayatGaji(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.pegawai_id = source["pegawai_id"];
	        this.pegawai = this.convertValues(source["pegawai"], Pegawai);
	        this.bulan = source["bulan"];
	        this.tahun = source["tahun"];
	        this.total_gaji = source["total_gaji"];
	        this.detail_tunjangan = source["detail_tunjangan"];
	        this.detail_potongan = source["detail_potongan"];
	        this.tanggal_bayar = source["tanggal_bayar"];
	        this.transaksi_id = source["transaksi_id"];
	        this.created_at = this.convertValues(source["created_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Siswa {
	    id: number;
	    entitas_id: number;
	    nis: string;
	    nisn: string;
	    nama_lengkap: string;
	    jenis_kelamin: string;
	    tempat_lahir: string;
	    tanggal_lahir: string;
	    agama: string;
	    alamat: string;
	    nama_ayah: string;
	    nama_ibu: string;
	    telepon_ortu: string;
	    kelas: string;
	    jenjang: string;
	    tarif_spp: number;
	    status: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Siswa(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.entitas_id = source["entitas_id"];
	        this.nis = source["nis"];
	        this.nisn = source["nisn"];
	        this.nama_lengkap = source["nama_lengkap"];
	        this.jenis_kelamin = source["jenis_kelamin"];
	        this.tempat_lahir = source["tempat_lahir"];
	        this.tanggal_lahir = source["tanggal_lahir"];
	        this.agama = source["agama"];
	        this.alamat = source["alamat"];
	        this.nama_ayah = source["nama_ayah"];
	        this.nama_ibu = source["nama_ibu"];
	        this.telepon_ortu = source["telepon_ortu"];
	        this.kelas = source["kelas"];
	        this.jenjang = source["jenjang"];
	        this.tarif_spp = source["tarif_spp"];
	        this.status = source["status"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class TagihanSiswa {
	    id: number;
	    entitas_id: number;
	    siswa_id: number;
	    siswa?: Siswa;
	    jenis_tagihan: string;
	    bulan?: number;
	    tahun?: number;
	    total_tagihan: number;
	    sisa_tagihan: number;
	    status: string;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new TagihanSiswa(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.entitas_id = source["entitas_id"];
	        this.siswa_id = source["siswa_id"];
	        this.siswa = this.convertValues(source["siswa"], Siswa);
	        this.jenis_tagihan = source["jenis_tagihan"];
	        this.bulan = source["bulan"];
	        this.tahun = source["tahun"];
	        this.total_tagihan = source["total_tagihan"];
	        this.sisa_tagihan = source["sisa_tagihan"];
	        this.status = source["status"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

