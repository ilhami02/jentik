package models

import (
	"time"
	"gorm.io/gorm"
)

type Role string
const (
	RoleUser  Role = "user"
	RoleKader Role = "kader"
	RoleAdmin Role = "admin"
)

type StatusLaporan string
const (
	StatusPending  StatusLaporan = "pending"
	StatusAccepted StatusLaporan = "accepted" 
	StatusRejected StatusLaporan = "rejected"
	StatusResolved StatusLaporan = "resolved" 
)

type User struct {
	ID        uint           `gorm:"primaryKey"`
	Nama      string         `gorm:"size:100;not null"`
	Email     string         `gorm:"size:100;unique;not null"`
	Password  string         `gorm:"not null"`
	Role      Role           `gorm:"type:varchar(20);default:'user'"`
	Lokasi    string         `gorm:"type:geometry(Point, 4326)" json:"-"` 
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Report struct {
	ID            uint           `gorm:"primaryKey"`
	UserID        uint           `gorm:"not null"`
	JenisLaporan  string         `gorm:"type:varchar(50);not null"`
	ImageURL      string         `gorm:"type:text;not null"`
	CatatanAdmin  string         `gorm:"type:text"`
	Status        StatusLaporan  `gorm:"type:varchar(20);default:'pending'"`
	Lokasi        string         `gorm:"type:geometry(Point, 4326);not null" json:"-"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Intervention struct {
	ID               uint      `gorm:"primaryKey"`
	AdminID          uint      `gorm:"not null"`
	JenisTindakan    string    `gorm:"type:varchar(50);not null"`
	Lokasi           string    `gorm:"type:geometry(Point, 4326);not null" json:"-"`
	RadiusArea       float64   `gorm:"not null"`
	Tanggal          time.Time `gorm:"not null"`
	CreatedAt        time.Time
}