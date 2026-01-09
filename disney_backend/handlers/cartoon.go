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

// GetCartoonsByCharacter returns cartoons filtered by character name
func GetCartoonsByCharacter(c *gin.Context) {
	characterName := c.Query("name")
	if characterName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Character name is required",
			"error":   "Please provide 'name' query parameter",
		})
		return
	}

	var cartoons []models.Cartoon
	if err := database.DB.Joins("JOIN characters ON characters.cartoon_id = cartoons.id").
		Where("characters.name ILIKE ?", "%"+characterName+"%").
		Preload("Genre").Preload("AgeGroup").
		Find(&cartoons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch cartoons",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cartoons fetched successfully",
		"data":    cartoons,
		"count":   len(cartoons),
	})
}

// GetCartoonsByGenre returns cartoons filtered by genre
func GetCartoonsByGenre(c *gin.Context) {
	genreName := c.Query("genre")
	if genreName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Genre name is required",
			"error":   "Please provide 'genre' query parameter",
		})
		return
	}

	var cartoons []models.Cartoon
	if err := database.DB.Joins("JOIN genres ON genres.id = cartoons.genre_id").
		Where("genres.name ILIKE ?", "%"+genreName+"%").
		Preload("Genre").Preload("AgeGroup").
		Find(&cartoons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch cartoons",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cartoons fetched successfully",
		"data":    cartoons,
		"count":   len(cartoons),
	})
}

// GetCartoonsByYear returns cartoons filtered by release year
func GetCartoonsByYear(c *gin.Context) {
	year := c.Query("year")
	if year == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Year is required",
			"error":   "Please provide 'year' query parameter",
		})
		return
	}

	var cartoons []models.Cartoon
	if err := database.DB.Where("release_year = ?", year).
		Preload("Genre").Preload("AgeGroup").
		Find(&cartoons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch cartoons",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cartoons fetched successfully",
		"data":    cartoons,
		"count":   len(cartoons),
	})
}

// GetCartoonsByAgeGroup returns cartoons filtered by age group
func GetCartoonsByAgeGroup(c *gin.Context) {
	ageGroupLabel := c.Query("age_group")
	if ageGroupLabel == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Age group label is required",
			"error":   "Please provide 'age_group' query parameter",
		})
		return
	}

	var cartoons []models.Cartoon
	if err := database.DB.Joins("JOIN age_groups ON age_groups.id = cartoons.age_group_id").
		Where("age_groups.label ILIKE ?", "%"+ageGroupLabel+"%").
		Preload("Genre").Preload("AgeGroup").
		Find(&cartoons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch cartoons",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cartoons fetched successfully",
		"data":    cartoons,
		"count":   len(cartoons),
	})
}
