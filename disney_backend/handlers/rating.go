package handlers

import (
	"disney/database"
	"disney/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddRatingRequest represents the request payload to add a rating
type AddRatingRequest struct {
	CartoonID uint `json:"cartoon_id" binding:"required"`
	Rating    int  `json:"rating" binding:"required,min=1,max=5"`
}

// AddRating adds or updates a rating for a cartoon (User only)
func AddRating(c *gin.Context) {
	userID := c.GetUint("userID")

	var req AddRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Validate rating is between 1-5
	if req.Rating < 1 || req.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Rating must be between 1 and 5",
		})
		return
	}

	// Check if cartoon exists
	var cartoon models.Cartoon
	if result := database.DB.First(&cartoon, req.CartoonID); result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Cartoon not found",
		})
		return
	}

	// Check if user already rated this cartoon
	var existingRating models.Rating
	if result := database.DB.Where("user_id = ? AND cartoon_id = ?", userID, req.CartoonID).First(&existingRating); result.RowsAffected > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "You have already rated this cartoon",
		})
		return
	}

	// Create new rating entry
	newRating := models.Rating{
		UserID:    userID,
		CartoonID: req.CartoonID,
		Rating:    req.Rating,
	}

	if err := database.DB.Create(&newRating).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to submit rating",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Rating submitted successfully",
		"data": gin.H{
			"id":         newRating.ID,
			"user_id":    newRating.UserID,
			"cartoon_id": newRating.CartoonID,
			"rating":     newRating.Rating,
			"created_at": newRating.CreatedAt,
		},
	})
}
