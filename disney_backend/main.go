package main

import (
	"disney/database"
	"disney/handlers"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	database.InitDB()

	// Create Gin router
	router := gin.Default()

	// Auth routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/signup", handlers.Signup)
		auth.POST("/login", handlers.Login)
	}

	port := ":8080"
	fmt.Println("Server running on port", port)
	router.Run(port)
}
