package controllers

import (
	"net/http"
	"fmt"
	"time"

	"jentik_be/config"
	"jentik_be/models"
	"jentik_be/utils"

	"github.com/gin-gonic/gin"
)

func GetPendingReports(c *gin.Context) {
	var pendingReports []PendingReportResponse

	err := config.DB.Table("reports").
		Select("id, image_url, tingkat_bahaya, ST_Y(lokasi::geometry) as lat, ST_X(lokasi::geometry) as lng, created_at").
		Where("status = ?", models.StatusPending).
		Order("created_at DESC").
		Scan(&pendingReports).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data laporan pending"})
		return
	}

	if pendingReports == nil {
		pendingReports = []PendingReportResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data laporan pending berhasil diambil",
		"data":    pendingReports,
	})
}

func VerifyReport(c *gin.Context) {
	reportID := c.Param("id")

	var req VerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid. Pastikan mengirim JSON {status, catatan}"})
		return
	}

	if req.Status != string(models.StatusAccepted) && req.Status != string(models.StatusRejected) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status hanya boleh 'accepted' atau 'rejected'"})
		return
	}

	err := config.DB.Model(&models.Report{}).
		Where("id = ?", reportID).
		Updates(map[string]interface{}{
			"status":        req.Status,
			"catatan_admin": req.Catatan,
		}).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memverifikasi laporan"})
		return
	}

	pesan := "Laporan berhasil ditolak."
	if req.Status == string(models.StatusAccepted) {
		pesan = "Laporan berhasil diterima dan sekarang muncul di HeatMap!"
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": pesan})
}

func StreamNotifications(c *gin.Context) {
	// Set header HTTP khusus untuk Server-Sent Events (SSE)
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

	// Buat channel khusus untuk Admin yang baru saja hit endpoint ini
	clientChan := make(chan string)

	// Daftarkan Admin ini ke dalam list
	utils.Mu.Lock()
	utils.AdminClients[clientChan] = true
	utils.Mu.Unlock()

	defer func() {
		utils.Mu.Lock()
		delete(utils.AdminClients, clientChan)
		close(clientChan)
		utils.Mu.Unlock()
	}()

	// Looping tanpa henti untuk mendengarkan channel
	clientGone := c.Writer.CloseNotify()
	for {
		select {
		case <-clientGone:
			return 
		case msg := <-clientChan:
			c.SSEvent("emergency", msg)
			c.Writer.Flush()
		}
	}
}

func CreateIntervention(c *gin.Context) {
	adminIDFloat, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Akses ditolak. User ID tidak ditemukan."})
		return
	}
	adminID := uint(adminIDFloat.(float64))

	var req Intervention
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid"})
		return
	}

	lokasi := fmt.Sprintf("SRID=4326;POINT(%f %f)", req.Lng, req.Lat) // Format PostGIS
	intervention := models.Intervention{
		AdminID:       adminID,
		JenisTindakan: req.JenisTindakan,
		Lokasi:        lokasi,
		RadiusArea:    req.RadiusArea,
		Tanggal:       time.Now(),
	}

	if err := config.DB.Create(&intervention).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencatat intervensi"})
		return
	}

	if err := config.DB.Model(&models.Report{}).
		Where("id = ?", req.ReportID).
		Update("status", models.StatusResolved).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Tindakan dicatat, tapi gagal mengupdate status laporan"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success", 
		"message": "Tindakan " + req.JenisTindakan + " berhasil dicatat dan laporan telah ditangani.",
	})
}