package routes

import (
	"disney/handlers"
	"disney/middleware"

	"github.com/gin-gonic/gin"
)

const (
	cartoonsByIDPath = "/cartoons/:id"
)

// SetupAdminRoutes sets up all admin routes
func SetupAdminRoutes(router *gin.RouterGroup) {
	// Routes accessible by all authenticated users (both users and admins)
	authenticated := router.Group("")
	authenticated.Use(middleware.AuthRequired())
	{
		// Get all cartoon names
		authenticated.GET("/cartoons/names", handlers.GetAllCartoonNames)

		// Get top 5 trending cartoons by IMDb rating
		authenticated.GET("/cartoons/trending", handlers.GetTrendingCartoons)

		// Get specific cartoon by ID
		authenticated.GET(cartoonsByIDPath, handlers.GetCartoonByID)

		// Get cartoons by filters
		authenticated.GET("/cartoons/by-character", handlers.GetCartoonsByCharacter)
		authenticated.GET("/cartoons/by-genre", handlers.GetCartoonsByGenre)
		authenticated.GET("/cartoons/by-year", handlers.GetCartoonsByYear)
		authenticated.GET("/cartoons/by-age-group", handlers.GetCartoonsByAgeGroup)

		// Get recently viewed cartoons
		authenticated.GET("/recently-viewed", handlers.GetRecentlyViewed)
	}

	// Routes accessible only by admins
	admin := router.Group("")
	admin.Use(middleware.AuthRequired(), middleware.AdminOnly())
	{
		// Create new cartoon with characters
		admin.POST("/cartoons", handlers.CreateCartoon)

		// Update cartoon by ID
		admin.PUT(cartoonsByIDPath, handlers.UpdateCartoon)

		// Delete cartoon by ID or title
		admin.DELETE("/cartoons", handlers.DeleteCartoon)

		// Character management
		admin.POST("/characters", handlers.CreateCharacter)
		admin.GET("/characters/cartoon/:cartoon_id", handlers.GetCharactersByCartoon)
		admin.PUT("/characters/:id", handlers.UpdateCharacter)
		admin.DELETE("/characters/:id", handlers.DeleteCharacter)

		// Admin logs management
		admin.POST("/logs", handlers.CreateAdminLog)
		admin.GET("/logs", handlers.GetAdminLogs)
		admin.GET("/logs/stats", handlers.GetAdminLogStats)

		// Request logs management
		admin.GET("/request-logs", handlers.GetRequestLogs)
		admin.GET("/request-logs/stats", handlers.GetRequestLogStats)
	}
}
