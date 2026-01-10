package services

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
)

// OMDbResponse represents the response structure from OMDb API
type OMDbResponse struct {
	Title      string `json:"Title"`
	Year       string `json:"Year"`
	ImdbRating string `json:"imdbRating"`
	Response   string `json:"Response"`
	Error      string `json:"Error"`
}

// FetchIMDbRating fetches the IMDb rating for a cartoon from OMDb API
// Returns the imdbRating string or "N/A" if not found or API fails
func FetchIMDbRating(cartoonTitle string) string {
	// Get API key from environment variables
	apiKey := os.Getenv("OMDB_API_KEY")
	if apiKey == "" {
		return "N/A"
	}

	// Build OMDb API URL with query parameters
	baseURL := "http://www.omdbapi.com/"
	params := url.Values{}
	params.Set("t", cartoonTitle)
	params.Set("apikey", apiKey)
	params.Set("type", "series") // Search for series/cartoons specifically

	fullURL := baseURL + "?" + params.Encode()

	// Make HTTP GET request
	resp, err := http.Get(fullURL)
	if err != nil {
		return "N/A"
	}
	defer resp.Body.Close()

	// Check if response status is OK
	if resp.StatusCode != http.StatusOK {
		return "N/A"
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "N/A"
	}

	// Parse JSON response
	var omdbResp OMDbResponse
	if err := json.Unmarshal(body, &omdbResp); err != nil {
		return "N/A"
	}

	// Check if API returned an error
	if omdbResp.Response == "False" {
		return "N/A"
	}

	// Return imdbRating or N/A if empty
	if omdbResp.ImdbRating == "" || omdbResp.ImdbRating == "N/A" {
		return "N/A"
	}

	return omdbResp.ImdbRating
}
