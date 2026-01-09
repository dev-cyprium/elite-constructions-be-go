package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/auth"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()
	perPage := 10
	offset := (page - 1) * perPage

	// Get users
	users, err := queries.ListUsers(ctx, sqlc.ListUsersParams{
		Limit:  int32(perPage),
		Offset: int32(offset),
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Get total count
	total, err := queries.CountUsers(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Map sqlc users to models
	userModels := make([]models.User, len(users))
	for i, u := range users {
		userModels[i] = mapSQLCUserToModel(u)
	}

	SuccessResponse(c, http.StatusOK, models.PaginationResponse{
		Data:    userModels,
		Page:    page,
		PerPage: perPage,
		Total:   total,
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

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	user, err := queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	userModel := mapSQLCUserToModel(user)
	// Don't return password
	userModel.Password = ""
	SuccessResponse(c, http.StatusOK, userModel)
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

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Check if user already exists
	_, err = queries.GetUserByEmail(ctx, req.Email)
	if err == nil {
		ErrorResponse(c, http.StatusBadRequest, "User with this email already exists")
		return
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Create user
	user, err := queries.CreateUser(ctx, sqlc.CreateUserParams{
		Name:                  req.Name,
		Email:                 req.Email,
		Password:              hashedPassword,
		PasswordResetRequired: pgtype.Bool{Bool: false, Valid: true},
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	userModel := mapSQLCUserToModel(user)
	userModel.Password = "" // Don't return password
	SuccessResponse(c, http.StatusCreated, userModel)
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

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Check if user exists
	_, err = queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Update user
	err = queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	})
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to update user")
		return
	}

	// Get updated user
	user, err := queries.GetUserByID(ctx, id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	userModel := mapSQLCUserToModel(user)
	userModel.Password = "" // Don't return password
	SuccessResponse(c, http.StatusOK, userModel)
}

// DeleteUser deletes a user (returns 400 if only 1 remains)
func DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Check count - if only 1 remains, return 400
	count, err := queries.CountUsers(ctx)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	if count <= 1 {
		ErrorResponse(c, http.StatusBadRequest, "Cannot delete the last user")
		return
	}

	// Check if user exists
	_, err = queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Delete user
	err = queries.DeleteUser(ctx, id)
	if err != nil {
		ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user")
		return
	}

	c.Status(http.StatusNoContent)
}
