package routes

import (
	"net/http"

	"jentik_be/controllers"
	"jentik_be/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// r.Use(cors.Default())

	r.Static("/uploads", "./uploads")

	// r.Use(func(c *gin.Context) {
	// 	c.Writer.Header().Set("Access-Control-Allow-Origin", "https://gdgoc.skyibe.my.id")
	// 	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	// 	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// 	if c.Request.Method == "OPTIONS" {
	// 		c.AbortWithStatus(204)
	// 		return
	// 	}
	// 	c.Next()
	// })

	r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"https://jentikmap.eyi.my.id"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

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

		v1.POST("/scan", controllers.PublicScanImage)
		v1.GET("/heatmap", controllers.GetHeatmap)
		v1.GET("/check-distance", controllers.PublicCheckDistance)

		user := v1.Group("/user", middlewares.AuthMiddleware())
		{
			user.POST("/scan", controllers.ScanImage)
			user.GET("/check-distance", controllers.CheckDistance)
			user.PUT("/location", controllers.UpdateLocation)
			user.POST("/reports", controllers.UserSubmitReport)
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