package database

import (
	"context"
	"fmt"
	"qotd/cmd/api/types"
	"sort"
)

// sortNodes sorts nodes based on the given field and order
func sortNodes(nodes []types.Node, sortBy, sortOrder string) {
	switch sortBy {
	case "device_id":
		if sortOrder == "desc" {
			sort.Slice(nodes, func(i, j int) bool { return nodes[i].DeviceID > nodes[j].DeviceID })
		} else {
			sort.Slice(nodes, func(i, j int) bool { return nodes[i].DeviceID < nodes[j].DeviceID })
		}
	case "status":
		if sortOrder == "desc" {
			sort.Slice(nodes, func(i, j int) bool { return string(nodes[i].Status) > string(nodes[j].Status) })
		} else {
			sort.Slice(nodes, func(i, j int) bool { return string(nodes[i].Status) < string(nodes[j].Status) })
		}
	}
}

// GetNodes fetches all nodes from the database
func (db *Database) GetNodes() ([]types.Node, error) {
	return db.GetNodesWithPagination(0, 0, "", "")
}

// GetNodesWithPagination fetches nodes from the database with pagination and sorting
func (db *Database) GetNodesWithPagination(limit, offset int, sortBy, sortOrder string) ([]types.Node, error) {
	switch db.dbType {
	case InMemory:
		nodes := make([]types.Node, len(InMemoryNodes))
		copy(nodes, InMemoryNodes)

		// Apply sorting
		if sortBy != "" {
			sortNodes(nodes, sortBy, sortOrder)
		}

		// Apply pagination
		if limit > 0 {
			start := offset
			if start > len(nodes) {
				return []types.Node{}, nil
			}
			end := start + limit
			if end > len(nodes) {
				end = len(nodes)
			}
			return nodes[start:end], nil
		}
		return nodes, nil
	case Postgres:
		// Build the query with sorting and pagination
		query := `
			SELECT device_id, status
			FROM "Nodes"
		`

		// Add ORDER BY clause
		if sortBy != "" {
			orderBy := "device_id" // default
			switch sortBy {
			case "device_id":
				orderBy = "device_id"
			case "status":
				orderBy = "status"
			}

			order := "ASC"
			if sortOrder == "desc" {
				order = "DESC"
			}
			query += fmt.Sprintf(" ORDER BY %s %s", orderBy, order)
		} else {
			query += " ORDER BY device_id ASC"
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

		var nodes []types.Node
		for rows.Next() {
			var n types.Node
			if err := rows.Scan(&n.DeviceID, &n.Status); err != nil {
				return nil, err
			}
			nodes = append(nodes, n)
		}
		return nodes, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// CreateNode creates a new node in the database
func (db *Database) CreateNode(node types.Node) error {
	switch db.dbType {
	case InMemory:
		// Find the last node's ID and assign the next ID
		lastID := 0
		for _, n := range InMemoryNodes {
			if n.DeviceID > lastID {
				lastID = n.DeviceID
			}
		}
		node.DeviceID = lastID + 1
		InMemoryNodes = append(InMemoryNodes, node)
		return nil
	case Postgres:
		// Create node in Postgres database
		query := `
			INSERT INTO "Nodes" (status)
			VALUES ($1)
			RETURNING device_id
		`
		args := []any{node.Status}
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		err := db.context.QueryRowContext(ctx, query, args...).Scan(&node.DeviceID)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

// GetNodeByID fetches a specific node by device ID from the database
func (db *Database) GetNodeByID(deviceID int) (*types.Node, error) {
	switch db.dbType {
	case InMemory:
		for _, node := range InMemoryNodes {
			if node.DeviceID == deviceID {
				return &node, nil
			}
		}
		return nil, fmt.Errorf("node not found")
	case Postgres:
		// Fetch node by ID from Postgres database
		query := `
			SELECT device_id, status
			FROM "Nodes"
			WHERE device_id = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		var n types.Node
		err := db.context.QueryRowContext(ctx, query, deviceID).Scan(&n.DeviceID, &n.Status)
		if err != nil {
			return nil, err
		}
		return &n, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// UpdateNode modifies a node in the database
func (db *Database) UpdateNode(deviceID int, node types.Node) error {
	switch db.dbType {
	case InMemory:
		for i, n := range InMemoryNodes {
			if n.DeviceID == deviceID {
				InMemoryNodes[i] = node
				return nil
			}
		}
		return fmt.Errorf("node not found")
	case Postgres:
		// Update node in Postgres database
		query := `
			UPDATE "Nodes"
			SET status = $1
			WHERE device_id = $2
		`
		args := []any{node.Status, deviceID}
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

// DeleteNode removes a node from the database
func (db *Database) DeleteNode(deviceID int) error {
	switch db.dbType {
	case InMemory:
		for i, node := range InMemoryNodes {
			if node.DeviceID == deviceID {
				InMemoryNodes = append(InMemoryNodes[:i], InMemoryNodes[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("node not found")
	case Postgres:
		// Delete node from Postgres database
		query := `
			DELETE FROM "Nodes"
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

// ValidateNode validates node data before database operations
func ValidateNode(node types.Node) error {
	if node.Status != types.NodeStatusOnline && node.Status != types.NodeStatusOffline && node.Status != types.NodeStatusError {
		return fmt.Errorf("invalid node status: must be ONLINE, OFFLINE, or ERROR")
	}
	return nil
}
