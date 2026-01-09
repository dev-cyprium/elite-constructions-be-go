package handlers

import (
	"net/http"
	"strconv"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/auth"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/gin-gonic/gin"
)

type CreateUserRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type UpdateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// GetUsers returns paginated users (10 per page)
func GetUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	// TODO: Implement with sqlc queries
	_ = db.Pool
	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    []models.User{},
		Page:    page,
		PerPage: 10,
		Total:   0,
	})
}

// GetUser returns a single user (no password)
func GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// TODO: Implement with sqlc queries
	_ = id
	_ = db.Pool
	ErrorResponse(c, http.StatusNotFound, "User not found")
}

// CreateUser creates a new user with Argon2id hashed password
func CreateUser(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Hash password with Argon2id
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// TODO: Implement with sqlc queries to insert user
	_ = db.Pool
	_ = hashedPassword

	user := models.User{
		Name:  req.Name,
		Email: req.Email,
		Password: hashedPassword,
		PasswordResetRequired: false,
	}
	SuccessResponse(c, http.StatusCreated, user)
}

// UpdateUser updates user name/email only (no password)
func UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// TODO: Implement with sqlc queries
	_ = id
	_ = db.Pool
	ErrorResponse(c, http.StatusNotFound, "User not found")
}

// DeleteUser deletes a user (returns 400 if only 1 remains)
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// TODO: Check count - if only 1 remains, return 400
	// Otherwise delete and return 204
	_ = id
	_ = db.Pool
	
	// Placeholder
	c.Status(http.StatusNoContent)
}
