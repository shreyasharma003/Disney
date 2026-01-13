package handlers

import (
	"disney/database"
	"disney/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CreateCharacter creates a new character
func CreateCharacter(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		ImageURL  string `json:"image_url" binding:"required"`
		CartoonID uint   `json:"cartoon_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
		return
	}

	// Verify cartoon exists
	var cartoon models.Cartoon
	if err := database.DB.First(&cartoon, req.CartoonID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Cartoon not found"})
		return
	}

	// Create character
	character := models.Character{
		Name:      req.Name,
		ImageURL:  req.ImageURL,
		CartoonID: req.CartoonID,
	}

	if err := database.DB.Create(&character).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to create character"})
		return
	}

	// Log admin action
	if adminID, exists := c.Get("userID"); exists {
		adminLog := models.AdminLog{
			AdminID: adminID.(uint),
			Action:  "CREATE",
			Entity:  "Character: " + character.Name,
		}
		database.DB.Create(&adminLog)
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":         character.ID,
		"name":       character.Name,
		"image_url":  character.ImageURL,
		"cartoon_id": character.CartoonID,
		"message":    "Character created successfully",
	})
}

// GetCharactersByCartoon gets all characters for a cartoon
func GetCharactersByCartoon(c *gin.Context) {
	cartoonID := c.Param("cartoon_id")

	var characters []models.Character
	if err := database.DB.Where("cartoon_id = ?", cartoonID).Find(&characters).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch characters"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"characters": characters,
		"count":      len(characters),
	})
}

// UpdateCharacter updates a character
func UpdateCharacter(c *gin.Context) {
	characterID := c.Param("id")

	var req struct {
		Name     string `json:"name" binding:"required"`
		ImageURL string `json:"image_url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
		return
	}

	var character models.Character
	if err := database.DB.First(&character, characterID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Character not found"})
		return
	}

	character.Name = req.Name
	character.ImageURL = req.ImageURL

	if err := database.DB.Save(&character).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to update character"})
		return
	}

	// Log admin action
	if adminID, exists := c.Get("userID"); exists {
		adminLog := models.AdminLog{
			AdminID: adminID.(uint),
			Action:  "UPDATE",
			Entity:  "Character: " + character.Name,
		}
		database.DB.Create(&adminLog)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Character updated successfully",
		"character": character,
	})
}

// DeleteCharacter deletes a character
func DeleteCharacter(c *gin.Context) {
	characterID := c.Param("id")

	// First, get the character name for logging
	var character models.Character
	if err := database.DB.First(&character, characterID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "Character not found"})
		return
	}

	if err := database.DB.Delete(&models.Character{}, characterID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to delete character"})
		return
	}

	// Log admin action
	if adminID, exists := c.Get("userID"); exists {
		adminLog := models.AdminLog{
			AdminID: adminID.(uint),
			Action:  "DELETE",
			Entity:  "Character: " + character.Name,
		}
		database.DB.Create(&adminLog)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Character deleted successfully"})
}
