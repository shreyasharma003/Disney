package handlers

import (
	"disney/database"
	"disney/models"
	"disney/utils"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SignupRequest represents signup request payload
type SignupRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Age      int    `json:"age" binding:"required,min=1,max=120"` // age is required and between 1-120
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AdminCreateRequest represents admin creation request payload
type AdminCreateRequest struct {
	Name      string `json:"name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	Age       int    `json:"age" binding:"required,min=1,max=120"`
	SecretKey string `json:"secret_key" binding:"required"`
}

// UserResponse represents the user response (without password)
type UserResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Age       int    `json:"age"`
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
	// Hash the password before storing
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Message: "Error processing password",
			Error:   err.Error(),
		})
		return
	}

	newUser := models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword, // Store hashed password
		Age:          req.Age,        // Use age from request
		Role:         "user",         // Regular users only have 'user' role
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
				Age:       newUser.Age,
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

	// Find user by email from User model
	var user models.User
	if result := database.DB.Where("email = ?", req.Email).First(&user); result.RowsAffected == 0 {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Message: "Authentication failed",
			Error:   "Invalid email or password",
		})
		return
	}

	// Verify password using bcrypt
	if !utils.VerifyPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Message: "Authentication failed",
			Error:   "Invalid email or password",
		})
		return
	}

	// Generate JWT token (only on successful login)
	token, err := utils.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Message: "Error generating token",
			Error:   err.Error(),
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
				Age:       user.Age,
				Role:      user.Role,
				CreatedAt: user.CreatedAt.String(),
				UpdatedAt: user.UpdatedAt.String(),
			},
		},
		Token: token, // JWT token generated on login
	})
}

// CreateAdmin handles admin user creation with secret key validation
func CreateAdmin(c *gin.Context) {
	var req AdminCreateRequest

	// Validate request payload
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Message: "Invalid request",
			Error:   err.Error(),
		})
		return
	}

	// Validate secret key from environment
	adminSecretKey := os.Getenv("ADMIN_SECRET_KEY")
	if adminSecretKey == "" {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Message: "Server configuration error",
			Error:   "Admin secret key not configured",
		})
		return
	}

	// Verify secret key
	if req.SecretKey != adminSecretKey {
		c.JSON(http.StatusForbidden, AuthResponse{
			Message: "Unauthorized",
			Error:   "Invalid secret key",
		})
		return
	}

	// Validate password length
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Message: "Invalid request",
			Error:   "Password must be at least 6 characters long",
		})
		return
	}

	// Check if email already exists
	var existingUser models.User
	if result := database.DB.Where("email = ?", req.Email).First(&existingUser); result.RowsAffected > 0 {
		c.JSON(http.StatusConflict, AuthResponse{
			Message: "Email already registered",
			Error:   "Email already in use",
		})
		return
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Message: "Error processing password",
			Error:   err.Error(),
		})
		return
	}

	// Create new admin user
	newAdmin := models.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Age:          req.Age,
		Role:         "admin", // Set role as admin
	}

	// Save admin to database
	if err := database.DB.Create(&newAdmin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Message: "Error creating admin",
			Error:   err.Error(),
		})
		return
	}

	// Generate JWT token for the new admin
	token, err := utils.GenerateToken(newAdmin.ID, newAdmin.Email, newAdmin.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Message: "Error generating token",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		Message: "Admin created successfully",
		Data: map[string]interface{}{
			"admin": UserResponse{
				ID:        newAdmin.ID,
				Name:      newAdmin.Name,
				Email:     newAdmin.Email,
				Age:       newAdmin.Age,
				Role:      newAdmin.Role,
				CreatedAt: newAdmin.CreatedAt.String(),
				UpdatedAt: newAdmin.UpdatedAt.String(),
			},
		},
		Token: token,
	})
}
