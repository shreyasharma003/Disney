package handlers

import (
	"disney/database"
	"disney/models"
	"disney/services"
	"disney/workers"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RecordViewRequest represents the request payload to record a view
type RecordViewRequest struct {
	CartoonID uint `json:"cartoon_id" binding:"required"`
}

// ViewWorkerPoolInstance is the global instance of the view worker pool
// Initialized in main.go and used by handlers
var ViewWorkerPoolInstance *workers.ViewWorkerPool

// RecordView records a view for a cartoon using the worker pool (authenticated users only)
// This handler queues the view job and returns immediately without waiting for database write
// The actual view recording happens asynchronously in background workers
func RecordView(c *gin.Context) {
	userID := c.GetUint("userID")

	var req RecordViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("ERROR: Invalid request body for view recording: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Printf("INFO: Received view recording request - user_id=%d, cartoon_id=%d", userID, req.CartoonID)

	// Check if cartoon exists (validation query)
	var cartoon models.Cartoon
	if result := database.DB.First(&cartoon, req.CartoonID); result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cartoon not found",
		})
		return
	}

	// Update recently viewed in Redis IMMEDIATELY (synchronous)
	// This ensures the UI updates instantly when user clicks a cartoon
	log.Printf("Recording view: user_id=%d, cartoon_id=%d", userID, req.CartoonID)
	if err := services.AddRecentlyViewed(int(userID), int(req.CartoonID)); err != nil {
		// Log error and return warning message
		log.Printf("WARNING: Failed to add to recently viewed (Redis may not be running): %v", err)
		// Continue with database view recording even if Redis fails
	} else {
		log.Printf("SUCCESS: Added cartoon %d to recently viewed for user %d", req.CartoonID, userID)
	}

	// Enqueue view job to worker pool for async processing (database write)
	// This returns immediately without blocking the HTTP request
	ViewWorkerPoolInstance.EnqueueViewJob(userID, req.CartoonID)

	// Return immediate response to client
	c.JSON(http.StatusAccepted, gin.H{
		"message":    "View recorded successfully",
		"cartoon_id": req.CartoonID,
	})
}

// GetCartoonViewCount retrieves total view count for a specific cartoon
func GetCartoonViewCount(c *gin.Context) {
	cartoonID := c.Param("cartoon_id")

	// Check if cartoon exists
	var cartoon models.Cartoon
	if result := database.DB.First(&cartoon, cartoonID); result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cartoon not found",
		})
		return
	}

	// Get view count
	var viewCount int64
	database.DB.Model(&models.View{}).Where("cartoon_id = ?", cartoonID).Count(&viewCount)

	c.JSON(http.StatusOK, gin.H{
		"message":     "View count retrieved successfully",
		"cartoon_id":  cartoonID,
		"total_views": viewCount,
	})
}
