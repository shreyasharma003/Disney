package handlers

import (
	"disney/database"
	"disney/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddFavouriteRequest represents the request payload to add a favourite
type AddFavouriteRequest struct {
	CartoonID uint `json:"cartoon_id" binding:"required"`
}

// AddFavourite adds a cartoon to user's favourites (User only)
func AddFavourite(c *gin.Context) {
	userID := c.GetUint("userID")

	var req AddFavouriteRequest
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

	// Check if already in favourites (unique constraint)
	var existingFav models.Favourite
	if result := database.DB.Where("user_id = ? AND cartoon_id = ?", userID, req.CartoonID).First(&existingFav); result.RowsAffected > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Cartoon already in favourites",
		})
		return
	}

	// Create new favourite entry
	newFavourite := models.Favourite{
		UserID:    userID,
		CartoonID: req.CartoonID,
	}

	if err := database.DB.Create(&newFavourite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to add favourite",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Cartoon added to favourites successfully",
		"data": gin.H{
			"id":         newFavourite.ID,
			"user_id":    newFavourite.UserID,
			"cartoon_id": newFavourite.CartoonID,
		},
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

// RemoveFavourite removes a cartoon from user's favourites
func RemoveFavourite(c *gin.Context) {
	userID := c.GetUint("userID")
	cartoonID := c.Param("cartoon_id")

	// Check if favourite exists
	var favourite models.Favourite
	if result := database.DB.Where("user_id = ? AND cartoon_id = ?", userID, cartoonID).First(&favourite); result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Favourite not found",
		})
		return
	}

	// Delete favourite
	if err := database.DB.Delete(&favourite).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to remove favourite",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cartoon removed from favourites successfully",
	})
}
