package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"qotd/cmd/api/types"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/argon2"
)

// JWT secret key - in production, this should be loaded from environment variable
var jwtSecret = []byte("your-secret-key-change-this-in-production")

// Argon2 parameters
const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024
	argon2Threads = 4
	argon2KeyLen  = 32
	saltLength    = 16
)

// AuthClaims represents JWT claims for authentication
type AuthClaims struct {
	UserID   int            `json:"user_id"`
	Username string         `json:"username"`
	Role     types.UserRole `json:"role"`
	jwt.RegisteredClaims
}

// PasswordHash represents a hashed password with salt
type PasswordHash struct {
	Hash string
	Salt string
}

// HashPassword hashes a password using Argon2id
func HashPassword(password string) (*PasswordHash, error) {
	// Generate random salt
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}

	// Hash password using Argon2id
	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	return &PasswordHash{
		Hash: base64.URLEncoding.EncodeToString(hash),
		Salt: base64.URLEncoding.EncodeToString(salt),
	}, nil
}

// VerifyPassword verifies a password against its hash
func VerifyPassword(password string, storedHash *PasswordHash) (bool, error) {
	// Decode the stored salt and hash
	salt, err := base64.URLEncoding.DecodeString(storedHash.Salt)
	if err != nil {
		return false, err
	}

	hash, err := base64.URLEncoding.DecodeString(storedHash.Hash)
	if err != nil {
		return false, err
	}

	// Hash the provided password with the same salt
	providedHash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Use constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(hash, providedHash) == 1, nil
}

// GenerateJWT generates a JWT token for a user
func GenerateJWT(user *types.User) (string, error) {
	// Create claims
	claims := AuthClaims{
		UserID:   user.UserID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Token expires in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "soil-quality-monitor",
			Subject:   fmt.Sprintf("user:%d", user.UserID),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateJWT validates a JWT token and returns the claims
func ValidateJWT(tokenString string) (*AuthClaims, error) {
	// Parse token
	token, err := jwt.ParseWithClaims(tokenString, &AuthClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// Validate token and extract claims
	if claims, ok := token.Claims.(*AuthClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ExtractTokenFromHeader extracts JWT token from Authorization header
func ExtractTokenFromHeader(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	// Expected format: "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return "", errors.New("authorization header format must be 'Bearer <token>'")
	}

	return parts[1], nil
}

// EncodePasswordHash encodes a PasswordHash for database storage
func EncodePasswordHash(ph *PasswordHash) string {
	return ph.Hash + ":" + ph.Salt
}

// DecodePasswordHash decodes a stored password hash from database
func DecodePasswordHash(encoded string) (*PasswordHash, error) {
	parts := strings.SplitN(encoded, ":", 2)
	if len(parts) != 2 {
		return nil, errors.New("invalid password hash format")
	}

	return &PasswordHash{
		Hash: parts[0],
		Salt: parts[1],
	}, nil
}

func GetUserIDFromJWT(tokenString string) (int, error) {
	claims, err := ValidateJWT(tokenString)
	if err != nil {
		return 0, err
	}
	return claims.UserID, nil
}
