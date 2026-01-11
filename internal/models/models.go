package models

import "time"

// User represents a user in the system
type User struct {
	ID                    int64      `json:"id"`
	Name                  string     `json:"name"`
	Email                 string     `json:"email"`
	EmailVerifiedAt       *time.Time `json:"email_verified_at,omitempty"`
	Password              string     `json:"-"` // Argon2id hash
	PasswordResetRequired bool       `json:"password_reset_required"`
	ResetTokenHash        *string    `json:"-"` // SHA256 of reset token
	ResetTokenExpiresAt   *time.Time `json:"-"`
	RememberToken         *string    `json:"-"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// Project represents a construction project
type Project struct {
	ID          int64          `json:"id"`
	Status      int            `json:"status"`
	Name        string         `json:"name"`
	Category    *string        `json:"category,omitempty"`
	Client      *string        `json:"client,omitempty"`
	Order       int            `json:"order"`
	Highlighted bool           `json:"highlighted"`
	Images      []ProjectImage `json:"images,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ProjectImage represents an image associated with a project
type ProjectImage struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	URL        string    `json:"url"` // /storage/img/filename.jpg
	ProjectID  int64     `json:"project_id"`
	Order      int       `json:"order"`
	BlurHash   *string   `json:"blur_hash,omitempty"` // data URL
	Highlighted bool     `json:"highlighted"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Testimonial represents a customer testimonial
type Testimonial struct {
	ID          int64     `json:"id"`
	FullName    string    `json:"full_name"`
	Profession  string    `json:"profession"`
	Testimonial string    `json:"testimonial"`
	Status      string    `json:"status"` // "ready", "pending", etc.
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// StaticText represents a static text content item
type StaticText struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Label     string    `json:"label"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Configuration represents a configuration key-value pair
type Configuration struct {
	ID        int64     `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VisitorMessage represents a message from a visitor
type VisitorMessage struct {
	ID          int64     `json:"id"`
	Email       string    `json:"email"`
	Address     string    `json:"address"`
	Description string    `json:"description"`
	Seen        bool      `json:"seen"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// PaginationResponse represents a paginated response
type PaginationResponse struct {
	Data    interface{} `json:"data"`
	Page    int         `json:"page"`
	PerPage int         `json:"per_page"`
	Total   int64       `json:"total"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}
