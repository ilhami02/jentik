package routes

import (
	"net/http"

	"jentik_be/controllers"
	"jentik_be/middlewares"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Static("/uploads", "./uploads")

	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Welcome to GDGOC Jentik API!"})
	})

	v1 := r.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}

		user := v1.Group("/user", middlewares.AuthMiddleware())
		{
			user.GET("/heatmap", controllers.GetHeatmap)
			user.POST("/scan", controllers.ScanImage)
			user.GET("/check-distance", controllers.CheckDistance)
			user.PUT("/location", controllers.UpdateLocation)
			user.POST("/reports", func(c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"status": "success", "message": "Laporan berhasil dikirim, menunggu verifikasi."})
			})
		}

		kader := v1.Group("/kader", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("kader"))
		{
			kader.POST("/reports", controllers.KaderSubmitReport)
			kader.GET("/history", controllers.KaderGetHistory)
			kader.GET("/blank-spots", controllers.KaderGetBlankSpots)
			kader.POST("/emergency", controllers.KaderReportEmergency)
		}

		admin := v1.Group("/admin", middlewares.AuthMiddleware(), middlewares.RoleMiddleware("admin"))
		{
			admin.GET("/reports/pending", controllers.GetPendingReports)
			admin.PUT("/reports/:id/verify", controllers.VerifyReport)
			admin.POST("/interventions", controllers.CreateIntervention)
		}
	}

	return r
}