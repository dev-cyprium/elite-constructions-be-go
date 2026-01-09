package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/dev-cyprium/elite-constructions-be-v2/internal/auth"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/config"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/db"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/models"
	"github.com/dev-cyprium/elite-constructions-be-v2/internal/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

		queries := sqlc.New(db.Pool)
		ctx := c.Request.Context()

		// Query user from database
		user, err := queries.GetUserByEmail(ctx, req.Email)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
				return
			}
			ErrorResponse(c, http.StatusInternalServerError, "Database error")
			return
		}

		// Check if password reset is required
		if user.PasswordResetRequired.Bool {
			// Generate reset token
			resetToken, err := auth.GenerateResetToken()
			if err != nil {
				ErrorResponse(c, http.StatusInternalServerError, "Failed to generate reset token")
				return
			}

			tokenHash := auth.HashResetToken(resetToken)
			expiresAt := time.Now().Add(auth.ResetTokenExpiry)

			// Update user with reset token
			err = queries.UpdateUserResetToken(ctx, sqlc.UpdateUserResetTokenParams{
				ID: user.ID,
				ResetTokenHash: pgtype.Text{
					String: tokenHash,
					Valid:  true,
				},
				ResetTokenExpiresAt: pgtype.Timestamp{
					Time:   expiresAt,
					Valid:  true,
				},
			})
			if err != nil {
				ErrorResponse(c, http.StatusInternalServerError, "Failed to update reset token")
				return
			}

			SuccessResponse(c, http.StatusOK, LoginResponse{
				ResetRequired: true,
				ResetToken:    resetToken,
			})
			return
		}

		// Verify password
		valid, err := auth.VerifyPassword(req.Password, user.Password)
		if err != nil {
			// Log the error for debugging
			c.Error(err)
			ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials", err.Error())
			return
		}
		if !valid {
			ErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		// Generate JWT token
		token, err := auth.GenerateToken(user.ID, user.Email, cfg.JWTSecret)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		SuccessResponse(c, http.StatusOK, LoginResponse{
			Token: token,
		})
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

	queries := sqlc.New(db.Pool)
	ctx := c.Request.Context()

	// Query user from database
	id, ok := userID.(int64)
	if !ok {
		ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	user, err := queries.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			ErrorResponse(c, http.StatusNotFound, "User not found")
			return
		}
		ErrorResponse(c, http.StatusInternalServerError, "Database error")
		return
	}

	// Map sqlc User to models.User
	userModel := mapSQLCUserToModel(user)
	SuccessResponse(c, http.StatusOK, userModel)
}

// CompletePasswordReset completes the password reset flow
func CompletePasswordReset(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req PasswordResetRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			ErrorResponse(c, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

		queries := sqlc.New(db.Pool)
		ctx := c.Request.Context()

		// Hash the reset token
		tokenHash := auth.HashResetToken(req.ResetToken)

		// Find user by reset_token_hash (query already checks expiration)
		user, err := queries.GetUserByResetTokenHash(ctx, pgtype.Text{
			String: tokenHash,
			Valid:  true,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				ErrorResponse(c, http.StatusBadRequest, "Invalid or expired reset token")
				return
			}
			ErrorResponse(c, http.StatusInternalServerError, "Database error")
			return
		}

		// Hash new password with Argon2id
		hashedPassword, err := auth.HashPassword(req.NewPassword)
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
			return
		}

		// Update user: set password, clear reset_token_hash, set password_reset_required=false
		err = queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
			ID:       user.ID,
			Password: hashedPassword,
		})
		if err != nil {
			ErrorResponse(c, http.StatusInternalServerError, "Failed to update password")
			return
		}

		SuccessResponse(c, http.StatusOK, gin.H{"message": "Password reset successfully"})
	}
}

// Helper function to map sqlc User to models.User
func mapSQLCUserToModel(u sqlc.User) models.User {
	var emailVerifiedAt *time.Time
	if u.EmailVerifiedAt.Valid {
		emailVerifiedAt = &u.EmailVerifiedAt.Time
	}

	var resetTokenHash *string
	if u.ResetTokenHash.Valid {
		resetTokenHash = &u.ResetTokenHash.String
	}

	var resetTokenExpiresAt *time.Time
	if u.ResetTokenExpiresAt.Valid {
		resetTokenExpiresAt = &u.ResetTokenExpiresAt.Time
	}

	var rememberToken *string
	if u.RememberToken.Valid {
		rememberToken = &u.RememberToken.String
	}

	return models.User{
		ID:                    u.ID,
		Name:                  u.Name,
		Email:                 u.Email,
		EmailVerifiedAt:       emailVerifiedAt,
		Password:              u.Password,
		PasswordResetRequired: u.PasswordResetRequired.Bool,
		ResetTokenHash:        resetTokenHash,
		ResetTokenExpiresAt:   resetTokenExpiresAt,
		RememberToken:         rememberToken,
		CreatedAt:             u.CreatedAt.Time,
		UpdatedAt:             u.UpdatedAt.Time,
	}
}
