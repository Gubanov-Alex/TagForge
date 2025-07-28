package model

// ErrorResponse represents API error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message,omitempty"`
	Details map[string]string `json:"details,omitempty"`
}

// SuccessResponse represents API success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status   string                       `json:"status"`
	Version  string                       `json:"version"`
	Services map[string]ServiceHealthInfo `json:"services"`
}

// ServiceHealthInfo represents individual service health info
type ServiceHealthInfo struct {
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Latency   string `json:"latency,omitempty"`
	LastCheck string `json:"last_check"`
}

// PaginationParams represents pagination parameters
type PaginationParams struct {
	Page     int `form:"page,default=1" validate:"min=1"`
	PageSize int `form:"page_size,default=20" validate:"min=1,max=100"`
}

// FilterParams represents common filter parameters
type FilterParams struct {
	Search string `form:"search"`
	Active *bool  `form:"active"`
}

// SortParams represents sorting parameters
type SortParams struct {
	SortBy    string `form:"sort_by,default=created_at"`
	SortOrder string `form:"sort_order,default=desc" validate:"oneof=asc desc"`
}
