package main

import (
	"fmt"
	"disney/database"
	"disney/handlers"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	database.InitDB()

	// Create Gin router
	router := gin.Default()

	// Public Auth routes (no authentication required)
	auth := router.Group("/api/auth")
	{
		auth.POST("/signup", handlers.Signup)
		auth.POST("/login", handlers.Login)
		auth.POST("/create-admin", handlers.CreateAdmin)
	}

	port := ":8080"
	fmt.Println("Server running on port", port)
	router.Run(port)

}