package main

import (
	"fmt"
	"sma/cmd/api/types"
)

// InitTestUsers creates test users for demo purposes
func InitTestUsers(config *serverConfig) error {
	// Create admin user
	adminHash, err := HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	adminUser := types.User{
		Username:   "admin",
		Password:   EncodePasswordHash(adminHash),
		IsVerified: true,
		Role:       types.RoleAdmin,
	}

	if err := config.db.CreateUser(adminUser); err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	// Create regular user
	userHash, err := HashPassword("user123")
	if err != nil {
		return fmt.Errorf("failed to hash user password: %w", err)
	}

	regularUser := types.User{
		Username:   "testuser",
		Password:   EncodePasswordHash(userHash),
		IsVerified: true,
		Role:       types.RoleUser,
	}

	if err := config.db.CreateUser(regularUser); err != nil {
		return fmt.Errorf("failed to create regular user: %w", err)
	}

	fmt.Println("✓ Test users created:")
	fmt.Println("  Admin user: username=admin, password=admin123")
	fmt.Println("  Regular user: username=testuser, password=user123")

	return nil
}
