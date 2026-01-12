package main

import (
	"disney/config"
	"disney/database"
	"disney/handlers"
	"disney/routes"
	"disney/services"
	"disney/workers"
	"fmt"

	"github.com/gin-contrib/cors"
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

	// Initialize and start view worker pool
	// 5 concurrent workers, buffer size of 100 jobs
	viewWorkerPool := workers.NewViewWorkerPool(5, 100)
	viewWorkerPool.Start()
	handlers.ViewWorkerPoolInstance = viewWorkerPool

	// Initialize and start favourite worker pool
	// 5 concurrent workers, buffer size of 100 jobs
	favouriteWorkerPool := workers.NewFavouriteWorkerPool(5, 100)
	favouriteWorkerPool.Start()
	handlers.FavouriteWorkerPoolInstance = favouriteWorkerPool

	// Create Gin router
	router := gin.Default()

	// Configure CORS - Allow all origins for development
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Authorization", "Accept"}
	corsConfig.ExposeHeaders = []string{"Content-Length"}
	corsConfig.AllowCredentials = false
	router.Use(cors.New(corsConfig))

	// Public Auth routes (no authentication required)
	auth := router.Group("/api/auth")
	{
		auth.POST("/signup", handlers.Signup)
		auth.POST("/login", handlers.Login)
		auth.POST("/create-admin", handlers.CreateAdmin)
	}

	// User routes with middleware
	routes.UserRoutes(router)
	// Admin routes (protected)
	adminGroup := router.Group("/api/admin")
	routes.SetupAdminRoutes(adminGroup)

	port := ":8080"
	fmt.Println("Server running on port", port)
	fmt.Println("View worker pool: 5 workers, buffer: 100")
	fmt.Println("Favourite worker pool: 5 workers, buffer: 100")
	router.Run(port)

}