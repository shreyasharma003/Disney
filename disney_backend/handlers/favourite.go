package handlers

import (
	"disney/database"
	"disney/models"
	"disney/workers"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddFavouriteRequest represents the request payload to add a favourite
type AddFavouriteRequest struct {
	CartoonID uint `json:"cartoon_id" binding:"required"`
}

// FavouriteWorkerPoolInstance is the global instance of the favourite worker pool
// Initialized in main.go and used by handlers
var FavouriteWorkerPoolInstance *workers.FavouriteWorkerPool

// AddFavourite adds a cartoon to user's favourites using the worker pool (User only)
// This handler queues the add job and returns immediately without waiting for database write
// The actual favourite add happens asynchronously in background workers
func AddFavourite(c *gin.Context) {
	userID := c.GetUint("userID")

	var req AddFavouriteRequest
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

	// Enqueue favourite add job to worker pool for async processing
	// This returns immediately without blocking the HTTP request
	// If the favourite already exists, the worker will handle it gracefully
	FavouriteWorkerPoolInstance.EnqueueFavouriteJob(userID, req.CartoonID, "add")

	// Return immediate response to client
	c.JSON(http.StatusAccepted, gin.H{
		"message":         "Favourite add request queued successfully",
		"user_id":         userID,
		"cartoon_id":      req.CartoonID,
		"queue_length":    FavouriteWorkerPoolInstance.GetQueueLength(),
		"processing_note": "Favourite is being processed asynchronously",
	})
}

// GetUserFavourites retrieves all favourites for the logged-in user
func GetUserFavourites(c *gin.Context) {
	userID := c.GetUint("userID")

	var favourites []models.Favourite
	if err := database.DB.Where("user_id = ?", userID).Preload("Cartoon").Find(&favourites).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch favourites",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Favourites retrieved successfully",
		"data":    favourites,
	})
}

// RemoveFavourite removes a cartoon from user's favourites using the worker pool
// This handler queues the remove job and returns immediately without waiting for database write
// The actual favourite removal happens asynchronously in background workers
func RemoveFavourite(c *gin.Context) {
	userID := c.GetUint("userID")
	cartoonID := c.Param("cartoon_id")

	// Quick validation - check if favourite exists for idempotency
	var favourite models.Favourite
	if result := database.DB.Where("user_id = ? AND cartoon_id = ?", userID, cartoonID).First(&favourite); result.RowsAffected == 0 {
		// Favourite doesn't exist - could be already removed or never existed
		// For idempotency, we return success anyway
		c.JSON(http.StatusOK, gin.H{
			"message": "Favourite removal request processed (already removed or never existed)",
		})
		return
	}

	// Extract numeric ID from favourite for queue
	// Enqueue favourite remove job to worker pool for async processing
	FavouriteWorkerPoolInstance.EnqueueFavouriteJob(userID, favourite.CartoonID, "remove")

	// Return immediate response to client
	c.JSON(http.StatusAccepted, gin.H{
		"message":         "Favourite remove request queued successfully",
		"user_id":         userID,
		"cartoon_id":      favourite.CartoonID,
		"queue_length":    FavouriteWorkerPoolInstance.GetQueueLength(),
		"processing_note": "Favourite removal is being processed asynchronously",
	})
}
