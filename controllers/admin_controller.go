package controllers

import (
	"net/http"
	"time"

	"jentik_be/config"
	"jentik_be/models"

	"github.com/gin-gonic/gin"
)

type PendingReportResponse struct {
	ID        uint      `json:"id"`
	ImageURL  string    `json:"image_url"`
	Lat       float64   `json:"lat"`
	Lng       float64   `json:"lng"`
	CreatedAt time.Time `json:"created_at"`
}

type VerifyRequest struct {
	Status  string `json:"status" binding:"required"`
	Catatan string `json:"catatan"`
}

func GetPendingReports(c *gin.Context) {
	var pendingReports []PendingReportResponse

	err := config.DB.Table("reports").
		Select("id, image_url, ST_Y(lokasi::geometry) as lat, ST_X(lokasi::geometry) as lng, created_at").
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

// Fitur Intervensi (Dummy sementara)
func CreateIntervention(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "Tindakan dicatat."})
}