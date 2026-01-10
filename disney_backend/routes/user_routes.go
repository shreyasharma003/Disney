package routes

import (
	"disney/handlers"
	"disney/middleware"

	"github.com/gin-gonic/gin"
)

// UserRoutes defines all user-specific routes
func UserRoutes(router *gin.Engine) {
	// Protected User routes (authentication required, not admin)
	user := router.Group("/api/user")
	user.Use(middleware.AuthRequired())
	{
		// Favourites endpoints
		user.POST("/favourites", handlers.AddFavourite)
		user.GET("/favourites", handlers.GetUserFavourites)
		user.DELETE("/favourites/:cartoon_id", handlers.RemoveFavourite)

		// Rating endpoints
		user.POST("/ratings", handlers.AddRating)
		user.GET("/ratings/:cartoon_id", handlers.GetUserRating)
		user.PUT("/ratings/:cartoon_id", handlers.UpdateRating)

		// View/Tracking endpoints
		user.POST("/views", handlers.RecordView)
		user.GET("/cartoons/:cartoon_id/views", handlers.GetCartoonViewCount)
	}
}
