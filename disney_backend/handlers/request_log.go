package handlers

import (
	"disney/database"
	"disney/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetRequestLogs retrieves all request logs with optional filtering
func GetRequestLogs(c *gin.Context) {
	var logs []models.RequestLog
	query := database.DB.Preload("User")

	// Exclude admin routes - only show user requests
	query = query.Where("endpoint NOT LIKE ?", "/api/admin%")

	// Filter by method if provided
	if method := c.Query("method"); method != "" {
		query = query.Where("method = ?", method)
	}

	// Filter by status code range if provided
	if statusFilter := c.Query("status"); statusFilter != "" {
		switch statusFilter {
		case "2xx":
			query = query.Where("status_code >= ? AND status_code < ?", 200, 300)
		case "4xx":
			query = query.Where("status_code >= ? AND status_code < ?", 400, 500)
		case "5xx":
			query = query.Where("status_code >= ?", 500)
		}
	}

	// Filter by date range if provided
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		query = query.Where("DATE(created_at) >= ?", dateFrom)
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		query = query.Where("DATE(created_at) <= ?", dateTo)
	}

	// Filter by user ID if provided
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}

	// Order by created_at descending (newest first)
	query = query.Order("created_at DESC")

	// Pagination
	page := 1
	pageSize := 100

	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsedSize, err := strconv.Atoi(ps); err == nil && parsedSize > 0 && parsedSize <= 500 {
			pageSize = parsedSize
		}
	}

	// Get total count for pagination
	var totalCount int64
	if err := query.Model(&models.RequestLog{}).Count(&totalCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to count logs"})
		return
	}

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch request logs"})
		return
	}

	// Build response
	var responseData []gin.H
	for _, log := range logs {
		userEmail := "Anonymous"
		var userID uint = 0

		// Safely handle nil User
		if log.User != nil && log.User.ID > 0 {
			userEmail = log.User.Email
			userID = log.User.ID
		}

		responseData = append(responseData, gin.H{
			"id":            log.ID,
			"user_id":       userID,
			"user_email":    userEmail,
			"endpoint":      log.Endpoint,
			"method":        log.Method,
			"status_code":   log.StatusCode,
			"response_time": log.ResponseTime,
			"created_at":    log.CreatedAt,
		})
	}

	totalPages := (int(totalCount) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"message": "Request logs fetched successfully",
		"data":    responseData,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_count":  totalCount,
			"total_pages":  totalPages,
		},
	})
}

// GetRequestLogStats retrieves statistics for request logs
func GetRequestLogStats(c *gin.Context) {
	var totalRequests int64
	var successCount int64
	var clientErrorCount int64
	var serverErrorCount int64

	// Base query - exclude admin routes
	baseQuery := database.DB.Model(&models.RequestLog{}).Where("endpoint NOT LIKE ?", "/api/admin%")

	// Get total requests (excluding admin routes)
	baseQuery.Count(&totalRequests)

	// Get counts by status code range (excluding admin routes)
	database.DB.Model(&models.RequestLog{}).Where("endpoint NOT LIKE ?", "/api/admin%").Where("status_code >= ? AND status_code < ?", 200, 300).Count(&successCount)
	database.DB.Model(&models.RequestLog{}).Where("endpoint NOT LIKE ?", "/api/admin%").Where("status_code >= ? AND status_code < ?", 400, 500).Count(&clientErrorCount)
	database.DB.Model(&models.RequestLog{}).Where("endpoint NOT LIKE ?", "/api/admin%").Where("status_code >= ?", 500).Count(&serverErrorCount)

	c.JSON(http.StatusOK, gin.H{
		"message": "Request log statistics fetched successfully",
		"data": gin.H{
			"total_requests":      totalRequests,
			"success_count":       successCount,
			"client_error_count":  clientErrorCount,
			"server_error_count":  serverErrorCount,
		},
	})
}
