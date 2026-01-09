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
		// Get all cartoon names
		authenticated.GET("/cartoons/names", handlers.GetAllCartoonNames)
	}

	// Routes accessible only by admins
	admin := router.Group("")
	admin.Use(middleware.AuthRequired(), middleware.AdminOnly())
	{
		// Add admin-only routes here
	}
}
