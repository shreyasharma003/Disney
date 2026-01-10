package handlers

import (
	"disney/database"
	"disney/models"
	"disney/services"
	"fmt"
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
	ID          uint             `json:"id"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	PosterURL   string           `json:"poster_url"`
	ReleaseYear int              `json:"release_year"`
	GenreID     uint             `json:"genre_id"`
	AgeGroupID  uint             `json:"age_group_id"`
	IsFeatured  bool             `json:"is_featured"`
	CreatedAt   string           `json:"created_at"`
	UpdatedAt   string           `json:"updated_at"`
	IMDbRating  string           `json:"imdb_rating"`
	Genre       *models.Genre    `json:"genre,omitempty"`
	AgeGroup    *models.AgeGroup `json:"age_group,omitempty"`
}

// GetCartoonDetail returns cartoon details with IMDb rating
func GetCartoonDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid cartoon ID",
			"error":   "Cartoon ID must be a valid number",
		})
		return
	}

	var cartoon models.Cartoon
	if err := database.DB.Preload("Genre").Preload("AgeGroup").First(&cartoon, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Cartoon not found",
			"error":   err.Error(),
		})
		return
	}

	// Fetch IMDb rating from OMDb API
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

	// Fetch all cartoons from database
	if err := database.DB.Preload("Genre").Preload("AgeGroup").Find(&cartoons).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed to fetch cartoons",
			"error":   err.Error(),
		})
		return
	}

	// Fetch IMDb ratings for each cartoon and filter
	type cartoonWithRating struct {
		cartoon models.Cartoon
		rating  float64
	}

	var cartoonRatings []cartoonWithRating

	for _, cartoon := range cartoons {
		imdbRating := services.FetchIMDbRating(cartoon.Title)

		// Skip cartoons with N/A rating
		if imdbRating == "N/A" || imdbRating == "" {
			continue
		}

		// Convert string rating to float64 for sorting
		var ratingFloat float64
		fmt.Sscanf(imdbRating, "%f", &ratingFloat)

		cartoonRatings = append(cartoonRatings, cartoonWithRating{
			cartoon: cartoon,
			rating:  ratingFloat,
		})
	}

	// Sort by rating in descending order
	sort.Slice(cartoonRatings, func(i, j int) bool {
		return cartoonRatings[i].rating > cartoonRatings[j].rating
	})

	// Get top 5
	var topCartoons []TrendingCartoonResponse
	limit := 5
	if len(cartoonRatings) < limit {
		limit = len(cartoonRatings)
	}

	for i := 0; i < limit; i++ {
		cr := cartoonRatings[i]
		// Get the rating string from OMDb API again (could optimize with caching)
		imdbRating := services.FetchIMDbRating(cr.cartoon.Title)

		topCartoons = append(topCartoons, TrendingCartoonResponse{
			ID:          cr.cartoon.ID,
			Title:       cr.cartoon.Title,
			Description: cr.cartoon.Description,
			PosterURL:   cr.cartoon.PosterURL,
			ReleaseYear: cr.cartoon.ReleaseYear,
			IMDbRating:  imdbRating,
			Genre:       &cr.cartoon.Genre,
			AgeGroup:    &cr.cartoon.AgeGroup,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Top trending cartoons fetched successfully",
		"data":    topCartoons,
		"count":   len(topCartoons),
	})
}
