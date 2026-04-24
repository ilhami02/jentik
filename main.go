package main

import (
	"fmt"
	"os"

	"jentik_be/config"
	"jentik_be/routes"
	"github.com/gin-contrib/cors"
)

func main() {
	config.ConnectDatabase()

	r.Use(cors.Default())

	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server berjalan di http://localhost:%s\n", port)
	r.Run(":" + port)
}