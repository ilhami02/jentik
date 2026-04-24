package controllers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"jentik_be/config"
	"jentik_be/models"
	"jentik_be/utils"

	"github.com/gin-gonic/gin"
)

func GetHeatmap(c *gin.Context) {
	var heatmapData []HeatMapResponse
	err := config.DB.Table("reports").
	Select("id, ST_Y(lokasi::geometry) as lat, ST_X(lokasi::geometry) as lng, tingkat_bahaya as level").
	Where("status = ?", models.StatusAccepted).
	Scan(&heatmapData).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Gagal mengambil data peta"})
		return
	}

	if heatmapData == nil {
		heatmapData = []HeatMapResponse{}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Data heatmap berhasil diambil", "data": heatmapData})
}

func ScanImage(c *gin.Context) {
	latStr := c.PostForm("lat")
	lngStr := c.PostForm("lng")

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Koordinat Latitude dan Longitude wajib dikirim"})
		return
	}
	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gambar tidak ditemukan"})
		return
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung."})
		return
	}

	os.MkdirAll("uploads", os.ModePerm)
	fileName := filepath.Base(fileHeader.Filename)
	imageURL := "/uploads/" + fileName

	if err := c.SaveUploadedFile(fileHeader, "uploads/"+fileName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan gambar"})
		return
	}

	file, _ := fileHeader.Open()
	defer file.Close()

	aiResponse, err := utils.AnalyzeImageWithGemini(file, fileHeader.Size, mimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI Error: " + err.Error()})
		return
	}

	cleanString := strings.ReplaceAll(aiResponse, "```json", "")
	cleanString = strings.ReplaceAll(cleanString, "```", "")
	cleanString = strings.TrimSpace(cleanString)

	var hasilDeteksi DeteksiJentik
	if err := json.Unmarshal([]byte(cleanString), &hasilDeteksi); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Format AI tidak valid"})
		return
	}

	pesanTambahan := " Lingkungan terdeteksi aman."
	if hasilDeteksi.IsRawan {
		userIDFloat, _ := c.Get("user_id")
		userID := uint(userIDFloat.(float64))

		// Jika rawan dari AI, tingkat bahaya adalah "rawan"
		query := `INSERT INTO reports (user_id, jenis_laporan, image_url, deskripsi, tingkat_bahaya, status, lokasi, created_at, updated_at) VALUES (?, 'jentik', ?, '', 'rawan', 'pending', ST_SetSRID(ST_MakePoint(?, ?), 4326), NOW(), NOW())`
		if err := config.DB.Exec(query, userID, imageURL, lng, lat).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan laporan"})
			return
		}
		pesanTambahan = " Gambar terindikasi rawan dan telah dibuatkan laporan."
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Analisis selesai." + pesanTambahan, "data": hasilDeteksi})
}

func PublicScanImage(c *gin.Context) {
	latStr := c.PostForm("lat")
	lngStr := c.PostForm("lng")

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Koordinat Latitude dan Longitude wajib dikirim"})
		return
	}
	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gambar tidak ditemukan"})
		return
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file tidak didukung."})
		return
	}

	os.MkdirAll("uploads", os.ModePerm)
	fileName := filepath.Base(fileHeader.Filename)
	imageURL := "/uploads/" + fileName

	if err := c.SaveUploadedFile(fileHeader, "uploads/"+fileName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan gambar"})
		return
	}

	file, _ := fileHeader.Open()
	defer file.Close()

	aiResponse, err := utils.AnalyzeImageWithGemini(file, fileHeader.Size, mimeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "AI Error: " + err.Error()})
		return
	}

	cleanString := strings.ReplaceAll(aiResponse, "```json", "")
	cleanString = strings.ReplaceAll(cleanString, "```", "")
	cleanString = strings.TrimSpace(cleanString)

	var hasilDeteksi DeteksiJentik
	if err := json.Unmarshal([]byte(cleanString), &hasilDeteksi); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Format AI tidak valid"})
		return
	}

	pesanTambahan := " Lingkungan terdeteksi aman."
	if hasilDeteksi.IsRawan {
		// Jika rawan dari AI, tingkat bahaya adalah "rawan"
		query := `INSERT INTO reports (user_id, jenis_laporan, image_url, deskripsi, tingkat_bahaya, status, lokasi, created_at, updated_at) VALUES (NULL, 'jentik', ?, '', 'rawan', 'pending', ST_SetSRID(ST_MakePoint(?, ?), 4326), NOW(), NOW())`
		if err := config.DB.Exec(query, imageURL, lng, lat).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan laporan"})
			return
		}
		pesanTambahan = " Gambar terindikasi rawan dan telah dibuatkan laporan."
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Analisis selesai." + pesanTambahan, "data": hasilDeteksi})
}

func CheckDistance(c *gin.Context) {
	userIDFloat, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID tidak valid"})
		return
	}
	userID := uint(userIDFloat.(float64))

	var jarakTerdekat *float64

	query := `
		SELECT MIN(ST_Distance(u.lokasi::geography, r.lokasi::geography))
		FROM users u, reports r
		WHERE u.id = ? 
		  AND r.status = 'accepted' 
		  AND u.lokasi IS NOT NULL
	`

	if err := config.DB.Raw(query, userID).Scan(&jarakTerdekat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung jarak: " + err.Error()})
		return
	}

	if jarakTerdekat == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":      "success",
			"jarak_meter": 0,
			"kategori":    "aman",
			"message":     "Tidak dapat menghitung jarak. Pastikan lokasi rumah Anda sudah diatur, atau saat ini belum ada titik rawan yang dilaporkan.",
		})
		return
	}

	dist := *jarakTerdekat
	var kategori, message string

	if dist <= 50 {
		kategori = "bahaya"
		message = "Waspada! Rumah Anda berada sangat dekat dari titik rawan jentik aktif."
	} else if dist <= 100 {
		kategori = "warning"
		message = "Perhatian! Rumah Anda berada di radius peringatan dari titik rawan."
	} else {
		kategori = "aman"
		message = "Lingkungan rumah Anda terdeteksi aman (cukup jauh) dari titik laporan jentik terdekat."
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"jarak_meter": dist,
		"kategori":    kategori,
		"message":     message,
	})
}

func PublicCheckDistance(c *gin.Context) {
	latStr := c.Query("lat")
	lngStr := c.Query("lng")

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Koordinat Latitude dan Longitude wajib dikirim sebagai query parameter"})
		return
	}

	lat, errLat := strconv.ParseFloat(latStr, 64)
	lng, errLng := strconv.ParseFloat(lngStr, 64)
	if errLat != nil || errLng != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format koordinat tidak valid. Gunakan angka desimal."})
		return
	}

	var jarakTerdekat *float64

	query := `
		SELECT MIN(ST_Distance(
			ST_SetSRID(ST_MakePoint(?, ?), 4326)::geography,
			r.lokasi::geography
		))
		FROM reports r
		WHERE r.status = 'accepted'
	`

	if err := config.DB.Raw(query, lng, lat).Scan(&jarakTerdekat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung jarak: " + err.Error()})
		return
	}

	if jarakTerdekat == nil {
		c.JSON(http.StatusOK, gin.H{
			"status":      "success",
			"jarak_meter": 0,
			"kategori":    "aman",
			"message":     "Saat ini belum ada titik rawan yang dilaporkan.",
		})
		return
	}

	dist := *jarakTerdekat
	var kategori, message string

	if dist <= 50 {
		kategori = "bahaya"
		message = "Waspada! Lokasi Anda berada sangat dekat dari titik rawan jentik aktif."
	} else if dist <= 100 {
		kategori = "warning"
		message = "Perhatian! Lokasi Anda berada di radius peringatan dari titik rawan."
	} else {
		kategori = "aman"
		message = "Lokasi Anda terdeteksi aman (cukup jauh) dari titik laporan jentik terdekat."
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      "success",
		"jarak_meter": dist,
		"kategori":    kategori,
		"message":     message,
	})
}


func UpdateLocation(c *gin.Context) {
	userIDFloat, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID tidak valid"})
		return
	}
	userID := uint(userIDFloat.(float64))

	var req UpdateLocationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid. Pastikan mengirim JSON {lat, lng} dalam bentuk angka."})
		return
	}

	query := `
		UPDATE users 
		SET lokasi = ST_SetSRID(ST_MakePoint(?, ?), 4326), updated_at = NOW() 
		WHERE id = ?
	`
	
	if err := config.DB.Exec(query, req.Lng, req.Lat, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui lokasi rumah: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Lokasi rumah berhasil diperbarui!",
	})
}

func UserSubmitReport(c *gin.Context) {
	userIDFloat, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID tidak valid"})
		return
	}
	userID := uint(userIDFloat.(float64))

	var req SubmitReportRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid. Pastikan mengirim lat, lng, dan deskripsi."})
		return
	}

	// Validasi koordinat - pastikan tidak NULL
	// Note: koordinat 0,0 adalah valid (Null Island), jadi kami tidak cek nilai == 0

	// Handle file gambar
	fileHeader, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Gambar wajib dikirim"})
		return
	}

	mimeType := fileHeader.Header.Get("Content-Type")
	if !strings.HasPrefix(mimeType, "image/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format file harus berupa gambar (jpg, png, dll)"})
		return
	}

	// Simpan file dengan nama unik
	os.MkdirAll("uploads", os.ModePerm)
	fileName := "report_" + strconv.Itoa(int(userID)) + "_" + strconv.FormatInt(time.Now().Unix(), 10) + filepath.Ext(fileHeader.Filename)
	imageURL := "/uploads/" + fileName

	if err := c.SaveUploadedFile(fileHeader, "uploads/"+fileName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan gambar"})
		return
	}

	// Insert laporan ke database
	query := `
		INSERT INTO reports (user_id, jenis_laporan, image_url, deskripsi, tingkat_bahaya, status, lokasi, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ST_SetSRID(ST_MakePoint(?, ?), 4326), NOW(), NOW())
	`

	if err := config.DB.Exec(query, userID, "jentik", imageURL, req.Deskripsi, req.TingkatBahaya, "pending", req.Lng, req.Lat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan laporan: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Laporan jentik berhasil dikirim! Admin akan memverifikasi dalam waktu singkat.",
	})
}