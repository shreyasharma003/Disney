package handlers

import (
	"disney/database"
	"disney/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetAllCartoonNames returns all cartoon names
func GetAllCartoonNames(c *gin.Context) {
	var cartoons []models.Cartoon

	// Query only ID and Title fields
	if err := database.DB.Select("id", "title").Find(&cartoons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch cartoons",
			"error":   err.Error(),
		})
		return
	}

	// Extract names into a simple array
	var cartoonNames []map[string]interface{}
	for _, cartoon := range cartoons {
		cartoonNames = append(cartoonNames, map[string]interface{}{
			"id":    cartoon.ID,
			"title": cartoon.Title,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cartoons fetched successfully",
		"data":    cartoonNames,
		"count":   len(cartoonNames),
	})
}
