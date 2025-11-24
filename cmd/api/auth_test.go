package main

import (
	"sma/cmd/api/types"
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if hash == nil {
		t.Fatal("HashPassword() returned nil")
	}

	if hash.Hash == "" {
		t.Error("Hash is empty")
	}

	if hash.Salt == "" {
		t.Error("Salt is empty")
	}

	// Hash same password again - should produce different hash (different salt)
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() second call error = %v", err)
	}

	if hash.Hash == hash2.Hash {
		t.Error("Same password should produce different hashes due to different salts")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "TestPassword123!"
	wrongPassword := "WrongPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// Test correct password
	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if !valid {
		t.Error("VerifyPassword() = false, want true for correct password")
	}

	// Test wrong password
	valid, err = VerifyPassword(wrongPassword, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if valid {
		t.Error("VerifyPassword() = true, want false for wrong password")
	}
}

func TestGenerateJWT(t *testing.T) {
	user := &types.User{
		UserID:   1,
		Username: "testuser",
		Role:     types.RoleUser,
	}

	token, err := GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateJWT() returned empty token")
	}

	// Token should have three parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("token has %d parts, want 3", len(parts))
	}
}

func TestValidateJWT(t *testing.T) {
	user := &types.User{
		UserID:   1,
		Username: "testuser",
		Role:     types.RoleUser,
	}

	token, err := GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	// Test valid token
	claims, err := ValidateJWT(token)
	if err != nil {
		t.Fatalf("ValidateJWT() error = %v", err)
	}

	if claims.UserID != user.UserID {
		t.Errorf("UserID = %v, want %v", claims.UserID, user.UserID)
	}

	if claims.Username != user.Username {
		t.Errorf("Username = %v, want %v", claims.Username, user.Username)
	}

	if claims.Role != user.Role {
		t.Errorf("Role = %v, want %v", claims.Role, user.Role)
	}

	// Test invalid token
	_, err = ValidateJWT("invalid.token.here")
	if err == nil {
		t.Error("ValidateJWT() should return error for invalid token")
	}
}

func TestExtractTokenFromHeader(t *testing.T) {
	tests := []struct {
		name        string
		authHeader  string
		wantToken   string
		wantErr     bool
		errContains string
	}{
		{
			name:       "valid bearer token",
			authHeader: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			wantToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			wantErr:    false,
		},
		{
			name:        "empty header",
			authHeader:  "",
			wantToken:   "",
			wantErr:     true,
			errContains: "required",
		},
		{
			name:        "missing bearer prefix",
			authHeader:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			wantToken:   "",
			wantErr:     true,
			errContains: "Bearer",
		},
		{
			name:        "wrong prefix",
			authHeader:  "Basic token",
			wantToken:   "",
			wantErr:     true,
			errContains: "Bearer",
		},
		{
			name:       "bearer with lowercase",
			authHeader: "bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			wantToken:  "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.test",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := ExtractTokenFromHeader(tt.authHeader)

			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTokenFromHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil {
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error message = %v, want to contain %v", err.Error(), tt.errContains)
				}
			}

			if !tt.wantErr && token != tt.wantToken {
				t.Errorf("token = %v, want %v", token, tt.wantToken)
			}
		})
	}
}

func TestEncodeDecodePasswordHash(t *testing.T) {
	original := &PasswordHash{
		Hash: "testhash123",
		Salt: "testsalt456",
	}

	// Test encoding
	encoded := EncodePasswordHash(original)
	if encoded == "" {
		t.Error("EncodePasswordHash() returned empty string")
	}

	// Test decoding
	decoded, err := DecodePasswordHash(encoded)
	if err != nil {
		t.Fatalf("DecodePasswordHash() error = %v", err)
	}

	if decoded.Hash != original.Hash {
		t.Errorf("decoded Hash = %v, want %v", decoded.Hash, original.Hash)
	}

	if decoded.Salt != original.Salt {
		t.Errorf("decoded Salt = %v, want %v", decoded.Salt, original.Salt)
	}

	// Test decoding invalid format
	_, err = DecodePasswordHash("invalid-format")
	if err == nil {
		t.Error("DecodePasswordHash() should return error for invalid format")
	}
}

func TestGetUserIDFromJWT(t *testing.T) {
	user := &types.User{
		UserID:   42,
		Username: "testuser",
		Role:     types.RoleUser,
	}

	token, err := GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	userID, err := GetUserIDFromJWT(token)
	if err != nil {
		t.Fatalf("GetUserIDFromJWT() error = %v", err)
	}

	if userID != user.UserID {
		t.Errorf("userID = %v, want %v", userID, user.UserID)
	}

	// Test with invalid token
	_, err = GetUserIDFromJWT("invalid.token")
	if err == nil {
		t.Error("GetUserIDFromJWT() should return error for invalid token")
	}
}

func TestPasswordHashRoundTrip(t *testing.T) {
	password := "MySecurePassword123!"

	// Hash the password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	// Encode for storage
	encoded := EncodePasswordHash(hash)

	// Decode from storage
	decoded, err := DecodePasswordHash(encoded)
	if err != nil {
		t.Fatalf("DecodePasswordHash() error = %v", err)
	}

	// Verify password with decoded hash
	valid, err := VerifyPassword(password, decoded)
	if err != nil {
		t.Fatalf("VerifyPassword() error = %v", err)
	}

	if !valid {
		t.Error("Password verification failed after encode/decode round trip")
	}
}
