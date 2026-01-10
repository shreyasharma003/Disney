package routes

import (
	"disney/handlers"
	"disney/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes sets up all admin routes
func SetupAdminRoutes(router *gin.RouterGroup) {
	// Routes accessible by all authenticated users (both users and admins)
	authenticated := router.Group("")
	authenticated.Use(middleware.AuthRequired())
	{
		// Get cartoon details with IMDb rating
		authenticated.GET("/cartoons/:id", handlers.GetCartoonDetail)

		// Get top 5 trending cartoons by IMDb rating
		authenticated.GET("/cartoons/trending", handlers.GetTrendingCartoons)

		// Get all cartoon names
		authenticated.GET("/cartoons/names", handlers.GetAllCartoonNames)

		// Get cartoons by filters
		authenticated.GET("/cartoons/by-character", handlers.GetCartoonsByCharacter)
		authenticated.GET("/cartoons/by-genre", handlers.GetCartoonsByGenre)
		authenticated.GET("/cartoons/by-year", handlers.GetCartoonsByYear)
		authenticated.GET("/cartoons/by-age-group", handlers.GetCartoonsByAgeGroup)

		// Get specific cartoon by ID (tracks as recently viewed)
		authenticated.GET("/cartoons/:id", handlers.GetCartoonByID)

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
		admin.PUT("/cartoons/:id", handlers.UpdateCartoon)
		
		// Delete cartoon by ID or title
		admin.DELETE("/cartoons", handlers.DeleteCartoon)
	}
}
