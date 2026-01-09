package handlers

import (
	"disney/database"
	"disney/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RecordViewRequest represents the request payload to record a view
type RecordViewRequest struct {
	CartoonID uint `json:"cartoon_id" binding:"required"`
}

// RecordView records a view for a cartoon (authenticated users only)
func RecordView(c *gin.Context) {
	userID := c.GetUint("userID")

	var req RecordViewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
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

	// Record view
	newView := models.View{
		CartoonID: req.CartoonID,
		UserID:    &userID,
		ViewedAt:  time.Now(),
	}

	if err := database.DB.Create(&newView).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to record view",
		})
		return
	}

	// Get total view count for this cartoon
	var viewCount int64
	database.DB.Model(&models.View{}).Where("cartoon_id = ?", req.CartoonID).Count(&viewCount)

	c.JSON(http.StatusCreated, gin.H{
		"message":     "View recorded successfully",
		"view_id":     newView.ID,
		"cartoon_id":  req.CartoonID,
		"total_views": viewCount,
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
