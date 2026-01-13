package handlers

import (
	"disney/database"
	"disney/models"
	"disney/services"
	"net/http"
	"sort"
	"strconv"

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

// CartoonDetailResponse represents the cartoon detail response with IMDb rating
type CartoonDetailResponse struct {
	ID          uint               `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	PosterURL   string             `json:"poster_url"`
	ReleaseYear int                `json:"release_year"`
	GenreID     uint               `json:"genre_id"`
	AgeGroupID  uint               `json:"age_group_id"`
	IsFeatured  bool               `json:"is_featured"`
	CreatedAt   string             `json:"created_at"`
	UpdatedAt   string             `json:"updated_at"`
	IMDbRating  string             `json:"imdb_rating"`
	Genre       *models.Genre      `json:"genre,omitempty"`
	AgeGroup    *models.AgeGroup   `json:"age_group,omitempty"`
	Characters  []models.Character `json:"characters,omitempty"`
}

// GetCartoonByID retrieves a specific cartoon by its ID and tracks it as recently viewed
func GetCartoonByID(c *gin.Context) {
	cartoonID := c.Param("id")
	if cartoonID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Cartoon ID is required",
			"error":   "Please provide cartoon ID in the URL",
		})
		return
	}

	var cartoon models.Cartoon
	if err := database.DB.Preload("Genre").Preload("AgeGroup").Preload("Characters").
		First(&cartoon, cartoonID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Cartoon not found",
			"error":   err.Error(),
		})
		return
	}

	// Track the viewed cartoon in recently viewed list
	userID, exists := c.Get("userID")
	if exists && userID != nil {
		// userID from context is uint, convert to int
		uid := int(userID.(uint))
		cid := int(cartoon.ID)

		// Add to recently viewed cache (async - don't block response if Redis fails)
		go func() {
			if err := services.AddRecentlyViewed(uid, cid); err != nil {
				// Log error but don't fail the request
				// You can implement proper logging here
			}
		}()
	}

	// Fetch IMDb rating
	imdbRating := services.FetchIMDbRating(cartoon.Title)

	// Build response with IMDb rating
	response := CartoonDetailResponse{
		ID:          cartoon.ID,
		Title:       cartoon.Title,
		Description: cartoon.Description,
		PosterURL:   cartoon.PosterURL,
		ReleaseYear: cartoon.ReleaseYear,
		GenreID:     cartoon.GenreID,
		AgeGroupID:  cartoon.AgeGroupID,
		IsFeatured:  cartoon.IsFeatured,
		CreatedAt:   cartoon.CreatedAt.String(),
		UpdatedAt:   cartoon.UpdatedAt.String(),
		IMDbRating:  imdbRating,
		Genre:       &cartoon.Genre,
		AgeGroup:    &cartoon.AgeGroup,
		Characters:  cartoon.Characters,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cartoon fetched successfully",
		"data":    response,
	})
}

// TrendingCartoonResponse represents a cartoon in the trending list with IMDb rating
type TrendingCartoonResponse struct {
	ID          uint             `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	PosterURL   string           `json:"poster_url"`
	ReleaseYear int              `json:"release_year"`
	IMDbRating  string           `json:"imdb_rating"`
	Genre       *models.Genre    `json:"genre,omitempty"`
	AgeGroup    *models.AgeGroup `json:"age_group,omitempty"`
}

// GetTrendingCartoons returns top 5 cartoons sorted by IMDb rating
func GetTrendingCartoons(c *gin.Context) {
	var cartoons []models.Cartoon

	// Fetch only top 20 cartoons to reduce load
	if err := database.DB.Preload("Genre").Preload("AgeGroup").Limit(20).Find(&cartoons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch cartoons",
			"error":   err.Error(),
		})
		return
	}

	// Build response with cached IMDb ratings
	var trendingList []TrendingCartoonResponse
	for _, cartoon := range cartoons {
		imdbRating := services.FetchIMDbRating(cartoon.Title)
		trendingList = append(trendingList, TrendingCartoonResponse{
			ID:          cartoon.ID,
			Title:       cartoon.Title,
			Description: cartoon.Description,
			PosterURL:   cartoon.PosterURL,
			ReleaseYear: cartoon.ReleaseYear,
			IMDbRating:  imdbRating,
			Genre:       &cartoon.Genre,
			AgeGroup:    &cartoon.AgeGroup,
		})
	}

	// Sort by IMDb rating (descending)
	sort.Slice(trendingList, func(i, j int) bool {
		ratingI, errI := strconv.ParseFloat(trendingList[i].IMDbRating, 64)
		ratingJ, errJ := strconv.ParseFloat(trendingList[j].IMDbRating, 64)
		if errI != nil || errJ != nil {
			return false
		}
		return ratingI > ratingJ
	})

	// Return top 5
	if len(trendingList) > 5 {
		trendingList = trendingList[:5]
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Trending cartoons fetched successfully",
		"data":    trendingList,
		"count":   len(trendingList),
	})
}

// CreateCartoonRequest represents the request to create a new cartoon with characters
type CreateCartoonRequest struct {
	Title       string                   `json:"title" binding:"required"`
	Description string                   `json:"description"`
	PosterURL   string                   `json:"poster_url"`
	ReleaseYear int                      `json:"release_year" binding:"required"`
	GenreID     uint                     `json:"genre_id" binding:"required"`
	AgeGroupID  uint                     `json:"age_group_id" binding:"required"`
	IsFeatured  bool                     `json:"is_featured"`
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

	// Log admin action
	if adminID, exists := c.Get("userID"); exists {
		adminLog := models.AdminLog{
			AdminID: adminID.(uint),
			Action:  "CREATE",
			Entity:  "Cartoon: " + cartoon.Title,
		}
		database.DB.Create(&adminLog)
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Cartoon created successfully",
		"data":    cartoon,
	})
}

// DeleteCartoon deletes a cartoon by ID or title
func DeleteCartoon(c *gin.Context) {
	cartoonID := c.Query("id")
	cartoonTitle := c.Query("title")

	if cartoonID == "" && cartoonTitle == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Cartoon ID or title is required",
			"error":   "Please provide 'id' or 'title' query parameter",
		})
		return
	}

	var cartoon models.Cartoon
	var query = database.DB

	// Search by ID or title
	if cartoonID != "" {
		query = query.Where("id = ?", cartoonID)
	} else {
		query = query.Where("title ILIKE ?", "%"+cartoonTitle+"%")
	}

	// Find cartoon
	if err := query.First(&cartoon).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Cartoon not found",
			"error":   err.Error(),
		})
		return
	}

	// Delete cartoon (characters will be deleted automatically due to CASCADE)
	if err := database.DB.Delete(&cartoon).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to delete cartoon",
			"error":   err.Error(),
		})
		return
	}

	// Log admin action
	if adminID, exists := c.Get("userID"); exists {
		adminLog := models.AdminLog{
			AdminID: adminID.(uint),
			Action:  "DELETE",
			Entity:  "Cartoon: " + cartoon.Title,
		}
		database.DB.Create(&adminLog)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Cartoon deleted successfully",
		"data": gin.H{
			"id":    cartoon.ID,
			"title": cartoon.Title,
		},
	})
}

// GetRecentlyViewedRequest is the response structure for recently viewed cartoons
type GetRecentlyViewedResponse struct {
	CartoonID   uint            `json:"cartoon_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	PosterURL   string          `json:"poster_url"`
	ReleaseYear int             `json:"release_year"`
	Genre       models.Genre    `json:"genre,omitempty"`
	AgeGroup    models.AgeGroup `json:"age_group,omitempty"`
}

// GetRecentlyViewed fetches the recently viewed cartoons for the authenticated user
func GetRecentlyViewed(c *gin.Context) {
	// Get user ID from context (set by AuthRequired middleware)
	userIDInterface, exists := c.Get("userID")
	if !exists || userIDInterface == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "User not authenticated",
			"error":   "User ID not found in context",
		})
		return
	}

	// Convert uint to int
	userID := int(userIDInterface.(uint))

	// Get cartoon IDs from Redis
	cartoonIDs, err := services.GetRecentlyViewed(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch recently viewed cartoons",
			"error":   err.Error(),
		})
		return
	}

	// If no recently viewed cartoons, return empty list
	if len(cartoonIDs) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No recently viewed cartoons",
			"data":    []GetRecentlyViewedResponse{},
			"count":   0,
		})
		return
	}

	// Fetch full cartoon details from database in the order they appear in Redis
	var response []GetRecentlyViewedResponse

	for _, cartoonID := range cartoonIDs {
		var cartoon models.Cartoon
		if err := database.DB.Preload("Genre").Preload("AgeGroup").
			First(&cartoon, cartoonID).Error; err != nil {
			// Skip if cartoon not found, continue with others
			continue
		}

		response = append(response, GetRecentlyViewedResponse{
			CartoonID:   cartoon.ID,
			Title:       cartoon.Title,
			Description: cartoon.Description,
			PosterURL:   cartoon.PosterURL,
			ReleaseYear: cartoon.ReleaseYear,
			Genre:       cartoon.Genre,
			AgeGroup:    cartoon.AgeGroup,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Recently viewed cartoons fetched successfully",
		"data":    response,
		"count":   len(response),
	})
}
