package main

import (
	"fmt"
	"os"

	"jentik_be/config"
	"jentik_be/routes"
)

func main() {
	config.ConnectDatabase()

	r := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server berjalan di http://localhost:%s\n", port)
	r.Run(":" + port)
}