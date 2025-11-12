package types

import (
	"sync"
	"time"
)

type Quote struct {
	ID        int       `json:"id"`
	Author    string    `json:"author"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type Comment struct {
	ID        int       `json:"id"`         // unique value for each comment
	Content   string    `json:"content"`    // the comment data
	Author    string    `json:"author"`     // the person who wrote the comment
	CreatedAt time.Time `json:"created_at"` // database timestamp
	Version   int       `json:"version"`    // incremented on each update
}

type User struct {
	ID           int        `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	PasswordHash string     `json:"-"` // Never include password hash in JSON response
	FirstName    *string    `json:"first_name,omitempty"`
	LastName     *string    `json:"last_name,omitempty"`
	Bio          *string    `json:"bio,omitempty"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	IsActive     bool       `json:"is_active"`
	IsVerified   bool       `json:"is_verified"`
	Role         string     `json:"role"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	Version      int        `json:"version"`
}

// UserCreateRequest represents the request payload for creating a new user
type UserCreateRequest struct {
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	Password  string  `json:"password"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Bio       *string `json:"bio,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// UserUpdateRequest represents the request payload for updating a user
type UserUpdateRequest struct {
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	Bio       *string `json:"bio,omitempty"`
	AvatarURL *string `json:"avatar_url,omitempty"`
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
