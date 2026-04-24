package controllers

import "time"

type HeatMapResponse struct {
	ID    uint    `json:"id"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Level string  `json:"level"`
}

type DeteksiJentik struct {
	IsRawan bool   `json:"is_rawan"`
	Alasan  string `json:"alasan"`
	Saran   string `json:"saran"`
}

type UpdateLocationRequest struct {
	Lat float64 `json:"lat" binding:"required"`
	Lng float64 `json:"lng" binding:"required"`
}

type SubmitReportRequest struct {
	Lat       float64 `form:"lat" binding:"required"`
	Lng       float64 `form:"lng" binding:"required"`
	Deskripsi string  `form:"deskripsi"`
}

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
