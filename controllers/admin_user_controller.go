package controllers

import (
	"net/http"
	"time"

	"jentik_be/config"
	"jentik_be/models"
	"jentik_be/utils"

	"github.com/gin-gonic/gin"
)

// ========== Types ==========

type UserResponse struct {
	ID        uint      `json:"id"`
	Nama      string    `json:"nama"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required,oneof=user kader admin"`
}

type AdminCreateUserRequest struct {
	Nama     string `json:"nama" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=user kader admin"`
}

// ========== Handlers ==========

// AdminGetUsers mengambil daftar semua user. Bisa difilter berdasarkan role via query parameter.
func AdminGetUsers(c *gin.Context) {
	var users []UserResponse
	roleFilter := c.Query("role")

	query := config.DB.Model(&models.User{}).Select("id, nama, email, role, created_at, updated_at")

	if roleFilter != "" {
		query = query.Where("role = ?", roleFilter)
	}

	if err := query.Order("created_at DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengambil data pengguna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Data pengguna berhasil diambil",
		"total":   len(users),
		"data":    users,
	})
}

// AdminGetUserByID mengambil detail satu user berdasarkan ID.
func AdminGetUserByID(c *gin.Context) {
	userID := c.Param("id")

	var user UserResponse
	err := config.DB.Model(&models.User{}).
		Select("id, nama, email, role, created_at, updated_at").
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengguna tidak ditemukan"})
		return
	}

	// Hitung jumlah laporan milik user ini
	var totalReports int64
	config.DB.Model(&models.Report{}).Where("user_id = ?", userID).Count(&totalReports)

	c.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   user,
		"stats": gin.H{
			"total_reports": totalReports,
		},
	})
}

// AdminCreateUser membuat user baru oleh admin.
func AdminCreateUser(c *gin.Context) {
	var req AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format request tidak valid. Pastikan nama, email, password, dan role terisi dengan benar."})
		return
	}

	// Cek apakah email sudah dipakai
	var existingCount int64
	config.DB.Model(&models.User{}).Where("email = ?", req.Email).Count(&existingCount)
	if existingCount > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Email sudah terdaftar"})
		return
	}

	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengenkripsi password"})
		return
	}

	newUser := models.User{
		Nama:     req.Nama,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     models.Role(req.Role),
	}

	if err := config.DB.Omit("Lokasi").Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal membuat pengguna: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"status":  "success",
		"message": "Pengguna berhasil dibuat",
		"data": gin.H{
			"id":   newUser.ID,
			"nama": newUser.Nama,
			"email": newUser.Email,
			"role": newUser.Role,
		},
	})
}

// AdminUpdateUserRole mengubah role user (user <-> kader). Admin tidak bisa diubah.
func AdminUpdateUserRole(c *gin.Context) {
	userID := c.Param("id")

	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role harus salah satu dari: user, kader, admin"})
		return
	}

	// Pastikan user target ada
	var targetUser models.User
	if err := config.DB.Where("id = ?", userID).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengguna tidak ditemukan"})
		return
	}

	// Cegah admin mengubah role dirinya sendiri
	adminIDFloat, _ := c.Get("user_id")
	adminID := uint(adminIDFloat.(float64))
	if targetUser.ID == adminID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak dapat mengubah role akun Anda sendiri"})
		return
	}

	if err := config.DB.Model(&models.User{}).Where("id = ?", userID).Update("role", req.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengubah role pengguna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Role pengguna berhasil diubah menjadi " + req.Role,
	})
}

// AdminDeleteUser menghapus (soft delete) user berdasarkan ID.
func AdminDeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Pastikan user target ada
	var targetUser models.User
	if err := config.DB.Where("id = ?", userID).First(&targetUser).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pengguna tidak ditemukan"})
		return
	}

	// Cegah admin menghapus dirinya sendiri
	adminIDFloat, _ := c.Get("user_id")
	adminID := uint(adminIDFloat.(float64))
	if targetUser.ID == adminID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak dapat menghapus akun Anda sendiri"})
		return
	}

	// Cegah menghapus sesama admin
	if targetUser.Role == models.RoleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Tidak dapat menghapus akun admin lain"})
		return
	}

	if err := config.DB.Delete(&targetUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghapus pengguna"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Pengguna berhasil dihapus",
	})
}
