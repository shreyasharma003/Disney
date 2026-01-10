package handlers

import (
	"disney/database"
	"disney/models"
	"disney/workers"
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Check if cartoon exists (validation query)
	var cartoon models.Cartoon
	if result := database.DB.First(&cartoon, req.CartoonID); result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cartoon not found",
		})
		return
	}

	// Enqueue view job to worker pool for async processing
	// This returns immediately without blocking the HTTP request
	ViewWorkerPoolInstance.EnqueueViewJob(userID, req.CartoonID)

	// Get current view count for response (might not include the just-enqueued view)
	var viewCount int64
	database.DB.Model(&models.View{}).Where("cartoon_id = ?", req.CartoonID).Count(&viewCount)

	// Return immediate response to client
	// Note: viewCount may not include the enqueued view yet due to async processing
	c.JSON(http.StatusAccepted, gin.H{
		"message":         "View recording queued successfully",
		"cartoon_id":      req.CartoonID,
		"current_views":   viewCount,
		"queue_length":    ViewWorkerPoolInstance.GetQueueLength(),
		"processing_note": "View is being processed asynchronously",
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
