package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"fmt"

	"jentik_be/config"
	"jentik_be/models"
	"jentik_be/utils"

	"github.com/gin-gonic/gin"
)

func KaderGetHistory(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))

	var history []Reports

	err := config.DB.Table("reports").
		Select("id, jenis_laporan, image_url, tingkat_bahaya, status, catatan_admin, ST_Y(lokasi::geometry) as lat, ST_X(lokasi::geometry) as lng, created_at").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(&history).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil riwayat laporan"})
		return
	}

	if history == nil {
		history = []Reports{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   history,
	})
}

// func KaderReportEmergency(c *gin.Context) {
// 	userIDFloat, _ := c.Get("user_id")
// 	userID := uint(userIDFloat.(float64))

// 	latStr := c.PostForm("lat")
// 	lngStr := c.PostForm("lng")

// 	if latStr == "" || lngStr == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Koordinat Latitude dan Longitude wajib dikirim untuk laporan darurat"})
// 		return
// 	}
// 	lat, _ := strconv.ParseFloat(latStr, 64)
// 	lng, _ := strconv.ParseFloat(lngStr, 64)

// 	imageURL := ""
// 	fileHeader, err := c.FormFile("image")
// 	if err == nil {
// 		os.MkdirAll("uploads", os.ModePerm)
// 		fileName := "darurat_" + strconv.FormatInt(time.Now().Unix(), 10) + filepath.Ext(fileHeader.Filename)
// 		imageURL = "/uploads/" + fileName
// 		c.SaveUploadedFile(fileHeader, "uploads/"+fileName)
// 	}

// 	query := `
// 		INSERT INTO reports (user_id, jenis_laporan, image_url, tingkat_bahaya, status, lokasi, created_at, updated_at) 
// 		VALUES (?, 'suspek_dbd', ?, 'rawan', 'pending', ST_SetSRID(ST_MakePoint(?, ?), 4326), NOW(), NOW())
// 	`
// 	if err := config.DB.Exec(query, userID, imageURL, lng, lat).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim laporan darurat: " + err.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusCreated, gin.H{
// 		"status":  "success",
// 		"message": "Peringatan darurat suspek DBD telah berhasil dikirim ke Puskesmas!",
// 	})
// }

func KaderSubmitReport(c *gin.Context) {
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
	fileName := "kader_" + strconv.Itoa(int(userID)) + "_" + strconv.FormatInt(time.Now().Unix(), 10) + filepath.Ext(fileHeader.Filename)
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
		"message": "Laporan jentik dari kader berhasil dikirim dengan GPS lock! Admin akan memverifikasi dalam waktu singkat.",
	})
}

func KaderGetBlankSpots(c *gin.Context) {
	var blankSpots []BlankSpotResponse

	err := config.DB.Table("reports").
		Select("id, ST_Y(lokasi::geometry) as lat, ST_X(lokasi::geometry) as lng, tingkat_bahaya").
		Where("status = ?", "accepted").
		Order("created_at DESC").
		Scan(&blankSpots).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data blank spots"})
		return
	}

	// Assign warna berdasarkan tingkat bahaya
	// for i := range blankSpots {
	// 	switch blankSpots[i].TingkatBahaya {
	// 	case "aman":
	// 		blankSpots[i].Color = "hijau"
	// 	case "warning":
	// 		blankSpots[i].Color = "kuning"
	// 	case "rawan":
	// 		blankSpots[i].Color = "merah"
	// 	default:
	// 		blankSpots[i].Color = "abu-abu"
	// 	}
	// }

	if blankSpots == nil {
		blankSpots = []BlankSpotResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data lokasi area dengan tingkat bahaya berhasil diambil",
		"data":    blankSpots,
	})
}

// func KaderReportEmergency(c *gin.Context) {
// 	userID, exists := c.Get("user_id")
// 	if !exists {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID tidak valid"})
// 		return
// 	}

// 	lat := c.PostForm("lat")
// 	lng := c.PostForm("lng")

// 	if (lat == "" || lng == "") {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Koordinat Latitude dan Longitude wajib dikirim untuk laporan darurat"})
// 		return
// 	}

// 	// lat, _ := strconv.ParseFloat(lat, 64)
// 	// lng, _ := strconv.ParseFloat(lng, 64)

// 	var imagePath string
// 	file, err:= c.FormFile("image")
// 	if err == nil {
// 		filename := "darurat_" + strconv.FormatInt(time.Now().Unix(), 10) + filepath.Ext(file.Filename)
// 		savepath := "uploads/" + filename
// 		if err := c.SaveUploadedFile(file, savepath); err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan gambar darurat"})
// 			return
// 		}
// 		imagePath = "/uploads/" + filename
// 	}

// 	report := models.Report{
// 		UserID:       userID.(uint),
// 		Lat:          strconv.ParseFloat(lat, 64),
// 		Lng:          strconv.ParseFloat(lng, 64),
// 		ImageURL:     imagePath,
// 		JenisLaporan: "suspek_dbd", // <-- Ini kuncinya
// 		Status:       "pending",
// 	}

// 	if err := config.DB.Create(&report).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan laporan darurat"})
// 		return
// 	}

// 	pesanDarurat := fmt.Sprintf(`{"message": "SUSPEK DBD BARU!", "lat": %f, "lng": %f}`, report.Lat, report.Lng)
// 	utils.BroadcastEmergency(pesanDarurat)

// 	c.JSON(http.StatusCreated, gin.H{
// 		"status":  "success",
// 		"message": "Peringatan darurat suspek DBD telah berhasil dikirim ke Puskesmas!",
// 		"data":    report,
// 	})
// }
func KaderReportEmergency(c *gin.Context) {
	// 1. Perbaiki konversi UserID dari float64 (bawaan JWT) ke uint, lalu siapkan pointernya
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID tidak valid"})
		return
	}
	userID := uint(userIDInterface.(float64)) // Konversi ke uint

	latStr := c.PostForm("lat")
	lngStr := c.PostForm("lng")

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Koordinat Latitude dan Longitude wajib dikirim untuk laporan darurat"})
		return
	}

	// 2. Parse string ke float64 di luar struct agar tidak error multiple-value
	latFloat, _ := strconv.ParseFloat(latStr, 64)
	lngFloat, _ := strconv.ParseFloat(lngStr, 64)

	var imagePath string
	file, err := c.FormFile("image")
	if err == nil {
		filename := "darurat_" + strconv.FormatInt(time.Now().Unix(), 10) + filepath.Ext(file.Filename)
		savepath := "uploads/" + filename
		if err := c.SaveUploadedFile(file, savepath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan gambar darurat"})
			return
		}
		imagePath = "/" + savepath
	}

	// 3. Ubah Lat dan Lng menjadi format PostGIS Point
	lokasiPostGIS := fmt.Sprintf("SRID=4326;POINT(%f %f)", lngFloat, latFloat)

	// 4. Masukkan ke dalam model. Gunakan pointer &userID dan field Lokasi
	report := models.Report{
		UserID:        &userID,
		Lokasi:        lokasiPostGIS,
		ImageURL:      imagePath,
		JenisLaporan:  "suspek_dbd",
		TingkatBahaya: models.TingkatRawan, // Darurat otomatis rawan
		Status:        models.StatusPending,
	}

	if err := config.DB.Create(&report).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyimpan laporan darurat: " + err.Error()})
		return
	}

	// 5. Gunakan variabel latFloat dan lngFloat untuk broadcast, bukan report.Lat
	pesanDarurat := fmt.Sprintf(`{"message": "SUSPEK DBD BARU!", "lat": %f, "lng": %f}`, latFloat, lngFloat)
	utils.BroadcastEmergency(pesanDarurat)

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Peringatan darurat suspek DBD telah berhasil dikirim ke Puskesmas!",
		"data":    report,
	})
}