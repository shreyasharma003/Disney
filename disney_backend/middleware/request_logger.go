package middleware

import (
	"disney/database"
	"disney/models"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger middleware logs all API requests to the database
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate response time
		responseTime := int(time.Since(startTime).Milliseconds())

		// Get user ID if authenticated
		var userID *uint
		if userIDInterface, exists := c.Get("userID"); exists {
			if uid, ok := userIDInterface.(uint); ok {
				userID = &uid
			}
		}

		// Create request log entry
		requestLog := models.RequestLog{
			UserID:       userID,
			Endpoint:     c.Request.URL.Path,
			Method:       c.Request.Method,
			StatusCode:   c.Writer.Status(),
			ResponseTime: responseTime,
		}

		// Save to database (async to not block response)
		go func() {
			if err := database.DB.Create(&requestLog).Error; err != nil {
				// Log error but don't fail the request
				// You can implement proper logging here
			}
		}()
	}
}
