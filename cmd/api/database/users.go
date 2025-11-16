package database

import (
	"context"
	"fmt"
	"qotd/cmd/api/types"
	"sort"
)

// sortUsers sorts users based on the given field and order
func sortUsers(users []types.User, sortBy, sortOrder string) {
	switch sortBy {
	case "user_id":
		if sortOrder == "desc" {
			sort.Slice(users, func(i, j int) bool { return users[i].UserID > users[j].UserID })
		} else {
			sort.Slice(users, func(i, j int) bool { return users[i].UserID < users[j].UserID })
		}
	case "username":
		if sortOrder == "desc" {
			sort.Slice(users, func(i, j int) bool { return users[i].Username > users[j].Username })
		} else {
			sort.Slice(users, func(i, j int) bool { return users[i].Username < users[j].Username })
		}
	}
}

// GetUsers fetches all users from the database
func (db *Database) GetUsers() ([]types.User, error) {
	return db.GetUsersWithPagination(0, 0, "", "")
}

// GetUsersWithPagination fetches users from the database with pagination and sorting
func (db *Database) GetUsersWithPagination(limit, offset int, sortBy, sortOrder string) ([]types.User, error) {
	switch db.dbType {
	case InMemory:
		users := make([]types.User, len(InMemoryUsers))
		copy(users, InMemoryUsers)

		// Apply sorting
		if sortBy != "" {
			sortUsers(users, sortBy, sortOrder)
		}

		// Apply pagination
		if limit > 0 {
			start := offset
			if start > len(users) {
				return []types.User{}, nil
			}
			end := start + limit
			if end > len(users) {
				end = len(users)
			}
			return users[start:end], nil
		}
		return users, nil
	case Postgres:
		// Build the query with sorting and pagination
		query := `
			SELECT user_id, username, password, is_verified
			FROM "Users"
		`

		// Add ORDER BY clause
		if sortBy != "" {
			orderBy := "user_id" // default
			switch sortBy {
			case "user_id":
				orderBy = "user_id"
			case "username":
				orderBy = "username"
			}

			order := "ASC"
			if sortOrder == "desc" {
				order = "DESC"
			}
			query += fmt.Sprintf(" ORDER BY %s %s", orderBy, order)
		} else {
			query += " ORDER BY user_id ASC"
		}

		// Add LIMIT and OFFSET
		if limit > 0 {
			query += fmt.Sprintf(" LIMIT %d OFFSET %d", limit, offset)
		}

		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		rows, err := db.context.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var users []types.User
		for rows.Next() {
			var u types.User
			if err := rows.Scan(&u.UserID, &u.Username, &u.Password, &u.IsVerified); err != nil {
				return nil, err
			}
			users = append(users, u)
		}
		return users, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// CreateUser creates a new user in the database
func (db *Database) CreateUser(user types.User) error {
	switch db.dbType {
	case InMemory:
		// Find the last user's ID and assign the next ID
		lastID := 0
		for _, u := range InMemoryUsers {
			if u.UserID > lastID {
				lastID = u.UserID
			}
		}
		user.UserID = lastID + 1
		InMemoryUsers = append(InMemoryUsers, user)
		return nil
	case Postgres:
		// Create user in Postgres database
		query := `
			INSERT INTO "Users" (username, password, is_verified)
			VALUES ($1, $2, $3)
			RETURNING user_id
		`
		args := []any{user.Username, user.Password, user.IsVerified}
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		err := db.context.QueryRowContext(ctx, query, args...).Scan(&user.UserID)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

// GetUserByID fetches a specific user by ID from the database
func (db *Database) GetUserByID(userID int) (*types.User, error) {
	switch db.dbType {
	case InMemory:
		for _, user := range InMemoryUsers {
			if user.UserID == userID {
				return &user, nil
			}
		}
		return nil, fmt.Errorf("user not found")
	case Postgres:
		// Fetch user by ID from Postgres database
		query := `
			SELECT user_id, username, password, is_verified
			FROM "Users"
			WHERE user_id = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		var u types.User
		err := db.context.QueryRowContext(ctx, query, userID).Scan(&u.UserID, &u.Username, &u.Password, &u.IsVerified)
		if err != nil {
			return nil, err
		}
		return &u, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// GetUserByUsername fetches a specific user by username from the database
func (db *Database) GetUserByUsername(username string) (*types.User, error) {
	switch db.dbType {
	case InMemory:
		for _, user := range InMemoryUsers {
			if user.Username == username {
				return &user, nil
			}
		}
		return nil, fmt.Errorf("user not found")
	case Postgres:
		// Fetch user by username from Postgres database
		query := `
			SELECT user_id, username, password, is_verified
			FROM "Users"
			WHERE username = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		var u types.User
		err := db.context.QueryRowContext(ctx, query, username).Scan(&u.UserID, &u.Username, &u.Password, &u.IsVerified)
		if err != nil {
			return nil, err
		}
		return &u, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// UpdateUser modifies a user in the database
func (db *Database) UpdateUser(userID int, user types.User) error {
	switch db.dbType {
	case InMemory:
		for i, u := range InMemoryUsers {
			if u.UserID == userID {
				user.UserID = userID // Preserve the original ID
				InMemoryUsers[i] = user
				return nil
			}
		}
		return fmt.Errorf("user not found")
	case Postgres:
		// Update user in Postgres database
		query := `
			UPDATE "Users"
			SET username = $1, password = $2, is_verified = $3
			WHERE user_id = $4
		`
		args := []any{user.Username, user.Password, user.IsVerified, userID}
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()
		_, err := db.context.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

// DeleteUser removes a user from the database
func (db *Database) DeleteUser(userID int) error {
	switch db.dbType {
	case InMemory:
		for i, user := range InMemoryUsers {
			if user.UserID == userID {
				InMemoryUsers = append(InMemoryUsers[:i], InMemoryUsers[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("user not found")
	case Postgres:
		// Delete user from Postgres database
		query := `
			DELETE FROM "Users"
			WHERE user_id = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		_, err := db.context.ExecContext(ctx, query, userID)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

// VerifyUser marks a user as verified in the database
func (db *Database) VerifyUser(userID int) error {
	switch db.dbType {
	case InMemory:
		for i, user := range InMemoryUsers {
			if user.UserID == userID {
				InMemoryUsers[i].IsVerified = true
				return nil
			}
		}
		return fmt.Errorf("user not found")
	case Postgres:
		// Update user verification status in Postgres database
		query := `
			UPDATE "Users"
			SET is_verified = true
			WHERE user_id = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		result, err := db.context.ExecContext(ctx, query, userID)
		if err != nil {
			return err
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}

		if rowsAffected == 0 {
			return fmt.Errorf("user not found")
		}

		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

// ValidateUser validates user data before database operations
func ValidateUser(user types.User) error {
	if user.Username == "" {
		return fmt.Errorf("username is required")
	}
	if len(user.Username) < 3 || len(user.Username) > 50 {
		return fmt.Errorf("username must be between 3 and 50 characters")
	}
	if user.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(user.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	return nil
}
