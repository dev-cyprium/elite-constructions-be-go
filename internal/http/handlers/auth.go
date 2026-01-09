package handlers

import (
	"net/http"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/auth"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/config"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token         string `json:"token,omitempty"`
	ResetRequired bool   `json:"reset_required,omitempty"`
	ResetToken    string `json:"reset_token,omitempty"`
}

type PasswordResetRequest struct {
	ResetToken  string `json:"reset_token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

// Login handles user login
func Login(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

		// TODO: Query user from database using sqlc
		// For now, this is a placeholder
		_ = db.Pool

		// Check if user exists and has Argon2id password
		// If password_reset_required is true, generate reset token
		// Otherwise, verify password and return JWT

		// Placeholder response
		ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
	}
}

// Logout handles user logout (stateless, can be no-op)
func Logout(c *gin.Context) {
	SuccessResponse(c, http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetMe returns current user info
func GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// TODO: Query user from database
	_ = userID
	_ = db.Pool
	ErrorResponse(c, http.StatusNotFound, "User not found")
}

// CompletePasswordReset completes the password reset flow
func CompletePasswordReset(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PasswordResetRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

		// Hash the reset token
		tokenHash := auth.HashResetToken(req.ResetToken)

		// TODO: Find user by reset_token_hash
		// Verify token is not expired
		// Hash new password with Argon2id
		// Update user: set password, clear reset_token_hash, set password_reset_required=false

		_ = db.Pool
		_ = tokenHash
		_ = cfg

		ErrorResponse(c, http.StatusBadRequest, "Invalid or expired reset token")
	}
}
