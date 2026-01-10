package handlers

import (
	"disney/database"
	"disney/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UpdateCartoonRequest represents the request to update a cartoon
type UpdateCartoonRequest struct {
	Title       *string                   `json:"title"`
	Description *string                   `json:"description"`
	PosterURL   *string                   `json:"poster_url"`
	ReleaseYear *int                      `json:"release_year"`
	GenreID     *uint                     `json:"genre_id"`
	AgeGroupID  *uint                     `json:"age_group_id"`
	IsFeatured  *bool                     `json:"is_featured"`
	Characters  []CreateCharacterRequest  `json:"characters"`
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

	// Verify genre if provided
	if req.GenreID != nil {
		var genre models.Genre
		if err := database.DB.First(&genre, *req.GenreID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid genre ID",
				"error":   "Genre not found",
			})
			return
		}
	}

	// Verify age group if provided
	if req.AgeGroupID != nil {
		var ageGroup models.AgeGroup
		if err := database.DB.First(&ageGroup, *req.AgeGroupID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "Invalid age group ID",
				"error":   "Age group not found",
			})
			return
		}
	}

	// Start transaction
	tx := database.DB.Begin()

	// Update cartoon fields if provided
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

	// Save cartoon updates
	if err := tx.Save(&cartoon).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to update cartoon",
			"error":   err.Error(),
		})
		return
	}

	// Update characters if provided
	if req.Characters != nil {
		// Delete existing characters
		if err := tx.Where("cartoon_id = ?", cartoon.ID).Delete(&models.Character{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to delete old characters",
				"error":   err.Error(),
			})
			return
		}

		// Create new characters
		for _, charReq := range req.Characters {
			character := models.Character{
				Name:      charReq.Name,
				ImageURL:  charReq.ImageURL,
				CartoonID: cartoon.ID,
			}
			if err := tx.Create(&character).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"message": "Failed to create character",
					"error":   err.Error(),
				})
				return
			}
		}
	}

	// Commit transaction
	tx.Commit()

	// Load relationships for response
	database.DB.Preload("Genre").Preload("AgeGroup").First(&cartoon, cartoon.ID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Cartoon updated successfully",
		"data":    cartoon,
	})
}