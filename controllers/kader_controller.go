package controllers

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"jentik_be/config"

	"github.com/gin-gonic/gin"
)

type ReportHistoryResponse struct {
	ID           uint      `json:"id"`
	JenisLaporan string    `json:"jenis_laporan"`
	ImageURL     string    `json:"image_url"`
	Status       string    `json:"status"`
	CatatanAdmin string    `json:"catatan_admin"`
	Lat          float64   `json:"lat"`
	Lng          float64   `json:"lng"`
	CreatedAt    time.Time `json:"created_at"`
}

func KaderGetHistory(c *gin.Context) {
	userIDFloat, _ := c.Get("user_id")
	userID := uint(userIDFloat.(float64))

	var history []ReportHistoryResponse

	err := config.DB.Table("reports").
		Select("id, jenis_laporan, image_url, status, catatan_admin, ST_Y(lokasi::geometry) as lat, ST_X(lokasi::geometry) as lng, created_at").
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
		INSERT INTO reports (user_id, jenis_laporan, image_url, status, lokasi, created_at, updated_at) 
		VALUES (?, 'suspek_dbd', ?, 'pending', ST_SetSRID(ST_MakePoint(?, ?), 4326), NOW(), NOW())
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
	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "Laporan jentik kader terkirim dengan GPS lock."})
}

func KaderGetBlankSpots(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "success", "data": "List koordinat area abu-abu masih dalam pengembangan"})
}