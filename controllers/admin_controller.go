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

func AdminGetDashboard(c *gin.Context) {
	var result DashboardResponse

	// Total warga terdaftar (role = user)
	config.DB.Model(&models.User{}).
		Where("role = ?", models.RoleUser).
		Count(&result.TotalWarga)

	// Kader aktif (role = kader)
	config.DB.Model(&models.User{}).
		Where("role = ?", models.RoleKader).
		Count(&result.KaderAktif)

	// Laporan pending
	config.DB.Model(&models.Report{}).
		Where("status = ?", models.StatusPending).
		Count(&result.LaporanPending)

	// Notifikasi darurat (laporan suspek DBD yang masih pending)
	config.DB.Model(&models.Report{}).
		Where("jenis_laporan = ? AND status = ?", "suspek_dbd", models.StatusPending).
		Count(&result.NotifikasiDarurat)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data dashboard berhasil diambil",
		"data":    result,
	})
}

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
		// Auto-assign report ke district berdasarkan lokasi
		assignQuery := `
			UPDATE reports 
			SET district_id = d.id
			FROM districts d
			WHERE reports.id = ?
			  AND ST_Contains(d.geometry, reports.lokasi::geometry)
		`
		config.DB.Exec(assignQuery, reportID)

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

// ========== DISTRICT MANAGEMENT ==========

// GetDistrictSummary mengambil summary semua districts dengan count reports per tingkat bahaya
func GetDistrictSummary(c *gin.Context) {
	var result DistrictCountResponse

	// Query count reports dengan status accepted per tingkat bahaya (tanpa filter district)
	config.DB.Table("reports").
		Where("status = ?", models.StatusAccepted).
		Select(
			"COUNT(CASE WHEN tingkat_bahaya = 'rawan' THEN 1 END) as rawan,"+
			"COUNT(CASE WHEN tingkat_bahaya = 'warning' THEN 1 END) as waspada,"+
			"COUNT(CASE WHEN tingkat_bahaya = 'aman' THEN 1 END) as aman",
		).
		Row().
		Scan(&result.Rawan, &result.Waspada, &result.Aman)

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Summary area jentik berhasil diambil",
		"data":    result,
	})
}

// CreateDistrict membuat wilayah/kecamatan baru dan auto-assign reports berdasarkan spatial query
func CreateDistrict(c *gin.Context) {
	var req CreateDistrictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid. Pastikan nama wilayah dan coordinates terisi. Coordinates format: [[lat, lng], [lat, lng], ...]"})
		return
	}

	// Validasi coordinates
	if len(req.Coordinates) < 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Polygon minimal harus 4 titik (first point harus sama dengan last point)"})
		return
	}

	// Cek apakah titik awal dan akhir sama (polygon ring harus tertutup)
	firstCoord := req.Coordinates[0]
	lastCoord := req.Coordinates[len(req.Coordinates)-1]
	if firstCoord[0] != lastCoord[0] || firstCoord[1] != lastCoord[1] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Polygon ring harus tertutup: titik pertama harus sama dengan titik terakhir"})
		return
	}

	// Cek apakah nama district sudah ada
	var existingCount int64
	config.DB.Model(&models.District{}).Where("nama = ?", req.Nama).Count(&existingCount)
	if existingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Wilayah dengan nama tersebut sudah ada"})
		return
	}

	// Build WKT polygon: POLYGON((lng lat, lng lat, ...))
	// PostGIS menggunakan format (lng, lat) bukan (lat, lng)
	var polygonCoords string
	for i, coord := range req.Coordinates {
		if i > 0 {
			polygonCoords += ", "
		}
		polygonCoords += fmt.Sprintf("%f %f", coord[1], coord[0]) // lng lat
	}
	geometryWKT := fmt.Sprintf("SRID=4326;POLYGON((%s))", polygonCoords)

	// Create district dengan geometry
	newDistrict := models.District{
		Nama:     req.Nama,
		Geometry: geometryWKT,
	}

	if err := config.DB.Create(&newDistrict).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat wilayah: " + err.Error()})
		return
	}

	// Auto-assign reports yang berada dalam polygon ke district ini
	// Update reports yang koordinatnya berada dalam polygon -> set district_id
	updateQuery := `
		UPDATE reports 
		SET district_id = ?
		WHERE status = 'accepted' 
		  AND ST_Contains(
			(SELECT geometry FROM districts WHERE id = ?),
			lokasi::geometry
		  )
	`
	
	if err := config.DB.Exec(updateQuery, newDistrict.ID, newDistrict.ID).Error; err != nil {
		// Log error tapi jangan return error, district sudah created
		fmt.Println("Warning: Gagal auto-assign reports ke district:", err)
	}

	// Count berapa reports yang ter-assign
	var assignedCount int64
	config.DB.Table("reports").
		Where("district_id = ?", newDistrict.ID).
		Count(&assignedCount)

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": fmt.Sprintf("Wilayah berhasil ditambahkan dan %d laporan ter-assign otomatis", assignedCount),
		"data": gin.H{
			"id":              newDistrict.ID,
			"nama":            newDistrict.Nama,
			"assigned_count":  assignedCount,
		},
	})
}

// GetAllDistricts mengambil semua districts dengan detail count reports per tingkat bahaya
func GetAllDistricts(c *gin.Context) {
	var districts []models.District
	
	if err := config.DB.Find(&districts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data wilayah"})
		return
	}

	var results []DistrictDetailResponse
	
	for _, district := range districts {
		var rawan, waspada, aman int64
		
		config.DB.Table("reports").
			Where("district_id = ? AND status = ?", district.ID, models.StatusAccepted).
			Where("tingkat_bahaya = ?", models.TingkatRawan).
			Count(&rawan)

		config.DB.Table("reports").
			Where("district_id = ? AND status = ?", district.ID, models.StatusAccepted).
			Where("tingkat_bahaya = ?", models.TingkatWarning).
			Count(&waspada)

		config.DB.Table("reports").
			Where("district_id = ? AND status = ?", district.ID, models.StatusAccepted).
			Where("tingkat_bahaya = ?", models.TingkatAman).
			Count(&aman)

		results = append(results, DistrictDetailResponse{
			ID:      district.ID,
			Nama:    district.Nama,
			Rawan:   rawan,
			Waspada: waspada,
			Aman:    aman,
		})
	}

	if results == nil {
		results = []DistrictDetailResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data wilayah berhasil diambil",
		"total":   len(results),
		"data":    results,
	})
}

// SearchDistricts mencari districts berdasarkan nama
func SearchDistricts(c *gin.Context) {
	keyword := c.Query("q")
	
	if keyword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Parameter pencarian 'q' tidak boleh kosong"})
		return
	}

	var districts []models.District
	
	if err := config.DB.Where("LOWER(nama) LIKE LOWER(?)", "%"+keyword+"%").
		Find(&districts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mencari wilayah"})
		return
	}

	var results []DistrictDetailResponse
	
	for _, district := range districts {
		var rawan, waspada, aman int64
		
		config.DB.Table("reports").
			Where("district_id = ? AND status = ?", district.ID, models.StatusAccepted).
			Where("tingkat_bahaya = ?", models.TingkatRawan).
			Count(&rawan)

		config.DB.Table("reports").
			Where("district_id = ? AND status = ?", district.ID, models.StatusAccepted).
			Where("tingkat_bahaya = ?", models.TingkatWarning).
			Count(&waspada)

		config.DB.Table("reports").
			Where("district_id = ? AND status = ?", district.ID, models.StatusAccepted).
			Where("tingkat_bahaya = ?", models.TingkatAman).
			Count(&aman)

		results = append(results, DistrictDetailResponse{
			ID:      district.ID,
			Nama:    district.Nama,
			Rawan:   rawan,
			Waspada: waspada,
			Aman:    aman,
		})
	}

	if results == nil {
		results = []DistrictDetailResponse{}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Hasil pencarian wilayah",
		"total":   len(results),
		"data":    results,
	})
}

// DeleteDistrict menghapus district
func DeleteDistrict(c *gin.Context) {
	districtID := c.Param("id")

	var district models.District
	if err := config.DB.Where("id = ?", districtID).First(&district).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Wilayah tidak ditemukan"})
		return
	}

	if err := config.DB.Delete(&district).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus wilayah"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Wilayah berhasil dihapus",
	})
}

// ReassignReportsToDistricts re-assign semua reports ke districts berdasarkan spatial query
// Useful jika ada perubahan geometry atau data reports baru
func ReassignReportsToDistricts(c *gin.Context) {
	// Clear semua district_id terlebih dahulu
	if err := config.DB.Table("reports").
		Update("district_id", nil).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal clear district assignment: " + err.Error()})
		return
	}

	// Get semua districts
	var districts []models.District
	if err := config.DB.Find(&districts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data wilayah"})
		return
	}

	totalAssigned := int64(0)

	// Assign reports ke setiap district
	for _, district := range districts {
		query := `
			UPDATE reports 
			SET district_id = ?
			WHERE status = 'accepted' 
			  AND district_id IS NULL
			  AND ST_Contains(
				(SELECT geometry FROM districts WHERE id = ?),
				lokasi::geometry
			  )
		`
		
		result := config.DB.Exec(query, district.ID, district.ID)
		if result.Error != nil {
			fmt.Println("Warning: Gagal assign reports untuk district", district.Nama, ":", result.Error)
		} else {
			totalAssigned += result.RowsAffected
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"message":          fmt.Sprintf("Re-assignment selesai. %d laporan ter-assign ke districts", totalAssigned),
		"assigned_count":   totalAssigned,
	})
}

// GetReportsUnassignedCount mengambil jumlah reports yang belum ter-assign ke district
func GetReportsUnassignedCount(c *gin.Context) {
	var count int64
	
	config.DB.Table("reports").
		Where("district_id IS NULL AND status = ?", models.StatusAccepted).
		Count(&count)

	c.JSON(http.StatusOK, gin.H{
		"status":           "success",
		"unassigned_count": count,
		"message":          fmt.Sprintf("%d laporan belum ter-assign ke wilayah", count),
	})
}