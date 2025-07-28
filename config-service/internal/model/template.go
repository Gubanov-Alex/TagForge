package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// ConfigFormat represents the format of configuration template
type ConfigFormat string

const (
	ConfigFormatJSON ConfigFormat = "json"
	ConfigFormatYAML ConfigFormat = "yaml"
	ConfigFormatTOML ConfigFormat = "toml"
	ConfigFormatEnv  ConfigFormat = "env"
)

// Scan implements sql.Scanner interface
func (cf *ConfigFormat) Scan(value interface{}) error {
	if value == nil {
		*cf = ConfigFormatJSON
		return nil
	}
	if str, ok := value.(string); ok {
		*cf = ConfigFormat(str)
		return nil
	}
	return fmt.Errorf("cannot scan %T into ConfigFormat", value)
}

// Value implements driver.Valuer interface
func (cf ConfigFormat) Value() (driver.Value, error) {
	return string(cf), nil
}

// JSONMap represents JSON object stored in database
type JSONMap map[string]interface{}

// Scan implements sql.Scanner interface
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	return json.Unmarshal(bytes, j)
}

// Value implements driver.Valuer interface
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Template represents a configuration template
type Template struct {
	ID            int64        `json:"id" db:"id"`
	Name          string       `json:"name" db:"name" validate:"required,min=1,max=200"`
	Description   string       `json:"description" db:"description" validate:"max=1000"`
	Format        ConfigFormat `json:"format" db:"format" validate:"required,oneof=json yaml toml env"`
	Content       string       `json:"content" db:"content" validate:"required"`
	Schema        JSONMap      `json:"schema" db:"schema"`
	DefaultValues JSONMap      `json:"default_values" db:"default_values"`
	Version       string       `json:"version" db:"version" validate:"required,semver"`
	EnvironmentID int64        `json:"environment_id" db:"environment_id" validate:"required"`
	TagIDs        []int64      `json:"tag_ids" db:"-"`
	Tags          []Tag        `json:"tags,omitempty" db:"-"`
	Environment   Environment  `json:"environment,omitempty" db:"-"`
	Active        bool         `json:"active" db:"active"`
	CreatedAt     time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at" db:"updated_at"`
	CreatedBy     string       `json:"created_by" db:"created_by"`
	UpdatedBy     string       `json:"updated_by" db:"updated_by"`
}

// CreateTemplateRequest represents request for creating a template
type CreateTemplateRequest struct {
	Name          string       `json:"name" validate:"required,min=1,max=200"`
	Description   string       `json:"description" validate:"max=1000"`
	Format        ConfigFormat `json:"format" validate:"required,oneof=json yaml toml env"`
	Content       string       `json:"content" validate:"required"`
	Schema        JSONMap      `json:"schema,omitempty"`
	DefaultValues JSONMap      `json:"default_values,omitempty"`
	Version       string       `json:"version" validate:"required,semver"`
	EnvironmentID int64        `json:"environment_id" validate:"required"`
	TagIDs        []int64      `json:"tag_ids,omitempty"`
	Active        *bool        `json:"active,omitempty"`
	CreatedBy     string       `json:"created_by" validate:"required"`
}

// UpdateTemplateRequest represents request for updating a template
type UpdateTemplateRequest struct {
	Name          *string      `json:"name,omitempty" validate:"omitempty,min=1,max=200"`
	Description   *string      `json:"description,omitempty" validate:"omitempty,max=1000"`
	Format        *ConfigFormat `json:"format,omitempty" validate:"omitempty,oneof=json yaml toml env"`
	Content       *string      `json:"content,omitempty" validate:"omitempty,min=1"`
	Schema        JSONMap      `json:"schema,omitempty"`
	DefaultValues JSONMap      `json:"default_values,omitempty"`
	Version       *string      `json:"version,omitempty" validate:"omitempty,semver"`
	EnvironmentID *int64       `json:"environment_id,omitempty"`
	TagIDs        []int64      `json:"tag_ids,omitempty"`
	Active        *bool        `json:"active,omitempty"`
	UpdatedBy     string       `json:"updated_by" validate:"required"`
}

// TemplateResponse represents template response
type TemplateResponse struct {
	ID            int64               `json:"id"`
	Name          string              `json:"name"`
	Description   string              `json:"description"`
	Format        ConfigFormat        `json:"format"`
	Content       string              `json:"content"`
	Schema        JSONMap             `json:"schema"`
	DefaultValues JSONMap             `json:"default_values"`
	Version       string              `json:"version"`
	Environment   EnvironmentResponse `json:"environment"`
	Tags          []TagResponse       `json:"tags"`
	Active        bool                `json:"active"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
	CreatedBy     string              `json:"created_by"`
	UpdatedBy     string              `json:"updated_by"`
}

// TemplateListResponse represents paginated template list response
type TemplateListResponse struct {
	Templates []TemplateResponse `json:"templates"`
	Total     int64              `json:"total"`
	Page      int                `json:"page"`
	PageSize  int                `json:"page_size"`
	HasNext   bool               `json:"has_next"`
}
