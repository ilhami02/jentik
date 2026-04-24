package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"jentik_be/config"

	"github.com/gin-gonic/gin"
)

func KaderGetHistory(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))

	var history []ReportHistoryResponse

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
		history = []ReportHistoryResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   history,
	})
}

func KaderReportEmergency(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))

	latStr := c.PostForm("lat")
	lngStr := c.PostForm("lng")

	if latStr == "" || lngStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Koordinat Latitude dan Longitude wajib dikirim untuk laporan darurat"})
		return
	}
	lat, _ := strconv.ParseFloat(latStr, 64)
	lng, _ := strconv.ParseFloat(lngStr, 64)

	imageURL := ""
	fileHeader, err := c.FormFile("image")
	if err == nil {
		os.MkdirAll("uploads", os.ModePerm)
		fileName := "darurat_" + strconv.FormatInt(time.Now().Unix(), 10) + filepath.Ext(fileHeader.Filename)
		imageURL = "/uploads/" + fileName
		c.SaveUploadedFile(fileHeader, "uploads/"+fileName)
	}

	query := `
		INSERT INTO reports (user_id, jenis_laporan, image_url, tingkat_bahaya, status, lokasi, created_at, updated_at) 
		VALUES (?, 'suspek_dbd', ?, 'rawan', 'pending', ST_SetSRID(ST_MakePoint(?, ?), 4326), NOW(), NOW())
	`
	if err := config.DB.Exec(query, userID, imageURL, lng, lat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengirim laporan darurat: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Peringatan darurat suspek DBD telah berhasil dikirim ke Puskesmas!",
	})
}

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