# JentikBE (GDGOC Jentik Backend)

**JentikBE** adalah backend service untuk aplikasi GDGOC Jentik, sebuah platform pendeteksi dan pelaporan titik rawan jentik nyamuk. Dibangun menggunakan **Go (Golang)** dengan framework **Gin**, ORM **GORM** (PostgreSQL), dan integrasi **Google Generative AI (Gemini)** untuk analisis gambar secara otomatis.

## 🚀 Fitur Utama

- **Autentikasi & Otorisasi:** Sistem login dengan JWT dan *Role-Based Access Control* (User, Kader, Admin).
- **Scan AI:** Deteksi genangan air atau potensi sarang jentik nyamuk pada gambar menggunakan Google Gemini AI.
- **Geolokasi (Heatmap & Jarak):** Pemetaan titik rawan dan pengecekan jarak pengguna terhadap titik bahaya terdekat.
- **Pelaporan & Verifikasi:** Sistem submit laporan oleh user/kader dan verifikasi oleh admin/puskesmas.

## 🛠️ Teknologi yang Digunakan

- **Bahasa Pemrograman:** [Go 1.25+](https://golang.org/)
- **Framework Web:** [Gin Web Framework](https://github.com/gin-gonic/gin)
- **Database:** PostgreSQL (dengan ORM [GORM](https://gorm.io/))
- **AI Integration:** [Google Generative AI SDK](https://github.com/google/generative-ai-go)
- **Keamanan:** JSON Web Token (JWT) & CORS (Gin-CORS)

## 📦 Prasyarat

Sebelum menjalankan proyek ini, pastikan sistem Anda telah memiliki:
- Go (versi 1.25.5 atau yang lebih baru)
- PostgreSQL
- (Opsional) [Air](https://github.com/cosmtrek/air) untuk *live-reloading* saat *development*.

## ⚙️ Cara Menjalankan (Setup & Installation)

1. **Clone repository ini**
   ```bash
   git clone <url-repo-anda>
   cd JentikBE
   ```

2. **Install Dependencies**
   ```bash
   go mod tidy
   ```

3. **Konfigurasi Environment Variables**
   Buat file `.env` di root direktori dengan referensi variabel berikut:
   ```env
   # Server Config
   PORT=8080

   # Database Configuration (PostgreSQL)
   DB_HOST=localhost
   DB_USER=postgres
   DB_PASSWORD=yourpassword
   DB_NAME=jentik_db
   DB_PORT=5432
   DB_SSLMODE=disable
   DB_TIMEZONE=Asia/Jakarta

   # Security & External APIs
   JWT_SECRET=your_super_secret_jwt_key
   GEMINI_API_KEY=your_google_gemini_api_key
   ```

4. **Jalankan Aplikasi**
   Anda bisa menjalankan aplikasi menggunakan perintah bawaan Go:
   ```bash
   go run main.go
   ```
   Atau menggunakan **Air** untuk *live-reloading*:
   ```bash
   air
   ```

Aplikasi akan berjalan secara default di `http://localhost:8080`.

## 📚 Dokumentasi API

Untuk daftar lengkap endpoint API, format request, response, dan role yang dibutuhkan, silakan merujuk pada file **[API_DOCS.md](./API_DOCS.md)** yang tersedia di direktori ini.

## 📁 Struktur Direktori

- `config/`: Konfigurasi *database* dan fungsi inisialisasi.
- `controllers/`: *Handler* utama (*logic*) untuk tiap *endpoint* (Autentikasi, Scan, Laporan).
- `middlewares/`: Penanganan autentikasi (JWT) dan *Role-based Middleware*.
- `models/`: Definisi struktur *struct* Go dan *schema database* GORM.
- `routes/`: Konfigurasi *routing* API Gin.
- `utils/`: *Helper function* seperti generator JWT dan pemanggil AI (Gemini).
- `uploads/`: Folder penyimpanan untuk *file* media statis atau gambar yang di-*upload*.
