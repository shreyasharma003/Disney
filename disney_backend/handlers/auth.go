package handlers

import (
	"disney/database"
	"disney/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SignupRequest represents signup request payload
type SignupRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=user admin"` // role: user or admin
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents the user response (without password)
type UserResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// AuthResponse represents the auth response
type AuthResponse struct {
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Token   string                 `json:"token,omitempty"`
}

// Signup handles user registration
func Signup(c *gin.Context) {
	var req SignupRequest

	// Validate request payload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	// Validate password length (minimum 6 characters)
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Message: "Invalid request",
			Error:   "Password must be at least 6 characters long",
		})
		return
	}

	// TODO: Check if email already exists
	var existingUser models.User
	if result := database.DB.Where("email = ?", req.Email).First(&existingUser); result.RowsAffected > 0 {
		c.JSON(http.StatusConflict, AuthResponse{
			Message: "Email already registered",
			Error:   "Email already in use",
		})
		return
	}

	// Create new user with fields from User model
	newUser := models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: req.Role, 
		Role:         "user",   
	}

	// Save user to database
	if err := database.DB.Create(&newUser).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Message: "Error creating user",
			Error:   err.Error(),
		})
		return
	}



	c.JSON(http.StatusCreated, AuthResponse{
		Message: "User registered successfully",
		Data: map[string]interface{}{
			"user": UserResponse{
				ID:        newUser.ID,
				Name:      newUser.Name,
				Email:     newUser.Email,
				Role:      newUser.Role,
				CreatedAt: newUser.CreatedAt.String(),
				UpdatedAt: newUser.UpdatedAt.String(),
			},
		},
	})
}

// Login handles user authentication
func Login(c *gin.Context) {
	var req LoginRequest

	// Validate request payload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	
	var user models.User
	if result := database.DB.Where("email = ?", req.Email).First(&user); result.RowsAffected == 0 {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Message: "Authentication failed",
			Error:   "Invalid email or password",
		})
		return
	}

	

	if req.Password != user.PasswordHash {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Message: "Authentication failed",
			Error:   "Invalid email or password",
		})
		return
	}

	

	c.JSON(http.StatusOK, AuthResponse{
		Message: "Login successful",
		Data: map[string]interface{}{
			"user": UserResponse{
				ID:        user.ID,
				Name:      user.Name,
				Email:     user.Email,
				Role:      user.Role,
				CreatedAt: user.CreatedAt.String(),
				UpdatedAt: user.UpdatedAt.String(),
			},
		},
		Token: "jwt_token_placeholder", // TODO: Replace with actual token
	})
}
