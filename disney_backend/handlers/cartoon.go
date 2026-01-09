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

// CreateCartoonRequest represents the request to create a new cartoon with characters
type CreateCartoonRequest struct {
	Title       string                 `json:"title" binding:"required"`
	Description string                 `json:"description"`
	PosterURL   string                 `json:"poster_url"`
	ReleaseYear int                    `json:"release_year" binding:"required"`
	GenreID     uint                   `json:"genre_id" binding:"required"`
	AgeGroupID  uint                   `json:"age_group_id" binding:"required"`
	IsFeatured  bool                   `json:"is_featured"`
	Characters  []CreateCharacterRequest `json:"characters"`
}

// CreateCharacterRequest represents a character in the cartoon
type CreateCharacterRequest struct {
	Name     string `json:"name" binding:"required"`
	ImageURL string `json:"image_url"`
}

// CreateCartoon creates a new cartoon with its characters
func CreateCartoon(c *gin.Context) {
	var req CreateCartoonRequest

	// Validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	// Verify genre exists
	var genre models.Genre
	if err := database.DB.First(&genre, req.GenreID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid genre ID",
			"error":   "Genre not found",
		})
		return
	}

	// Verify age group exists
	var ageGroup models.AgeGroup
	if err := database.DB.First(&ageGroup, req.AgeGroupID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid age group ID",
			"error":   "Age group not found",
		})
		return
	}

	// Create cartoon
	cartoon := models.Cartoon{
		Title:       req.Title,
		Description: req.Description,
		PosterURL:   req.PosterURL,
		ReleaseYear: req.ReleaseYear,
		GenreID:     req.GenreID,
		AgeGroupID:  req.AgeGroupID,
		IsFeatured:  req.IsFeatured,
	}

	// Start transaction
	tx := database.DB.Begin()

	// Create cartoon
	if err := tx.Create(&cartoon).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to create cartoon",
			"error":   err.Error(),
		})
		return
	}

	// Create characters if provided
	if len(req.Characters) > 0 {
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "Cartoon created successfully",
		"data":    cartoon,
	})
}
