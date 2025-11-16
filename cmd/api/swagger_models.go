package main

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error" example:"Invalid request"`
	Message string `json:"message,omitempty" example:"Detailed error message"`
}

// OkResponse represents a successful operation response
type OkResponse struct {
	Message string `json:"message" example:"Operation successful"`
}

// AuthResponse represents authentication response
type AuthResponse struct {
	User  AuthUser `json:"user"`
	Token string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// AuthUser represents user data in auth response
type AuthUser struct {
	UserID   int    `json:"user_id" example:"1"`
	Username string `json:"username" example:"john_doe"`
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// PaginationResponse represents paginated response metadata
type PaginationResponse struct {
	Page       int `json:"page" example:"1"`
	PerPage    int `json:"per_page" example:"10"`
	Total      int `json:"total" example:"100"`
	TotalPages int `json:"total_pages" example:"10"`
}

// UsersResponse represents users list response
type UsersResponse struct {
	Data []UserResponse     `json:"data"`
	Meta PaginationResponse `json:"meta,omitempty"`
}

// UserResponse represents user response
type UserResponse struct {
	UserID   int    `json:"user_id" example:"1"`
	Username string `json:"username" example:"john_doe"`
}

// NodesResponse represents nodes list response
type NodesResponse struct {
	Data []NodeResponse     `json:"data"`
	Meta PaginationResponse `json:"meta,omitempty"`
}

// NodeResponse represents node response
type NodeResponse struct {
	DeviceID int    `json:"device_id" example:"1"`
	Status   string `json:"status" example:"ONLINE" enums:"ONLINE,OFFLINE,ERROR"`
}

// NodeDataResponse represents node data list response
type NodeDataResponse struct {
	Data []NodeDataItem     `json:"data"`
	Meta PaginationResponse `json:"meta,omitempty"`
}

// NodeDataItem represents a single node data item
type NodeDataItem struct {
	ID              int      `json:"id" example:"1"`
	UserID          int      `json:"user_id" example:"1"`
	DeviceID        int      `json:"device_id" example:"1"`
	MoistureContent *float64 `json:"moisture_content" example:"45.5"`
	Timestamp       string   `json:"timestamp" example:"2023-01-01T12:00:00Z"`
}

// NodeFavoritesResponse represents node favorites list response
type NodeFavoritesResponse struct {
	Data []NodeFavoriteItem `json:"data"`
}

// NodeFavoriteItem represents a single node favorite item
type NodeFavoriteItem struct {
	DeviceID int `json:"device_id" example:"1"`
}
