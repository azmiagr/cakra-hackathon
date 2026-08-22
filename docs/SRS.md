# Software Requirements Specification (SRS)

## Cakra — Predictive Replenishment

| Field | Detail |
|---|---|
| **Versi Dokumen** | 1.0 |
| **Status** | Draft — Siap Implementasi MVP Penyisihan |
| **Standar Acuan** | Struktur mengikuti kerangka IEEE 830 (disesuaikan untuk konteks MVP kompetisi) |
| **Dokumen Terkait** | PRD — Cakra: Predictive Replenishment v1.0 |
| **Tanggal** | 21 Agustus 2026 |

---

## 1. Pendahuluan

### 1.1 Tujuan Dokumen

Dokumen ini menspesifikasikan kebutuhan perangkat lunak secara teknis dan presisi untuk pengembangan Cakra — Predictive Replenishment, sebagai turunan langsung dari PRD yang telah disusun sebelumnya. SRS ini menjadi rujukan tunggal bagi tim engineering (frontend, backend, AI/ML) selama periode pengembangan MVP penyisihan.

### 1.2 Ruang Lingkup

Perangkat lunak yang dispesifikasikan adalah aplikasi web dengan satu alur inti: menerima riwayat penjualan historis satu SKU sebagai input, dan menghasilkan rekomendasi keputusan replenishment sebagai output. Sistem tidak mencakup fungsi pencatatan transaksi (POS), otentikasi pengguna, maupun penyimpanan riwayat penggunaan lintas sesi, sesuai batasan yang ditetapkan pada PRD Bagian 4.

### 1.3 Definisi, Akronim, dan Singkatan

| Istilah | Definisi |
|---|---|
| SKU | Stock Keeping Unit — satuan unik yang mengidentifikasi satu jenis barang |
| ADI | Average Demand Interval — rata-rata interval waktu antar-kejadian permintaan non-nol |
| CV² | Coefficient of Variation Squared — kuadrat koefisien variasi, mengukur variabilitas besaran permintaan |
| ROP | Reorder Point — titik stok di mana pemesanan ulang harus dilakukan |
| ROQ | Reorder Quantity — jumlah unit yang direkomendasikan untuk dipesan |
| Lead Time | Waktu antara pemesanan dilakukan hingga barang diterima dari pemasok |
| Service Level | Target probabilitas tidak terjadi stockout selama lead time |
| Intermittent Demand | Pola permintaan dengan banyak periode tanpa transaksi (nol permintaan) |
| SBC | Syntetos-Boylan Categorization — metode klasifikasi pola permintaan berbasis ADI dan CV² |
| MASE | Mean Absolute Scaled Error — metrik evaluasi akurasi forecasting yang sesuai untuk data intermiten |

### 1.4 Referensi

PRD Cakra — Predictive Replenishment v1.0; Ketentuan Batasan Ruang Lingkup MVP, Rulebook COMPFEST 18 AIC; Syntetos, A. & Boylan, J. (2005), *Categorisation for Improved Inventory Forecasting*, untuk metode klasifikasi pola permintaan.

### 1.5 Gambaran Umum Dokumen

Bagian 2 menjelaskan deskripsi umum sistem. Bagian 3 memuat spesifikasi kebutuhan fungsional secara rinci per modul. Bagian 4 memuat kebutuhan antarmuka eksternal (API, UI). Bagian 5 memuat kebutuhan non-fungsional. Bagian 6 memuat struktur data dan skema basis data. Bagian 7 memuat spesifikasi algoritma inti. Bagian 8 memuat use case terperinci. Bagian 9 memuat lampiran formula.

---

## 2. Deskripsi Umum

### 2.1 Perspektif Produk

Cakra adalah sistem berdiri sendiri (standalone), bukan modul tambahan dari sistem lain. Sistem menerima data dari pengguna secara langsung (unggahan berkas), bukan melalui integrasi API dengan sistem POS pihak ketiga pada tahap MVP ini.

### 2.2 Fungsi Utama Produk

1. Menerima dan memvalidasi data riwayat penjualan per SKU.
2. Mengklasifikasikan pola permintaan SKU menggunakan metode Syntetos-Boylan Categorization (SBC).
3. Menghasilkan proyeksi permintaan menggunakan model yang sesuai kategori.
4. Menghitung reorder point dan reorder quantity.
5. Menentukan label risiko SKU.
6. Menyajikan hasil dengan penjelasan yang dapat dipahami pengguna non-teknis.

### 2.3 Karakteristik Pengguna

Pengguna sistem adalah pemilik/pengelola usaha retail-grosir kecil-menengah dengan literasi digital dasar hingga menengah, tidak memiliki latar belakang statistik/data science, mengakses sistem melalui browser desktop atau mobile.

### 2.4 Batasan Umum

Sistem wajib beroperasi synchronous (permintaan-respons tunggal per analisis, tanpa job asinkron/terjadwal). Parameter model bersifat statis selama sesi demonstrasi (tidak ada auto-tuning berjalan pada saat demo). Cakupan antarmuka dibatasi pada dua layar (input dan hasil), tanpa otentikasi maupun riwayat lintas sesi.

### 2.5 Asumsi dan Dependensi

Pengguna memiliki data riwayat penjualan minimal dalam format tabular yang dapat diekspor (CSV) atau bersedia menginput data ringkas secara manual. Sistem bergantung pada ketersediaan pustaka Prophet/LightGBM dan implementasi Croston/TSB pada lingkungan backend Python. Detail asumsi lain mengacu pada PRD Bagian 13 (Risks & Assumptions).

---

## 3. Kebutuhan Fungsional Rinci

### SRS-FR-001: Unggah dan Validasi Data Historis

**Deskripsi:** Sistem menerima berkas CSV berisi riwayat penjualan satu SKU.

**Input:** Berkas CSV dengan kolom wajib: `tanggal` (format YYYY-MM-DD), `jumlah_terjual` (integer ≥ 0). Kolom opsional: `nama_sku`, `harga_satuan`.

**Proses Validasi:**
1. Periksa keberadaan kolom wajib.
2. Periksa format tanggal valid dan berurutan (tidak wajib tanpa jeda, tapi tidak boleh duplikat tanggal untuk SKU yang sama).
3. Periksa `jumlah_terjual` bertipe numerik non-negatif.
4. Hitung total baris data valid.

**Output Sukses:** Data tersimpan sementara, sistem lanjut ke SRS-FR-002.

**Output Gagal:** Pesan kesalahan spesifik per jenis kegagalan (lihat Tabel 3.1).

**Tabel 3.1 — Matriks Pesan Kesalahan**

| Kondisi Gagal | Pesan yang Ditampilkan |
|---|---|
| Kolom `tanggal` atau `jumlah_terjual` tidak ditemukan | "Berkas tidak memiliki kolom [nama kolom]. Pastikan berkas menggunakan format templat yang disediakan." |
| Format tanggal tidak valid pada baris tertentu | "Format tanggal tidak valid pada baris [nomor baris]. Gunakan format YYYY-MM-DD." |
| Nilai `jumlah_terjual` negatif atau bukan angka | "Nilai penjualan tidak valid pada baris [nomor baris]. Harus berupa angka nol atau lebih." |
| Tanggal duplikat untuk SKU yang sama | "Ditemukan tanggal ganda pada baris [nomor baris]. Gabungkan menjadi satu baris per tanggal." |

### SRS-FR-002: Klasifikasi Pola Permintaan (SBC)

**Deskripsi:** Sistem menghitung ADI dan CV² dari data yang telah divalidasi, lalu mengklasifikasikan SKU ke salah satu dari empat kategori permintaan menurut metode Syntetos-Boylan.

**Ambang Klasifikasi (ditetapkan, acuan literatur SBC):**

| Kategori | Kondisi | Jalur Model yang Digunakan |
|---|---|---|
| Smooth (Reguler) | ADI ≤ 1,32 dan CV² ≤ 0,49 | Prophet/LightGBM |
| Erratic | ADI ≤ 1,32 dan CV² > 0,49 | Prophet/LightGBM dengan penyesuaian interval ketidakpastian lebih lebar |
| Intermittent | ADI > 1,32 dan CV² ≤ 0,49 | Croston/TSB |
| Lumpy | ADI > 1,32 dan CV² > 0,49 | Croston/TSB (varian TSB direkomendasikan karena lebih stabil untuk kasus paling sulit ini) |

**Kebutuhan Ambang Data Minimum (Keputusan Open Question #2):**

| Kategori Terdeteksi | Ambang Minimum Data | Jika Tidak Terpenuhi |
|---|---|---|
| Smooth/Erratic | Minimal 90 hari data historis berurutan | Kembalikan status `INSUFFICIENT_DATA` dengan pesan "Data historis minimal 90 hari diperlukan untuk pola permintaan reguler. Saat ini tersedia [n] hari." |
| Intermittent/Lumpy | Minimal 10 kejadian permintaan non-nol | Kembalikan status `INSUFFICIENT_DATA` dengan pesan "Minimal 10 kejadian penjualan diperlukan untuk SKU dengan pola permintaan jarang. Saat ini tersedia [n] kejadian." |

### SRS-FR-003: Peramalan Permintaan

**Deskripsi:** Sistem menjalankan model peramalan sesuai hasil SRS-FR-002.

**Untuk kategori Smooth/Erratic:**
Model Prophet atau LightGBM menghasilkan proyeksi permintaan harian untuk horizon 14 hari ke depan, dengan interval ketidakpastian P50 (median) dan P90 (batas atas konservatif untuk perhitungan safety stock).

**Untuk kategori Intermittent/Lumpy:**
Model Croston (untuk Intermittent) atau TSB (untuk Lumpy) menghasilkan estimasi rata-rata permintaan per periode kejadian dan estimasi rata-rata interval antar-kejadian, dikonversi menjadi estimasi permintaan per hari untuk horizon yang sama (14 hari).

**Output:** Objek proyeksi berisi nilai per hari (P50 dan P90) untuk horizon 14 hari, beserta metadata kategori yang terdeteksi.

### SRS-FR-004: Perhitungan Reorder Point dan Reorder Quantity

**Deskripsi:** Modul logika keputusan (deterministik, bukan bagian dari model ML) mengonversi hasil SRS-FR-003 menjadi rekomendasi konkret.

**Parameter Input Modul (Keputusan Open Question #3):**

| Parameter | Nilai Default MVP | Sumber |
|---|---|---|
| Lead time pemasok | 3 hari | Estimasi rata-rata waktu kirim distributor FMCG lokal |
| Target service level | 95% (z = 1,65) | Standar keseimbangan industri ritel antara risiko stockout dan overstock |

**Formula (lihat juga Bagian 9 — Lampiran Formula):**

```
Demand_LeadTime = rata-rata proyeksi permintaan harian × lead time
Safety_Stock = z × σ_demand × √(lead time)
Reorder_Point (ROP) = Demand_LeadTime + Safety_Stock
Order_Up_To_Level = proyeksi permintaan selama (lead time + review period) + Safety_Stock
Reorder_Quantity (ROQ) = Order_Up_To_Level − Posisi_Stok_Saat_Ini
```

Di mana σ_demand diturunkan dari interval P50–P90 hasil SRS-FR-003 sebagai proksi deviasi standar permintaan.

**Output:** Nilai ROP (dibulatkan ke atas ke bilangan bulat terdekat) dan ROQ (dibulatkan ke atas), keduanya dalam satuan unit.

### SRS-FR-005: Penentuan Label Risiko

**Deskripsi:** Sistem menentukan label risiko berdasarkan aturan yang beroperasi di atas hasil SRS-FR-003 dan SRS-FR-004 (Keputusan Open Question #4).

**Input Tambahan yang Diperlukan:** Posisi stok saat ini (`current_stock`), wajib diinput pengguna pada layar input (field tambahan pada SRS-FR-001).

**Logika Penentuan (dievaluasi berurutan):**

```
Hari_Ketersediaan = current_stock ÷ rata-rata proyeksi permintaan harian (P50)

JIKA Hari_Ketersediaan < Lead_Time:
    Label = "Stockout Imminent" (merah)
ATAU JIKA (total proyeksi permintaan 14 hari < 0.2 × current_stock)
        ATAU (tidak ada kejadian permintaan non-nol dalam periode
              2 × rata-rata interval permintaan historis SKU):
    Label = "Deadstock Candidate" (abu-abu)
SELAIN ITU:
    Label = "Normal" (kuning/hijau)
```

**Output:** Label kategorikal beserta satu kalimat alasan yang dihasilkan dari nilai numerik yang memicu label tersebut (mendukung SRS-FR-006).

### SRS-FR-006: Penyajian Hasil dengan Penjelasan

**Deskripsi:** Sistem menyusun narasi penjelas otomatis berdasarkan nilai-nilai yang dihasilkan SRS-FR-002 hingga SRS-FR-005.

**Template Narasi (contoh, disesuaikan nilai aktual saat runtime):**

> "Berdasarkan riwayat penjualan [n] hari terakhir, [nama_sku] menunjukkan pola permintaan [kategori SBC]. Proyeksi permintaan rata-rata adalah [x] unit/hari. Dengan lead time pemasok [y] hari dan target ketersediaan [service level]%, sebaiknya pemesanan ulang dilakukan saat stok mencapai [ROP] unit, sejumlah [ROQ] unit. Status saat ini: [label risiko] — [alasan singkat sesuai pemicu SRS-FR-005]."

**Kriteria Penerimaan:** Narasi tidak boleh mengandung istilah statistik tanpa penjelasan (mis. tidak menampilkan "MASE" atau "CV²" mentah di antarmuka pengguna akhir; istilah ini hanya muncul di lapisan teknis/log).

---

## 4. Kebutuhan Antarmuka Eksternal

### 4.1 Antarmuka Pengguna (UI)

| Layar | Elemen Wajib |
|---|---|
| Layar Input | Area unggah berkas CSV (drag-and-drop atau klik pilih berkas); tautan unduh templat CSV contoh; field input `nama_sku` (opsional, teks); field input `current_stock` (wajib, numerik); tombol "Analisis"; area pesan validasi real-time |
| Layar Hasil | Nama SKU; badge label risiko dengan warna sesuai SRS-FR-005; angka ROP dan ROQ ditampilkan besar/menonjol; grafik proyeksi permintaan 14 hari sederhana (garis P50, area P50–P90); narasi penjelas sesuai SRS-FR-006; tombol "Analisis SKU Lain" (kembali ke Layar Input) |

### 4.2 Antarmuka Perangkat Lunak (API)

Seluruh endpoint menggunakan protokol REST melalui FastAPI, format `application/json` kecuali endpoint unggah berkas.

**POST /api/v1/analyze**

Request (multipart/form-data):
```
file: <berkas CSV>
sku_name: string (opsional)
current_stock: integer (wajib)
```

Response Sukses (200):
```json
{
  "status": "success",
  "sku_name": "string",
  "demand_category": "smooth | erratic | intermittent | lumpy",
  "forecast": {
    "horizon_days": 14,
    "daily_p50": [/* array 14 nilai numerik */],
    "daily_p90": [/* array 14 nilai numerik */]
  },
  "recommendation": {
    "reorder_point": "integer",
    "reorder_quantity": "integer",
    "lead_time_days": 3,
    "service_level": 0.95
  },
  "risk_label": {
    "label": "stockout_imminent | deadstock_candidate | normal",
    "reason": "string"
  },
  "explanation_text": "string"
}
```

Response Gagal — Validasi (422):
```json
{
  "status": "validation_error",
  "error_code": "MISSING_COLUMN | INVALID_DATE_FORMAT | INVALID_VALUE | DUPLICATE_DATE",
  "message": "string",
  "row_number": "integer (opsional)"
}
```

Response Gagal — Data Tidak Memadai (422):
```json
{
  "status": "insufficient_data",
  "detected_category": "string",
  "required_minimum": "string",
  "actual_value": "string",
  "message": "string"
}
```

**GET /api/v1/template**

Mengembalikan berkas CSV templat contoh untuk diunduh pengguna (`text/csv`).

### 4.3 Antarmuka Perangkat Keras

Tidak berlaku — produk tidak melibatkan komponen perangkat keras pada tahap MVP.

---

## 5. Kebutuhan Non-Fungsional

| ID | Kategori | Spesifikasi |
|---|---|---|
| SRS-NFR-001 | Performa | Waktu respons endpoint `/analyze` tidak melebihi 10 detik untuk berkas hingga 2 tahun data harian (± 730 baris) pada lingkungan demo lokal |
| SRS-NFR-002 | Usability | Pesan kesalahan wajib berbahasa Indonesia, spesifik per baris/kolom, tanpa istilah teknis pemrograman (mis. tidak menampilkan stack trace) |
| SRS-NFR-003 | Explainability | Setiap response sukses wajib menyertakan field `explanation_text` yang tidak kosong |
| SRS-NFR-004 | Reliabilitas Model | Sistem wajib mengembalikan status `insufficient_data` alih-alih memaksakan forecast ketika ambang SRS-FR-002 tidak terpenuhi — dilarang menghasilkan output forecast dengan confidence yang menyesatkan |
| SRS-NFR-005 | Portabilitas | Aplikasi dapat dijalankan sepenuhnya secara lokal melalui `docker compose up`, sesuai README.md, tanpa dependensi layanan cloud eksternal berbayar |
| SRS-NFR-006 | Kepatuhan Batasan Lomba | Tidak ada proses asinkron/terjadwal (background job, cron, queue); seluruh pemrosesan `/analyze` bersifat synchronous request-response |
| SRS-NFR-007 | Maintainability | Modul klasifikasi (SRS-FR-002), forecasting (SRS-FR-003), dan logika keputusan (SRS-FR-004/005) dipisahkan sebagai unit kode independen agar dapat diuji dan diperbarui terpisah |

---

## 6. Struktur Data

### 6.1 Skema Basis Data (PostgreSQL)

**Tabel `analysis_session`**

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID (PK) | Identifier unik sesi analisis |
| sku_name | VARCHAR(255) | Nama SKU (opsional dari input pengguna) |
| current_stock | INTEGER | Posisi stok saat input |
| created_at | TIMESTAMP | Waktu analisis dijalankan |
| demand_category | VARCHAR(20) | Hasil SRS-FR-002 |
| adi_value | FLOAT | Nilai ADI terhitung |
| cv_squared_value | FLOAT | Nilai CV² terhitung |

**Tabel `sales_history`**

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID (PK) | Identifier baris |
| session_id | UUID (FK → analysis_session.id) | Relasi ke sesi analisis |
| sale_date | DATE | Tanggal transaksi |
| quantity_sold | INTEGER | Jumlah unit terjual pada tanggal tersebut |

**Tabel `recommendation_result`**

| Kolom | Tipe | Keterangan |
|---|---|---|
| id | UUID (PK) | Identifier hasil |
| session_id | UUID (FK → analysis_session.id) | Relasi ke sesi analisis |
| reorder_point | INTEGER | Hasil SRS-FR-004 |
| reorder_quantity | INTEGER | Hasil SRS-FR-004 |
| risk_label | VARCHAR(30) | Hasil SRS-FR-005 |
| risk_reason | TEXT | Alasan label |
| explanation_text | TEXT | Narasi hasil SRS-FR-006 |
| forecast_p50 | JSONB | Array proyeksi P50 14 hari |
| forecast_p90 | JSONB | Array proyeksi P90 14 hari |

**Catatan kepatuhan batasan MVP:** meskipun tabel di atas mendukung penyimpanan, sistem MVP tidak menampilkan riwayat lintas sesi kepada pengguna (sesuai PRD Bagian 4.2, poin 3) — penyimpanan ini semata untuk keperluan audit/debugging internal tim selama pengembangan, bukan fitur yang di-expose ke UI.

---

## 7. Spesifikasi Algoritma Inti

### 7.1 Perhitungan ADI dan CV²

```
ADI = jumlah_total_periode ÷ jumlah_kejadian_permintaan_non_nol

CV² = (standar_deviasi_ukuran_permintaan_saat_terjadi ÷ 
       rata_rata_ukuran_permintaan_saat_terjadi)²
```

Kedua metrik dihitung hanya dari periode dengan permintaan non-nol untuk CV² (mengukur variabilitas besaran saat transaksi terjadi), sementara ADI mengukur jarang-tidaknya transaksi terjadi secara keseluruhan.

### 7.2 Model Croston/TSB (Ringkasan Pendekatan)

Metode Croston mendekomposisi data menjadi dua deret terpisah: ukuran permintaan saat terjadi transaksi, dan interval waktu antar-transaksi, masing-masing di-smoothing secara eksponensial terpisah, lalu digabungkan menjadi estimasi permintaan rata-rata per periode. Metode TSB (Teunter-Syntetos-Babai) merupakan varian yang lebih stabil untuk kasus lumpy karena mengestimasi probabilitas kejadian permintaan alih-alih interval, mengurangi bias saat terjadi periode permintaan nol yang sangat panjang.

Implementasi direkomendasikan menggunakan pustaka `statsforecast` (Nixtla) yang menyediakan implementasi Croston Classic, Croston SBA, dan TSB siap pakai, untuk menghindari implementasi manual dari nol yang berisiko tinggi kesalahan dalam periode pengembangan terbatas.

---

## 8. Use Case Terperinci

### UC-01: Analisis SKU dengan Data Memadai (Alur Utama)

**Aktor:** Pengguna (pemilik usaha)

**Prakondisi:** Pengguna memiliki berkas CSV riwayat penjualan yang sesuai format.

**Alur Utama:**
1. Pengguna membuka Layar Input.
2. Pengguna mengunggah berkas CSV dan mengisi `current_stock`.
3. Sistem memvalidasi format (SRS-FR-001) — berhasil.
4. Sistem menghitung ADI/CV² dan mengklasifikasikan kategori (SRS-FR-002).
5. Sistem memverifikasi ambang data minimum terpenuhi.
6. Sistem menjalankan model forecasting sesuai kategori (SRS-FR-003).
7. Sistem menghitung ROP dan ROQ (SRS-FR-004).
8. Sistem menentukan label risiko (SRS-FR-005).
9. Sistem menyusun narasi penjelas (SRS-FR-006).
10. Sistem menampilkan Layar Hasil kepada pengguna.

**Postkondisi:** Pengguna melihat rekomendasi keputusan lengkap untuk SKU yang dianalisis.

### UC-02: Analisis SKU dengan Data Tidak Memadai (Alur Alternatif)

**Alur Alternatif (bercabang dari langkah 5 UC-01):**
5a. Sistem mendeteksi jumlah data di bawah ambang minimum sesuai kategori terdeteksi.
5b. Sistem mengembalikan status `insufficient_data` dengan pesan spesifik jumlah data yang dibutuhkan vs tersedia.
5c. Layar Hasil menampilkan pesan tersebut beserta arahan ("Tambahkan data hingga minimal [n] [hari/kejadian] untuk mendapatkan rekomendasi yang andal"), bukan memaksakan output forecast.

**Postkondisi:** Pengguna memahami kekurangan data tanpa menerima rekomendasi yang berisiko menyesatkan.

### UC-03: Unggah Berkas dengan Format Tidak Valid (Alur Alternatif)

**Alur Alternatif (bercabang dari langkah 3 UC-01):**
3a. Validasi gagal pada satu atau lebih kriteria (Tabel 3.1).
3b. Sistem mengembalikan pesan kesalahan spesifik sesuai jenis kegagalan.
3c. Pengguna memperbaiki berkas dan mengunggah ulang (kembali ke langkah 2).

---

## 9. Lampiran Formula

```
1. Rata-rata Permintaan Harian (P50) = rata-rata nilai proyeksi P50 selama horizon forecast

2. Estimasi Deviasi Standar Permintaan (σ_demand) 
   = (Proyeksi_P90 − Proyeksi_P50) ÷ 1,2816
   (1,2816 adalah z-score untuk persentil ke-90 pada distribusi normal standar,
    digunakan untuk menurunkan estimasi σ dari lebar interval P50–P90)

3. Demand_LeadTime = Rata-rata Permintaan Harian (P50) × Lead_Time

4. Safety_Stock = 1,65 × σ_demand × √(Lead_Time)
   (1,65 adalah z-score untuk target service level 95%)

5. Reorder_Point (ROP) = ⌈Demand_LeadTime + Safety_Stock⌉

6. Order_Up_To_Level = Rata-rata Permintaan Harian (P50) × (Lead_Time + Review_Period) 
                        + Safety_Stock
   (Review_Period diasumsikan 7 hari untuk MVP — siklus evaluasi mingguan)

7. Reorder_Quantity (ROQ) = ⌈Order_Up_To_Level − current_stock⌉,
   dengan nilai minimum 0 (tidak ada rekomendasi negatif)

8. Hari_Ketersediaan = current_stock ÷ Rata-rata Permintaan Harian (P50)
```

---

*Dokumen SRS ini adalah turunan teknis langsung dari PRD Cakra — Predictive Replenishment v1.0, menjawab seluruh Open Question yang tercatat pada PRD Bagian 14 dengan keputusan desain eksplisit sebagaimana didokumentasikan pada tiap bagian relevan di atas.*
