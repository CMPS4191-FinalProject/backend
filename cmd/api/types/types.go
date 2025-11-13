package types

import (
	"sync"
	"time"
)

// NodeStatus represents the status of a soil monitoring node
type NodeStatus string

const (
	NodeStatusOnline  NodeStatus = "ONLINE"
	NodeStatusOffline NodeStatus = "OFFLINE"
	NodeStatusError   NodeStatus = "ERROR"
)

// User represents a user in the soil monitoring system
type User struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Password string `json:"-"` // Never include password in JSON response
}

// Node represents a soil monitoring IoT device
type Node struct {
	DeviceID int        `json:"device_id"`
	Status   NodeStatus `json:"status"`
}

// NodeData represents sensor data from a soil monitoring node
type NodeData struct {
	ID              int       `json:"id"`
	UserID          int       `json:"user_id"`
	DeviceID        int       `json:"device_id"`
	MoistureContent *float64  `json:"moisture_content"` // Nullable float in database
	Timestamp       time.Time `json:"timestamp"`
}

// NodeFavorite represents a user's favorite monitoring node
type NodeFavorite struct {
	UserID   int `json:"user_id"`
	DeviceID int `json:"device_id"`
}

// UserCreateRequest represents the request payload for creating a new user
type UserCreateRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// UserUpdateRequest represents the request payload for updating a user
type UserUpdateRequest struct {
	Username *string `json:"username,omitempty"`
	Password *string `json:"password,omitempty"`
}

// NodeCreateRequest represents the request payload for creating a new node
type NodeCreateRequest struct {
	Status NodeStatus `json:"status"`
}

// NodeUpdateRequest represents the request payload for updating a node
type NodeUpdateRequest struct {
	Status *NodeStatus `json:"status,omitempty"`
}

// NodeDataCreateRequest represents the request payload for creating node data
type NodeDataCreateRequest struct {
	DeviceID        int      `json:"device_id"`
	MoistureContent *float64 `json:"moisture_content"`
}

// NodeFavoriteCreateRequest represents the request payload for adding a favorite node
type NodeFavoriteCreateRequest struct {
	DeviceID int `json:"device_id"`
}

// RateLimiter represents a token bucket rate limiter
type RateLimiter struct {
	Tokens     float64
	MaxTokens  float64
	RefillRate float64
	LastRefill time.Time
	Mutex      sync.Mutex
}

// Allow checks if a request is allowed based on the rate limit
func (rl *RateLimiter) Allow() bool {
	rl.Mutex.Lock()
	defer rl.Mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.LastRefill).Seconds()

	// Refill tokens based on elapsed time
	rl.Tokens += elapsed * rl.RefillRate
	if rl.Tokens > rl.MaxTokens {
		rl.Tokens = rl.MaxTokens
	}

	rl.LastRefill = now

	// Check if we have enough tokens
	if rl.Tokens >= 1 {
		rl.Tokens--
		return true
	}

	return false
}
