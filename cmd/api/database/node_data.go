package database

import (
	"context"
	"fmt"
	"qotd/cmd/api/types"
	"sort"
	"time"
)

// sortNodeData sorts node data based on the given field and order
func sortNodeData(nodeData []types.NodeData, sortBy, sortOrder string) {
	switch sortBy {
	case "id":
		if sortOrder == "desc" {
			sort.Slice(nodeData, func(i, j int) bool { return nodeData[i].ID > nodeData[j].ID })
		} else {
			sort.Slice(nodeData, func(i, j int) bool { return nodeData[i].ID < nodeData[j].ID })
		}
	case "user_id":
		if sortOrder == "desc" {
			sort.Slice(nodeData, func(i, j int) bool { return nodeData[i].UserID > nodeData[j].UserID })
		} else {
			sort.Slice(nodeData, func(i, j int) bool { return nodeData[i].UserID < nodeData[j].UserID })
		}
	case "device_id":
		if sortOrder == "desc" {
			sort.Slice(nodeData, func(i, j int) bool { return nodeData[i].DeviceID > nodeData[j].DeviceID })
		} else {
			sort.Slice(nodeData, func(i, j int) bool { return nodeData[i].DeviceID < nodeData[j].DeviceID })
		}
	case "moisture_content":
		if sortOrder == "desc" {
			sort.Slice(nodeData, func(i, j int) bool {
				if nodeData[i].MoistureContent == nil && nodeData[j].MoistureContent == nil {
					return false
				}
				if nodeData[i].MoistureContent == nil {
					return false
				}
				if nodeData[j].MoistureContent == nil {
					return true
				}
				return *nodeData[i].MoistureContent > *nodeData[j].MoistureContent
			})
		} else {
			sort.Slice(nodeData, func(i, j int) bool {
				if nodeData[i].MoistureContent == nil && nodeData[j].MoistureContent == nil {
					return false
				}
				if nodeData[i].MoistureContent == nil {
					return true
				}
				if nodeData[j].MoistureContent == nil {
					return false
				}
				return *nodeData[i].MoistureContent < *nodeData[j].MoistureContent
			})
		}
	case "timestamp":
		if sortOrder == "desc" {
			sort.Slice(nodeData, func(i, j int) bool { return nodeData[i].Timestamp.After(nodeData[j].Timestamp) })
		} else {
			sort.Slice(nodeData, func(i, j int) bool { return nodeData[i].Timestamp.Before(nodeData[j].Timestamp) })
		}
	}
}

// GetNodeData fetches all node data from the database
func (db *Database) GetNodeData() ([]types.NodeData, error) {
	return db.GetNodeDataWithPagination(0, 0, "", "")
}

// GetNodeDataWithPagination fetches node data from the database with pagination and sorting
func (db *Database) GetNodeDataWithPagination(limit, offset int, sortBy, sortOrder string) ([]types.NodeData, error) {
	switch db.dbType {
	case InMemory:
		nodeData := make([]types.NodeData, len(InMemoryNodeData))
		copy(nodeData, InMemoryNodeData)

		// Apply sorting
		if sortBy != "" {
			sortNodeData(nodeData, sortBy, sortOrder)
		}

		// Apply pagination
		if limit > 0 {
			start := offset
			if start > len(nodeData) {
				return []types.NodeData{}, nil
			}
			end := start + limit
			if end > len(nodeData) {
				end = len(nodeData)
			}
			return nodeData[start:end], nil
		}
		return nodeData, nil
	case Postgres:
		// Build the query with sorting and pagination
		query := `
			SELECT id, user_id, device_id, moisture_content, timestamp
			FROM "NodeData"
		`

		// Add ORDER BY clause
		if sortBy != "" {
			orderBy := "timestamp" // default
			switch sortBy {
			case "id":
				orderBy = "id"
			case "user_id":
				orderBy = "user_id"
			case "device_id":
				orderBy = "device_id"
			case "moisture_content":
				orderBy = "moisture_content"
			case "timestamp":
				orderBy = "timestamp"
			}

			order := "ASC"
			if sortOrder == "desc" {
				order = "DESC"
			}
			query += fmt.Sprintf(" ORDER BY %s %s", orderBy, order)
		} else {
			query += " ORDER BY timestamp DESC"
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

		var nodeData []types.NodeData
		for rows.Next() {
			var nd types.NodeData
			if err := rows.Scan(&nd.ID, &nd.UserID, &nd.DeviceID, &nd.MoistureContent, &nd.Timestamp); err != nil {
				return nil, err
			}
			nodeData = append(nodeData, nd)
		}
		return nodeData, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// CreateNodeData creates new sensor data in the database
func (db *Database) CreateNodeData(nodeData types.NodeData) error {
	switch db.dbType {
	case InMemory:
		// Find the last node data ID and assign the next ID
		lastID := 0
		for _, nd := range InMemoryNodeData {
			if nd.ID > lastID {
				lastID = nd.ID
			}
		}
		nodeData.ID = lastID + 1
		nodeData.Timestamp = time.Now()
		InMemoryNodeData = append(InMemoryNodeData, nodeData)
		return nil
	case Postgres:
		// Create node data in Postgres database
		query := `
			INSERT INTO "NodeData" (user_id, device_id, moisture_content)
			VALUES ($1, $2, $3)
			RETURNING id, timestamp
		`
		args := []any{nodeData.UserID, nodeData.DeviceID, nodeData.MoistureContent}
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		err := db.context.QueryRowContext(ctx, query, args...).Scan(&nodeData.ID, &nodeData.Timestamp)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

// GetNodeDataByID fetches specific node data by ID from the database
func (db *Database) GetNodeDataByID(id int) (*types.NodeData, error) {
	switch db.dbType {
	case InMemory:
		for _, nodeData := range InMemoryNodeData {
			if nodeData.ID == id {
				return &nodeData, nil
			}
		}
		return nil, fmt.Errorf("node data not found")
	case Postgres:
		// Fetch node data by ID from Postgres database
		query := `
			SELECT id, user_id, device_id, moisture_content, timestamp
			FROM "NodeData"
			WHERE id = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		var nd types.NodeData
		err := db.context.QueryRowContext(ctx, query, id).Scan(&nd.ID, &nd.UserID, &nd.DeviceID, &nd.MoistureContent, &nd.Timestamp)
		if err != nil {
			return nil, err
		}
		return &nd, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// GetNodeDataByDeviceID fetches all sensor data for a specific device
func (db *Database) GetNodeDataByDeviceID(deviceID int) ([]types.NodeData, error) {
	switch db.dbType {
	case InMemory:
		var result []types.NodeData
		for _, nodeData := range InMemoryNodeData {
			if nodeData.DeviceID == deviceID {
				result = append(result, nodeData)
			}
		}
		return result, nil
	case Postgres:
		query := `
			SELECT id, user_id, device_id, moisture_content, timestamp
			FROM "NodeData"
			WHERE device_id = $1
			ORDER BY timestamp DESC
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		rows, err := db.context.QueryContext(ctx, query, deviceID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var nodeData []types.NodeData
		for rows.Next() {
			var nd types.NodeData
			if err := rows.Scan(&nd.ID, &nd.UserID, &nd.DeviceID, &nd.MoistureContent, &nd.Timestamp); err != nil {
				return nil, err
			}
			nodeData = append(nodeData, nd)
		}
		return nodeData, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// GetNodeDataByUserID fetches all sensor data for a specific user
func (db *Database) GetNodeDataByUserID(userID int) ([]types.NodeData, error) {
	switch db.dbType {
	case InMemory:
		var result []types.NodeData
		for _, nodeData := range InMemoryNodeData {
			if nodeData.UserID == userID {
				result = append(result, nodeData)
			}
		}
		return result, nil
	case Postgres:
		query := `
			SELECT id, user_id, device_id, moisture_content, timestamp
			FROM "NodeData"
			WHERE user_id = $1
			ORDER BY timestamp DESC
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		rows, err := db.context.QueryContext(ctx, query, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var nodeData []types.NodeData
		for rows.Next() {
			var nd types.NodeData
			if err := rows.Scan(&nd.ID, &nd.UserID, &nd.DeviceID, &nd.MoistureContent, &nd.Timestamp); err != nil {
				return nil, err
			}
			nodeData = append(nodeData, nd)
		}
		return nodeData, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// DeleteNodeData removes node data from the database
func (db *Database) DeleteNodeData(id int) error {
	switch db.dbType {
	case InMemory:
		for i, nodeData := range InMemoryNodeData {
			if nodeData.ID == id {
				InMemoryNodeData = append(InMemoryNodeData[:i], InMemoryNodeData[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("node data not found")
	case Postgres:
		// Delete node data from Postgres database
		query := `
			DELETE FROM "NodeData"
			WHERE id = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		result, err := db.context.ExecContext(ctx, query, id)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rowsAffected == 0 {
			return fmt.Errorf("node data not found")
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

// ValidateNodeData validates node data before database operations
func ValidateNodeData(nodeData types.NodeData) error {
	if nodeData.UserID <= 0 {
		return fmt.Errorf("invalid user_id: must be positive")
	}
	if nodeData.DeviceID <= 0 {
		return fmt.Errorf("invalid device_id: must be positive")
	}
	if nodeData.MoistureContent != nil && (*nodeData.MoistureContent < 0 || *nodeData.MoistureContent > 100) {
		return fmt.Errorf("invalid moisture_content: must be between 0 and 100")
	}
	return nil
}
