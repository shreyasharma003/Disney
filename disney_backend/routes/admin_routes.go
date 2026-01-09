package routes

import (
	"disney/handlers"
	"disney/middleware"

	"github.com/gin-gonic/gin"
)

// SetupAdminRoutes sets up all admin routes
func SetupAdminRoutes(router *gin.RouterGroup) {
	// Apply admin authentication middleware
	admin := router.Group("")
	admin.Use(middleware.AuthRequired(), middleware.AdminOnly())
	{
		// Get all cartoon names
		admin.GET("/cartoons/names", handlers.GetAllCartoonNames)
	}
}
