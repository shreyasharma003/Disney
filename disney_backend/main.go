package main

import (
	"disney/config"
	"disney/database"
	"disney/handlers"
	"disney/routes"
	"disney/services"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	
	// Initialize database
	database.InitDB()
	
	// Initialize Redis
	config.InitRedis()
	defer config.CloseRedis()
	
	// Set Redis client in services
	services.SetRedisClient(config.RedisClient)

	// Create Gin router
	router := gin.Default()

	// Public Auth routes (no authentication required)
	auth := router.Group("/api/auth")
	{
		auth.POST("/signup", handlers.Signup)
		auth.POST("/login", handlers.Login)
		auth.POST("/create-admin", handlers.CreateAdmin)
	}

	// Admin routes (protected)
	adminGroup := router.Group("/api/admin")
	routes.SetupAdminRoutes(adminGroup)

	port := ":8080"
	fmt.Println("Server running on port", port)
	router.Run(port)

}