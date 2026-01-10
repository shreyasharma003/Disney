package handlers

import (
	"disney/database"
	"disney/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CharacterUpdateRequest represents a character in the cartoon update
type CharacterUpdateRequest struct {
	Name     string `json:"name" binding:"required"`
	ImageURL string `json:"image_url"`
}

// UpdateCartoonRequest represents the request to update a cartoon
type UpdateCartoonRequest struct {
	Title       *string                   `json:"title"`
	Description *string                   `json:"description"`
	PosterURL   *string                   `json:"poster_url"`
	ReleaseYear *int                      `json:"release_year"`
	GenreID     *uint                     `json:"genre_id"`
	AgeGroupID  *uint                     `json:"age_group_id"`
	IsFeatured  *bool                     `json:"is_featured"`
	Characters  []CharacterUpdateRequest  `json:"characters"`
}

// UpdateCartoon updates an existing cartoon
func UpdateCartoon(c *gin.Context) {
	cartoonID := c.Param("id")
	if cartoonID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Cartoon ID is required",
			"error":   "Please provide cartoon ID in URL",
		})
		return
	}

	var req UpdateCartoonRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	// Find existing cartoon
	var cartoon models.Cartoon
	if err := database.DB.First(&cartoon, cartoonID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Cartoon not found",
			"error":   err.Error(),
		})
		return
	}

	// Validate foreign keys
	if err := validateUpdateRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"error":   "Validation failed",
		})
		return
	}

	// Start transaction
	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Update cartoon and characters
	if err := performCartoonUpdate(tx, &cartoon, req); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update cartoon",
			"error":   err.Error(),
		})
		return
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to commit transaction",
			"error":   err.Error(),
		})
		return
	}

	// Load relationships for response
	database.DB.Preload("Genre").Preload("AgeGroup").First(&cartoon, cartoon.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Cartoon updated successfully",
		"data":    cartoon,
	})
}

// validateUpdateRequest validates genre and age group IDs
func validateUpdateRequest(req UpdateCartoonRequest) error {
	if req.GenreID != nil {
		var genre models.Genre
		if err := database.DB.First(&genre, *req.GenreID).Error; err != nil {
			return err
		}
	}

	if req.AgeGroupID != nil {
		var ageGroup models.AgeGroup
		if err := database.DB.First(&ageGroup, *req.AgeGroupID).Error; err != nil {
			return err
		}
	}

	return nil
}

// performCartoonUpdate updates cartoon fields and characters in transaction
func performCartoonUpdate(tx *gorm.DB, cartoon *models.Cartoon, req UpdateCartoonRequest) error {
	// Update cartoon fields
	updateCartoonFields(cartoon, req)

	// Save cartoon updates
	if err := tx.Save(cartoon).Error; err != nil {
		return err
	}

	// Update characters if provided
	if req.Characters != nil {
		if err := updateCharacters(tx, cartoon.ID, req.Characters); err != nil {
			return err
		}
	}

	return nil
}

// updateCartoonFields updates individual cartoon fields if provided
func updateCartoonFields(cartoon *models.Cartoon, req UpdateCartoonRequest) {
	if req.Title != nil {
		cartoon.Title = *req.Title
	}
	if req.Description != nil {
		cartoon.Description = *req.Description
	}
	if req.PosterURL != nil {
		cartoon.PosterURL = *req.PosterURL
	}
	if req.ReleaseYear != nil {
		cartoon.ReleaseYear = *req.ReleaseYear
	}
	if req.GenreID != nil {
		cartoon.GenreID = *req.GenreID
	}
	if req.AgeGroupID != nil {
		cartoon.AgeGroupID = *req.AgeGroupID
	}
	if req.IsFeatured != nil {
		cartoon.IsFeatured = *req.IsFeatured
	}
}

// updateCharacters deletes old characters and creates new ones
func updateCharacters(tx *gorm.DB, cartoonID uint, characters []CharacterUpdateRequest) error {
	// Delete existing characters
	if err := tx.Where("cartoon_id = ?", cartoonID).Delete(&models.Character{}).Error; err != nil {
		return err
	}

	// Create new characters
	for _, charReq := range characters {
		character := models.Character{
			Name:      charReq.Name,
			ImageURL:  charReq.ImageURL,
			CartoonID: cartoonID,
		}
		if err := tx.Create(&character).Error; err != nil {
			return err
		}
	}

	return nil
}