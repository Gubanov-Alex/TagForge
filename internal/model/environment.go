package model

import (
	"time"
)

// Environment represents a deployment environment
type Environment struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Slug        string    `json:"slug" db:"slug" validate:"required,min=1,max=100,alphanum"`
	Description string    `json:"description" db:"description" validate:"max=500"`
	Active      bool      `json:"active" db:"active"`
	Priority    int       `json:"priority" db:"priority" validate:"min=0,max=100"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateEnvironmentRequest represents request for creating an environment
type CreateEnvironmentRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Slug        string `json:"slug" validate:"required,min=1,max=100,alphanum"`
	Description string `json:"description" validate:"max=500"`
	Active      *bool  `json:"active,omitempty"`
	Priority    *int   `json:"priority,omitempty" validate:"omitempty,min=0,max=100"`
}

// UpdateEnvironmentRequest represents request for updating an environment
type UpdateEnvironmentRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Slug        *string `json:"slug,omitempty" validate:"omitempty,min=1,max=100,alphanum"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Active      *bool   `json:"active,omitempty"`
	Priority    *int    `json:"priority,omitempty" validate:"omitempty,min=0,max=100"`
}

// EnvironmentResponse represents environment response
type EnvironmentResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	Active      bool      `json:"active"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
