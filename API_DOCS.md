# 🦟 GDGOC Jentik API Documentation

**Base URL:** `https://your-server.com` atau `http://localhost:8080`  
**API Version:** v1  
**Prefix:** `/api/v1`

---

## 📋 Daftar Isi

- [Autentikasi](#-autentikasi)
- [Public Endpoints](#-public-endpoints-tanpa-login)
- [User Endpoints](#-user-endpoints-login-required)
- [Kader Endpoints](#-kader-endpoints-role-kader)
- [Admin Endpoints](#-admin-endpoints-role-admin)
- [Error Handling](#-error-handling)
- [Static Files](#-static-files)

---

## 🔐 Autentikasi

Beberapa endpoint membutuhkan **JWT Token** di header:

```
Authorization: Bearer <token>
```

Token didapat dari endpoint `/api/v1/auth/login` dan berlaku selama **72 jam (3 hari)**.

> **💡 Catatan:** Fitur utama (scan, heatmap, check-distance) bisa diakses **tanpa login**. Login hanya dibutuhkan untuk fitur kader, admin, dan beberapa fitur user tambahan.

---

## 1. Auth Endpoints (Public)

### 1.1 Register

Mendaftarkan akun baru.

```
POST /api/v1/auth/register
```

**Content-Type:** `application/json`

**Request Body:**

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `nama` | string | ✅ | Nama lengkap pengguna |
| `email` | string | ✅ | Email valid |
| `password` | string | ✅ | Minimal 6 karakter |
| `role` | string | ❌ | `"user"` (default), `"kader"`, atau `"admin"` |

**Contoh Request:**
```json
{
    "nama": "Ilham Maulana",
    "email": "ilham@example.com",
    "password": "password123",
    "role": "user"
}
```

**Response Sukses (201):**
```json
{
    "message": "Registrasi berhasil",
    "data": {
        "id": 1,
        "nama": "Ilham Maulana",
        "role": "user"
    }
}
```

**Response Error (400):**
```json
{
    "error": "Key: 'RegisterRequest.Email' Error:... (detail validasi)"
}
```

**Response Error (500):**
```json
{
    "error": "Gagal mendaftar: ERROR: duplicate key value violates unique constraint..."
}
```

---

### 1.2 Login

Login dan mendapatkan JWT token.

```
POST /api/v1/auth/login
```

**Content-Type:** `application/json`

**Request Body:**

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `email` | string | ✅ | Email yang terdaftar |
| `password` | string | ✅ | Password akun |

**Contoh Request:**
```json
{
    "email": "ilham@example.com",
    "password": "password123"
}
```

**Response Sukses (200):**
```json
{
    "message": "Login berhasil",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "role": "user"
}
```

**Response Error (401):**
```json
{
    "error": "Email atau password salah"
}
```

> **💡 Penting untuk Frontend:** Simpan `token` di localStorage/cookie. Simpan `role` untuk menentukan navigasi (dashboard user/kader/admin).

---

## 2. 🌍 Public Endpoints (Tanpa Login)

Semua endpoint di bagian ini **tidak memerlukan login atau token**. Bisa langsung dipanggil dari frontend.

### 2.1 Scan Gambar

Scan gambar lingkungan menggunakan AI untuk mendeteksi area rawan jentik. Jika terdeteksi rawan, laporan otomatis dibuat dengan status `pending`.

```
POST /api/v1/scan
```

**Content-Type:** `multipart/form-data`

**Request Body (Form Data):**

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `image` | file | ✅ | File gambar (JPG, PNG, dll.) |
| `sng` | string | ✅ | Longitude lokasi (contoh: `"106.816666"`) |

**Contoh Request (JavaScript Fetch):**
```javascript
const formData = new FormData();
formData.append('image', fileInput.files[0]);
formData.append('lat', '-6.200000');
formData.append('lng', '106.816666');

const response = await fetch('/api/v1/scan', {
    method: 'POST',
    body: formData,
    // Jangan set Content-Type manual, biarkan browser set otomatis
});
const data = await response.json();
```

**Contoh Request (cURL):**
```bash
curl -X POST http://localhost:8080/api/v1/scan \
  -F "image=@foto_lingkungan.jpg" \
  -F "lat=-6.200000" \
  -F "lng=106.816666"
```

**Response — Aman (200):**
```json
{
    "status": "success",
    "message": "Analisis selesai. Lingkungan terdeteksi aman.",
    "data": {
        "is_rawan": false,
        "alasan": "Tidak ditemukan genangan air atau tempat penampungan air terbuka di gambar.",
        "saran": "Tetap jaga kebersihan lingkungan dan pastikan tidak ada genangan air."
    }
}
```

**Response — Rawan (200):**
```json
{
    "status": "success",
    "message": "Analisis selesai. Gambar terindikasi rawan dan telah dibuatkan laporan.",
    "data": {
        "is_rawan": true,
        "alasan": "Terlihat genangan air di ban bekas yang berpotensi menjadi tempat berkembang biak nyamuk Aedes aegypti.",
        "saran": "Segera buang air yang tergenang dan bersihkan barang bekas yang menampung air."
    }
}
```

**Response Error (400):**
```json
{ "error": "Koordinat Latitude dan Longitude wajib dikirim" }
```
```json
{ "error": "Gambar tidak ditemukan" }
```
```json
{ "error": "Format file tidak didukung." }
```

---

### 2.2 Get Heatmap

Mengambil data titik-titik rawan jentik yang sudah diverifikasi untuk ditampilkan di peta.

```
GET /api/v1/heatmap
```

**Tidak ada parameter.**

**Contoh Request (JavaScript Fetch):**
```javascript
const response = await fetch('/api/v1/heatmap');
const data = await response.json();
// data.data berisi array titik-titik untuk heatmap
```

**Response Sukses (200):**
```json
{
    "status": "success",
    "message": "Data heatmap berhasil diambil",
    "data": [
        {
            "id": 1,
            "lat": -6.200000,
            "lng": 106.816666,
            "level": "bahaya"
        },
        {
            "id": 2,
            "lat": -6.210000,
            "lng": 106.820000,
            "level": "bahaya"
        }
    ]
}
```

**Response Kosong (200):**
```json
{
    "status": "success",
    "message": "Data heatmap berhasil diambil",
    "data": []
}
```

> **💡 Tips Frontend:** Gunakan library seperti Leaflet.js atau Google Maps untuk menampilkan titik-titik ini sebagai heatmap layer. Setiap object memiliki `lat`, `lng` yang langsung bisa dipetakan.

---

### 2.3 Check Distance

Mengecek jarak lokasi user ke titik rawan jentik terdekat (dalam meter). Frontend langsung kirim koordinat GPS user.

```
GET /api/v1/check-distance?lat={latitude}&lng={longitude}
```

**Query Parameters:**

| Param | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `lat` | number | ✅ | Latitude lokasi user |
| `lng` | number | ✅ | Longitude lokasi user |

**Contoh Request (JavaScript Fetch):**
```javascript
// Dapatkan lokasi user dari browser
navigator.geolocation.getCurrentPosition(async (position) => {
    const lat = position.coords.latitude;
    const lng = position.coords.longitude;

    const response = await fetch(`/api/v1/check-distance?lat=${lat}&lng=${lng}`);
    const data = await response.json();

    console.log(data.kategori);    // "bahaya" / "warning" / "aman"
    console.log(data.jarak_meter); // jarak dalam meter
});
```

**Contoh Request (cURL):**
```bash
curl "http://localhost:8080/api/v1/check-distance?lat=-6.200000&lng=106.816666"
```

**Response Sukses (200):**
```json
{
    "status": "success",
    "jarak_meter": 35.7,
    "kategori": "bahaya",
    "message": "Waspada! Lokasi Anda berada sangat dekat dari titik rawan jentik aktif."
}
```

**Kategori Jarak:**

| Jarak | Kategori | Warna yang disarankan |
|---|---|---|
| ≤ 50 meter | `"bahaya"` | 🔴 Merah |
| ≤ 100 meter | `"warning"` | 🟡 Kuning |
| > 100 meter | `"aman"` | 🟢 Hijau |

**Response — Belum ada data (200):**
```json
{
    "status": "success",
    "jarak_meter": 0,
    "kategori": "aman",
    "message": "Saat ini belum ada titik rawan yang dilaporkan."
}
```

**Response Error (400):**
```json
{ "error": "Koordinat Latitude dan Longitude wajib dikirim sebagai query parameter" }
```
```json
{ "error": "Format koordinat tidak valid. Gunakan angka desimal." }
```

---

## 3. 👤 User Endpoints (Login Required)

Endpoint tambahan untuk user yang sudah login. Membutuhkan header `Authorization: Bearer <token>`.

### 3.1 Scan Gambar (Authenticated)

Sama seperti public scan, tapi laporan yang dibuat akan **terhubung dengan akun user** (`user_id` terisi).

```
POST /api/v1/user/scan
```

**Header:**
```
Authorization: Bearer <token>
```

**Content-Type:** `multipart/form-data`

**Request Body:** Sama dengan [Public Scan](#21-scan-gambar).

**Response:** Sama dengan [Public Scan](#21-scan-gambar).

> **💡 Perbedaan dengan public scan:** Laporan yang dibuat menyimpan `user_id`, sehingga bisa dilacak siapa yang melaporkan.

---

### 3.2 Check Distance (Authenticated)

Mengecek jarak **lokasi rumah yang tersimpan di akun** ke titik rawan terdekat. Lokasi harus diatur dulu via [Update Location](#33-update-location).

```
GET /api/v1/user/check-distance
```

**Header:**
```
Authorization: Bearer <token>
```

**Tidak ada parameter.** Lokasi diambil otomatis dari data akun.

**Response:** Sama formatnya dengan [Public Check Distance](#23-check-distance).

---

### 3.3 Update Location

Mengatur/memperbarui koordinat lokasi rumah user (tersimpan permanen di akun).

```
PUT /api/v1/user/location
```

**Header:**
```
Authorization: Bearer <token>
```

**Content-Type:** `application/json`

**Request Body:**

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `lat` | number | ✅ | Latitude rumah |
| `lng` | number | ✅ | Longitude rumah |

**Contoh Request:**
```json
{
    "lat": -6.200000,
    "lng": 106.816666
}
```

**Response Sukses (200):**
```json
{
    "status": "success",
    "message": "Lokasi rumah berhasil diperbarui!"
}
```

**Response Error (400):**
```json
{
    "error": "Format request tidak valid. Pastikan mengirim JSON {lat, lng} dalam bentuk angka."
}
```

---

### 3.4 Submit Report (User)

> ⚠️ **Status: Dummy** — Endpoint ini belum diimplementasikan sepenuhnya.

```
POST /api/v1/user/reports
```

**Response (201):**
```json
{
    "status": "success",
    "message": "Laporan berhasil dikirim, menunggu verifikasi."
}
```

---

## 4. 🏥 Kader Endpoints (Role: kader)

Endpoint khusus untuk kader kesehatan. Membutuhkan login dengan akun ber-role `"kader"`.

**Header wajib:**
```
Authorization: Bearer <token_kader>
```

### 4.1 Submit Report (Kader)

> ⚠️ **Status: Dummy** — Endpoint ini belum diimplementasikan sepenuhnya.

```
POST /api/v1/kader/reports
```

**Response (201):**
```json
{
    "status": "success",
    "message": "Laporan jentik kader terkirim dengan GPS lock."
}
```

---

### 4.2 Get Report History

Mengambil riwayat semua laporan milik kader yang sedang login.

```
GET /api/v1/kader/history
```

**Header:**
```
Authorization: Bearer <token_kader>
```

**Response Sukses (200):**
```json
{
    "status": "success",
    "data": [
        {
            "id": 5,
            "jenis_laporan": "suspek_dbd",
            "image_url": "/uploads/darurat_1711590000.jpg",
            "status": "accepted",
            "catatan_admin": "Terverifikasi, sudah dilakukan fogging",
            "lat": -6.200000,
            "lng": 106.816666,
            "created_at": "2026-03-28T08:00:00+07:00"
        },
        {
            "id": 3,
            "jenis_laporan": "jentik",
            "image_url": "/uploads/foto_scan.jpg",
            "status": "pending",
            "catatan_admin": "",
            "lat": -6.210000,
            "lng": 106.820000,
            "created_at": "2026-03-27T14:30:00+07:00"
        }
    ]
}
```

**Field `status` yang mungkin:**

| Status | Keterangan |
|---|---|
| `"pending"` | Menunggu verifikasi admin |
| `"accepted"` | Diterima, muncul di heatmap |
| `"rejected"` | Ditolak oleh admin |
| `"resolved"` | Sudah ditangani |

---

### 4.3 Get Blank Spots

> ⚠️ **Status: Dummy** — Endpoint ini belum diimplementasikan sepenuhnya.

```
GET /api/v1/kader/blank-spots
```

**Response (200):**
```json
{
    "status": "success",
    "data": "List koordinat area abu-abu masih dalam pengembangan"
}
```

---

### 4.4 Report Emergency

Mengirim laporan darurat suspek DBD ke admin/puskesmas.

```
POST /api/v1/kader/emergency
```

**Header:**
```
Authorization: Bearer <token_kader>
```

**Content-Type:** `multipart/form-data`

**Request Body (Form Data):**

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `lat` | string | ✅ | Latitude lokasi |
| `lng` | string | ✅ | Longitude lokasi |
| `image` | file | ❌ | Foto bukti (opsional) |

**Contoh Request (JavaScript Fetch):**
```javascript
const formData = new FormData();
formData.append('lat', '-6.200000');
formData.append('lng', '106.816666');
formData.append('image', fileInput.files[0]); // opsional

const response = await fetch('/api/v1/kader/emergency', {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` },
    body: formData
});
```

**Response Sukses (201):**
```json
{
    "status": "success",
    "message": "Peringatan darurat suspek DBD telah berhasil dikirim ke Puskesmas!"
}
```

**Response Error (400):**
```json
{
    "error": "Koordinat Latitude dan Longitude wajib dikirim untuk laporan darurat"
}
```

---

## 5. 🛡️ Admin Endpoints (Role: admin)

Endpoint khusus untuk admin/petugas puskesmas. Membutuhkan login dengan akun ber-role `"admin"`.

**Header wajib:**
```
Authorization: Bearer <token_admin>
```

### 5.1 Get Pending Reports

Mengambil daftar semua laporan yang menunggu verifikasi.

```
GET /api/v1/admin/reports/pending
```

**Header:**
```
Authorization: Bearer <token_admin>
```

**Response Sukses (200):**
```json
{
    "status": "success",
    "message": "Data laporan pending berhasil diambil",
    "data": [
        {
            "id": 7,
            "image_url": "/uploads/foto_scan.jpg",
            "lat": -6.200000,
            "lng": 106.816666,
            "created_at": "2026-03-28T09:00:00+07:00"
        },
        {
            "id": 6,
            "image_url": "/uploads/darurat_1711590000.jpg",
            "lat": -6.210000,
            "lng": 106.820000,
            "created_at": "2026-03-28T08:30:00+07:00"
        }
    ]
}
```

> **💡 Tips Frontend:** Tampilkan gambar dengan URL: `{BASE_URL}{image_url}`, contoh: `http://localhost:8080/uploads/foto_scan.jpg`

---

### 5.2 Verify Report

Memverifikasi (menerima/menolak) sebuah laporan.

```
PUT /api/v1/admin/reports/:id/verify
```

**Header:**
```
Authorization: Bearer <token_admin>
```

**Content-Type:** `application/json`

**URL Parameter:**

| Param | Keterangan |
|---|---|
| `:id` | ID laporan yang akan diverifikasi |

**Request Body:**

| Field | Tipe | Wajib | Keterangan |
|---|---|---|---|
| `status` | string | ✅ | `"accepted"` atau `"rejected"` |
| `catatan` | string | ❌ | Catatan/alasan dari admin |

**Contoh Request — Accept:**
```json
{
    "status": "accepted",
    "catatan": "Terkonfirmasi ada genangan air di lokasi."
}
```

**Contoh Request — Reject:**
```json
{
    "status": "rejected",
    "catatan": "Gambar tidak jelas, tidak terlihat potensi jentik."
}
```

**Response Sukses — Accepted (200):**
```json
{
    "status": "success",
    "message": "Laporan berhasil diterima dan sekarang muncul di HeatMap!"
}
```

**Response Sukses — Rejected (200):**
```json
{
    "status": "success",
    "message": "Laporan berhasil ditolak."
}
```

**Response Error (400):**
```json
{
    "error": "Status hanya boleh 'accepted' atau 'rejected'"
}
```

---

### 5.3 Create Intervention

> ⚠️ **Status: Dummy** — Endpoint ini belum diimplementasikan sepenuhnya.

```
POST /api/v1/admin/interventions
```

**Response (201):**
```json
{
    "status": "success",
    "message": "Tindakan dicatat."
}
```

---

## ❌ Error Handling

### Format Error

Semua error dikembalikan dalam format:
```json
{
    "error": "Pesan error yang menjelaskan masalah"
}
```

### HTTP Status Code

| Code | Keterangan | Kapan Muncul |
|---|---|---|
| `200` | OK | Request berhasil |
| `201` | Created | Data berhasil dibuat |
| `400` | Bad Request | Validasi gagal / format request salah |
| `401` | Unauthorized | Token tidak ada / tidak valid / expired |
| `403` | Forbidden | Role tidak sesuai |
| `500` | Internal Server Error | Error server (DB, AI, dll.) |

### Error Autentikasi (401)

```json
{ "error": "Akses ditolak. Token tidak ditemukan." }
```
```json
{ "error": "Token tidak valid atau sudah kedaluwarsa." }
```

### Error Otorisasi (403)

```json
{ "error": "Akses ditolak. Endpoint ini khusus untuk kader" }
```
```json
{ "error": "Akses ditolak. Endpoint ini khusus untuk admin" }
```

---

## 📁 Static Files

Gambar yang diupload bisa diakses langsung:

```
GET /uploads/<nama_file>
```

**Contoh:**
```
http://localhost:8080/uploads/foto_lingkungan.jpg
http://localhost:8080/uploads/darurat_1711590000.jpg
```

---

## 📊 Ringkasan Semua Endpoint

### Public (Tanpa Login)

| # | Method | Endpoint | Keterangan | Status |
|---|---|---|---|---|
| 1 | `POST` | `/api/v1/auth/register` | Registrasi akun baru | ✅ Active |
| 2 | `POST` | `/api/v1/auth/login` | Login, dapat JWT token | ✅ Active |
| 3 | `POST` | `/api/v1/scan` | Scan gambar + AI deteksi jentik | ✅ Active |
| 4 | `GET` | `/api/v1/heatmap` | Data titik rawan untuk peta | ✅ Active |
| 5 | `GET` | `/api/v1/check-distance?lat=&lng=` | Cek jarak ke titik rawan | ✅ Active |

### User (Login Required)

| # | Method | Endpoint | Keterangan | Status |
|---|---|---|---|---|
| 6 | `POST` | `/api/v1/user/scan` | Scan gambar (tersimpan dgn user_id) | ✅ Active |
| 7 | `GET` | `/api/v1/user/check-distance` | Cek jarak dari lokasi tersimpan | ✅ Active |
| 8 | `PUT` | `/api/v1/user/location` | Set lokasi rumah ke akun | ✅ Active |
| 9 | `POST` | `/api/v1/user/reports` | Submit laporan user | ⚠️ Dummy |

### Kader (Login + Role kader)

| # | Method | Endpoint | Keterangan | Status |
|---|---|---|---|---|
| 10 | `POST` | `/api/v1/kader/reports` | Submit laporan kader | ⚠️ Dummy |
| 11 | `GET` | `/api/v1/kader/history` | Riwayat laporan kader | ✅ Active |
| 12 | `GET` | `/api/v1/kader/blank-spots` | Area yang belum tercover | ⚠️ Dummy |
| 13 | `POST` | `/api/v1/kader/emergency` | Laporan darurat suspek DBD | ✅ Active |

### Admin (Login + Role admin)

| # | Method | Endpoint | Keterangan | Status |
|---|---|---|---|---|
| 14 | `GET` | `/api/v1/admin/reports/pending` | Lihat laporan pending | ✅ Active |
| 15 | `PUT` | `/api/v1/admin/reports/:id/verify` | Accept/reject laporan | ✅ Active |
| 16 | `POST` | `/api/v1/admin/interventions` | Catat intervensi | ⚠️ Dummy |
