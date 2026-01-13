package handlers

import (
	"disney/database"
	"disney/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// CreateAdminLogRequest represents the request to create an admin log
type CreateAdminLogRequest struct {
	Action string `json:"action" binding:"required"`
	Entity string `json:"entity" binding:"required"`
}

// AdminLogResponse represents the admin log response
type AdminLogResponse struct {
	ID      uint `json:"id"`
	AdminID uint `json:"admin_id"`
	Admin   struct {
		ID    uint   `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"admin"`
	Action    string    `json:"action"`
	Entity    string    `json:"entity"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateAdminLog creates a new admin log entry
func CreateAdminLog(c *gin.Context) {
	log.Println("[CreateAdminLog] Received request")

	var req CreateAdminLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[CreateAdminLog] JSON binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Invalid request data"})
		return
	}

	log.Printf("[CreateAdminLog] Request: action=%s, entity=%s\n", req.Action, req.Entity)

	// Get admin ID from context (set by AuthRequired middleware)
	adminIDInterface, exists := c.Get("userID")
	if !exists {
		log.Println("[CreateAdminLog] userID not found in context")
		log.Printf("[CreateAdminLog] Available keys in context: %v\n", c.Keys)
		c.JSON(http.StatusUnauthorized, gin.H{"message": "User ID not found in context"})
		return
	}

	log.Printf("[CreateAdminLog] userID from context: %v (type: %T)\n", adminIDInterface, adminIDInterface)

	adminID, ok := adminIDInterface.(uint)
	if !ok {
		log.Println("[CreateAdminLog] userID is not uint, attempting to convert")
		// Try to convert if it's a different type
		if idStr, ok := adminIDInterface.(string); ok {
			if id, err := strconv.ParseUint(idStr, 10, 32); err == nil {
				adminID = uint(id)
				log.Printf("[CreateAdminLog] Successfully converted string to uint: %d\n", adminID)
			} else {
				log.Printf("[CreateAdminLog] Failed to parse userID string: %v\n", err)
				c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user ID"})
				return
			}
		} else if idInt, ok := adminIDInterface.(int); ok {
			adminID = uint(idInt)
			log.Printf("[CreateAdminLog] Converted int to uint: %d\n", adminID)
		} else if idInt64, ok := adminIDInterface.(int64); ok {
			adminID = uint(idInt64)
			log.Printf("[CreateAdminLog] Converted int64 to uint: %d\n", adminID)
		} else {
			log.Printf("[CreateAdminLog] Cannot convert userID to uint: type=%T, value=%v\n", adminIDInterface, adminIDInterface)
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid user ID type"})
			return
		}
	}

	log.Printf("[CreateAdminLog] Final adminID: %d\n", adminID)

	// Create admin log
	adminLog := models.AdminLog{
		AdminID: adminID,
		Action:  req.Action,
		Entity:  req.Entity,
	}

	log.Printf("[CreateAdminLog] Creating log entry: AdminID=%d, Action=%s, Entity=%s\n", adminLog.AdminID, adminLog.Action, adminLog.Entity)

	if err := database.DB.Create(&adminLog).Error; err != nil {
		log.Printf("[CreateAdminLog] Database error on Create: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("Failed to create admin log: %v", err)})
		return
	}

	log.Printf("[CreateAdminLog] Log created successfully with ID: %d\n", adminLog.ID)

	// Fetch the created log with admin details
	var logWithAdmin models.AdminLog
	if err := database.DB.Preload("Admin").First(&logWithAdmin, adminLog.ID).Error; err != nil {
		log.Printf("[CreateAdminLog] Database error on Preload: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch created log"})
		return
	}

	log.Printf("[CreateAdminLog] Log fetched successfully: %+v\n", logWithAdmin)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Admin log created successfully",
		"data": gin.H{
			"id":          logWithAdmin.ID,
			"admin_id":    logWithAdmin.AdminID,
			"admin_email": logWithAdmin.Admin.Email,
			"admin_name":  logWithAdmin.Admin.Name,
			"action":      logWithAdmin.Action,
			"entity":      logWithAdmin.Entity,
			"created_at":  logWithAdmin.CreatedAt,
		},
	})
}

// GetAdminLogs retrieves all admin logs with optional filtering
func GetAdminLogs(c *gin.Context) {
	log.Println("[GetAdminLogs] Received request")

	var logs []models.AdminLog
	query := database.DB.Preload("Admin")

	// Filter by action if provided
	if action := c.Query("action"); action != "" {
		log.Printf("[GetAdminLogs] Filtering by action: %s\n", action)
		query = query.Where("action = ?", action)
	}

	// Filter by entity if provided
	if entity := c.Query("entity"); entity != "" {
		log.Printf("[GetAdminLogs] Filtering by entity: %s\n", entity)
		query = query.Where("entity = ?", entity)
	}

	// Filter by date range if provided
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		log.Printf("[GetAdminLogs] Filtering from date: %s\n", dateFrom)
		query = query.Where("DATE(created_at) >= ?", dateFrom)
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		log.Printf("[GetAdminLogs] Filtering to date: %s\n", dateTo)
		query = query.Where("DATE(created_at) <= ?", dateTo)
	}

	// Filter by admin ID if provided (for viewing specific admin's logs)
	if adminID := c.Query("admin_id"); adminID != "" {
		log.Printf("[GetAdminLogs] Filtering by admin_id: %s\n", adminID)
		query = query.Where("admin_id = ?", adminID)
	}

	// Order by created_at descending (newest first)
	query = query.Order("created_at DESC")

	// Pagination
	page := 1
	pageSize := 50

	if p := c.Query("page"); p != "" {
		if parsedPage, err := strconv.Atoi(p); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	if ps := c.Query("page_size"); ps != "" {
		if parsedSize, err := strconv.Atoi(ps); err == nil && parsedSize > 0 && parsedSize <= 100 {
			pageSize = parsedSize
		}
	}

	log.Printf("[GetAdminLogs] Pagination: page=%d, pageSize=%d\n", page, pageSize)

	// Get total count for pagination
	var totalCount int64
	if err := query.Model(&models.AdminLog{}).Count(&totalCount).Error; err != nil {
		log.Printf("[GetAdminLogs] Database error on count: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to count logs"})
		return
	}

	log.Printf("[GetAdminLogs] Total logs found: %d\n", totalCount)

	// Apply pagination
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		log.Printf("[GetAdminLogs] Database error on find: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": "Failed to fetch admin logs"})
		return
	}

	log.Printf("[GetAdminLogs] Retrieved %d logs for this page\n", len(logs))

	// Build response
	var responseData []gin.H
	for _, logItem := range logs {
		adminEmail := "Unknown"
		adminName := "Unknown"

		// Safely handle nil Admin
		if logItem.Admin.ID > 0 {
			adminEmail = logItem.Admin.Email
			adminName = logItem.Admin.Name
		}

		log.Printf("[GetAdminLogs] Log ID=%d, AdminID=%d, AdminEmail=%s, Action=%s, Entity=%s\n",
			logItem.ID, logItem.AdminID, adminEmail, logItem.Action, logItem.Entity)

		responseData = append(responseData, gin.H{
			"id":       logItem.ID,
			"admin_id": logItem.AdminID,
			"admin": gin.H{
				"id":    logItem.Admin.ID,
				"email": adminEmail,
				"name":  adminName,
			},
			"action":     logItem.Action,
			"entity":     logItem.Entity,
			"created_at": logItem.CreatedAt,
		})
	}

	totalPages := (int(totalCount) + pageSize - 1) / pageSize

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin logs fetched successfully",
		"data":    responseData,
		"pagination": gin.H{
			"current_page": page,
			"page_size":    pageSize,
			"total_count":  totalCount,
			"total_pages":  totalPages,
		},
	})
}

// GetAdminLogStats retrieves statistics for admin logs
func GetAdminLogStats(c *gin.Context) {
	var totalLogs int64
	var createCount int64
	var updateCount int64
	var deleteCount int64

	// Get total logs
	database.DB.Model(&models.AdminLog{}).Count(&totalLogs)

	// Get counts by action
	database.DB.Model(&models.AdminLog{}).Where("action = ?", "CREATE").Count(&createCount)
	database.DB.Model(&models.AdminLog{}).Where("action = ?", "UPDATE").Count(&updateCount)
	database.DB.Model(&models.AdminLog{}).Where("action = ?", "DELETE").Count(&deleteCount)

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin log statistics fetched successfully",
		"data": gin.H{
			"total_logs":   totalLogs,
			"create_count": createCount,
			"update_count": updateCount,
			"delete_count": deleteCount,
		},
	})
}