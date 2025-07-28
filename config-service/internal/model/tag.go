package model

import (
	"time"
)

// Tag represents a configuration tag
type Tag struct {
	ID          int64     `json:"id" db:"id"`
	Name        string    `json:"name" db:"name" validate:"required,min=1,max=100"`
	Description string    `json:"description" db:"description" validate:"max=500"`
	Color       string    `json:"color" db:"color" validate:"hexcolor"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateTagRequest represents request for creating a tag
type CreateTagRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	Color       string `json:"color" validate:"required,hexcolor"`
}

// UpdateTagRequest represents request for updating a tag
type UpdateTagRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Color       *string `json:"color,omitempty" validate:"omitempty,hexcolor"`
}

// TagResponse represents tag response
type TagResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
