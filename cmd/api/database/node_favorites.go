package database

import (
	"context"
	"fmt"
	"qotd/cmd/api/types"
)

// GetNodeFavorites fetches all node favorites from the database
func (db *Database) GetNodeFavorites() ([]types.NodeFavorite, error) {
	switch db.dbType {
	case InMemory:
		favorites := make([]types.NodeFavorite, len(InMemoryNodeFavorites))
		copy(favorites, InMemoryNodeFavorites)
		return favorites, nil
	case Postgres:
		query := `
			SELECT user_id, device_id
			FROM "NodeFavorites"
			ORDER BY user_id, device_id
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		rows, err := db.context.QueryContext(ctx, query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var favorites []types.NodeFavorite
		for rows.Next() {
			var nf types.NodeFavorite
			if err := rows.Scan(&nf.UserID, &nf.DeviceID); err != nil {
				return nil, err
			}
			favorites = append(favorites, nf)
		}
		return favorites, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// GetNodeFavoritesByUserID fetches all favorite nodes for a specific user
func (db *Database) GetNodeFavoritesByUserID(userID int) ([]types.NodeFavorite, error) {
	switch db.dbType {
	case InMemory:
		var result []types.NodeFavorite
		for _, favorite := range InMemoryNodeFavorites {
			if favorite.UserID == userID {
				result = append(result, favorite)
			}
		}
		return result, nil
	case Postgres:
		query := `
			SELECT user_id, device_id
			FROM "NodeFavorites"
			WHERE user_id = $1
			ORDER BY device_id
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		rows, err := db.context.QueryContext(ctx, query, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var favorites []types.NodeFavorite
		for rows.Next() {
			var nf types.NodeFavorite
			if err := rows.Scan(&nf.UserID, &nf.DeviceID); err != nil {
				return nil, err
			}
			favorites = append(favorites, nf)
		}
		return favorites, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// CreateNodeFavorite adds a node to a user's favorites
func (db *Database) CreateNodeFavorite(favorite types.NodeFavorite) error {
	// Check if the favorite already exists
	exists, err := db.NodeFavoriteExists(favorite.UserID, favorite.DeviceID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("node is already in user's favorites")
	}

	switch db.dbType {
	case InMemory:
		InMemoryNodeFavorites = append(InMemoryNodeFavorites, favorite)
		return nil
	case Postgres:
		// Create node favorite in Postgres database
		query := `
			INSERT INTO "NodeFavorites" (user_id, device_id)
			VALUES ($1, $2)
		`
		args := []any{favorite.UserID, favorite.DeviceID}
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

// NodeFavoriteExists checks if a specific user-device favorite relationship exists
func (db *Database) NodeFavoriteExists(userID, deviceID int) (bool, error) {
	switch db.dbType {
	case InMemory:
		for _, favorite := range InMemoryNodeFavorites {
			if favorite.UserID == userID && favorite.DeviceID == deviceID {
				return true, nil
			}
		}
		return false, nil
	case Postgres:
		query := `
			SELECT EXISTS(
				SELECT 1 FROM "NodeFavorites"
				WHERE user_id = $1 AND device_id = $2
			)
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		var exists bool
		err := db.context.QueryRowContext(ctx, query, userID, deviceID).Scan(&exists)
		if err != nil {
			return false, err
		}
		return exists, nil
	}
	return false, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// DeleteNodeFavorite removes a node from a user's favorites
func (db *Database) DeleteNodeFavorite(userID, deviceID int) error {
	switch db.dbType {
	case InMemory:
		for i, favorite := range InMemoryNodeFavorites {
			if favorite.UserID == userID && favorite.DeviceID == deviceID {
				InMemoryNodeFavorites = append(InMemoryNodeFavorites[:i], InMemoryNodeFavorites[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("favorite not found")
	case Postgres:
		// Delete node favorite from Postgres database
		query := `
			DELETE FROM "NodeFavorites"
			WHERE user_id = $1 AND device_id = $2
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		result, err := db.context.ExecContext(ctx, query, userID, deviceID)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return fmt.Errorf("favorite not found")
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

// DeleteNodeFavoritesByUserID removes all favorites for a specific user
func (db *Database) DeleteNodeFavoritesByUserID(userID int) error {
	switch db.dbType {
	case InMemory:
		var filteredFavorites []types.NodeFavorite
		for _, favorite := range InMemoryNodeFavorites {
			if favorite.UserID != userID {
				filteredFavorites = append(filteredFavorites, favorite)
			}
		}
		InMemoryNodeFavorites = filteredFavorites
		return nil
	case Postgres:
		query := `
			DELETE FROM "NodeFavorites"
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

// DeleteNodeFavoritesByDeviceID removes all favorites for a specific device
func (db *Database) DeleteNodeFavoritesByDeviceID(deviceID int) error {
	switch db.dbType {
	case InMemory:
		var filteredFavorites []types.NodeFavorite
		for _, favorite := range InMemoryNodeFavorites {
			if favorite.DeviceID != deviceID {
				filteredFavorites = append(filteredFavorites, favorite)
			}
		}
		InMemoryNodeFavorites = filteredFavorites
		return nil
	case Postgres:
		query := `
			DELETE FROM "NodeFavorites"
			WHERE device_id = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		_, err := db.context.ExecContext(ctx, query, deviceID)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

// ValidateNodeFavorite validates node favorite data before database operations
func ValidateNodeFavorite(favorite types.NodeFavorite) error {
	if favorite.UserID <= 0 {
		return fmt.Errorf("invalid user_id: must be positive")
	}
	if favorite.DeviceID <= 0 {
		return fmt.Errorf("invalid device_id: must be positive")
	}
	return nil
}
