package database

import (
	"context"
	"fmt"
	"qotd/cmd/api/types"
	"sort"
	"time"
)

// sortComments sorts comments based on the given field and order
func sortComments(comments []types.Comment, sortBy, sortOrder string) {
	switch sortBy {
	case "id":
		if sortOrder == "desc" {
			sort.Slice(comments, func(i, j int) bool { return comments[i].ID > comments[j].ID })
		} else {
			sort.Slice(comments, func(i, j int) bool { return comments[i].ID < comments[j].ID })
		}
	case "author":
		if sortOrder == "desc" {
			sort.Slice(comments, func(i, j int) bool { return comments[i].Author > comments[j].Author })
		} else {
			sort.Slice(comments, func(i, j int) bool { return comments[i].Author < comments[j].Author })
		}
	case "content":
		if sortOrder == "desc" {
			sort.Slice(comments, func(i, j int) bool { return comments[i].Content > comments[j].Content })
		} else {
			sort.Slice(comments, func(i, j int) bool { return comments[i].Content < comments[j].Content })
		}
	case "created_at":
		if sortOrder == "desc" {
			sort.Slice(comments, func(i, j int) bool { return comments[i].CreatedAt.After(comments[j].CreatedAt) })
		} else {
			sort.Slice(comments, func(i, j int) bool { return comments[i].CreatedAt.Before(comments[j].CreatedAt) })
		}
	case "version":
		if sortOrder == "desc" {
			sort.Slice(comments, func(i, j int) bool { return comments[i].Version > comments[j].Version })
		} else {
			sort.Slice(comments, func(i, j int) bool { return comments[i].Version < comments[j].Version })
		}
	}
}

func (db *Database) GetComments() ([]types.Comment, error) {
	return db.GetCommentsWithPagination(0, 0, "", "")
}

// GetCommentsWithPagination fetches comments from the database with pagination and sorting
func (db *Database) GetCommentsWithPagination(limit, offset int, sortBy, sortOrder string) ([]types.Comment, error) {
	switch db.dbType {
	case InMemory:
		comments := make([]types.Comment, len(InMemoryComments))
		copy(comments, InMemoryComments)

		// Apply sorting
		if sortBy != "" {
			sortComments(comments, sortBy, sortOrder)
		}

		// Apply pagination
		if limit > 0 {
			start := offset
			if start > len(comments) {
				return []types.Comment{}, nil
			}
			end := start + limit
			if end > len(comments) {
				end = len(comments)
			}
			return comments[start:end], nil
		}
		return comments, nil
	case Postgres:
		// Build the query with sorting and pagination
		query := `
			SELECT id, content, author, created_at, version
			FROM comments
		`

		// Add ORDER BY clause
		if sortBy != "" {
			orderBy := "created_at" // default
			switch sortBy {
			case "id":
				orderBy = "id"
			case "author":
				orderBy = "author"
			case "content":
				orderBy = "content"
			case "created_at":
				orderBy = "created_at"
			case "version":
				orderBy = "version"
			}

			order := "ASC"
			if sortOrder == "desc" {
				order = "DESC"
			}
			query += fmt.Sprintf(" ORDER BY %s %s", orderBy, order)
		} else {
			query += " ORDER BY created_at DESC"
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

		var comments []types.Comment
		for rows.Next() {
			var c types.Comment
			if err := rows.Scan(&c.ID, &c.Content, &c.Author, &c.CreatedAt, &c.Version); err != nil {
				return nil, err
			}
			comments = append(comments, c)
		}
		return comments, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

func (db *Database) WriteComment(comment types.Comment) error {
	// @completed Implement writing comments to the database
	/*
	 * Already completed
	 * ===============
	 * 1. In-memory database writing
	 * 2. PostgreSQL writing
	 */
	switch db.dbType {
	case InMemory:
		// Find the last comment's ID and assign the next ID
		lastID := int(0)
		for _, c := range InMemoryComments {
			if c.ID > lastID {
				lastID = c.ID
			}
		}
		comment.ID = lastID + 1
		InMemoryComments = append(InMemoryComments, comment)
		return nil
	case Postgres:
		// Write comment to Postgres database
		query := `
			INSERT INTO comments (content, author)
			VALUES ($1, $2)
			RETURNING id, created_at, version
		`
		args := []any{comment.Content, comment.Author}
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		err := db.context.QueryRowContext(ctx, query, args...).Scan(&comment.ID, &comment.CreatedAt, &comment.Version)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

func (db *Database) GetCommentByID(id int) (*types.Comment, error) {
	// @completed Implement fetching a single comment by ID from the database
	switch db.dbType {
	case InMemory:
		for _, comment := range InMemoryComments {
			if comment.ID == id {
				return &comment, nil
			}
		}
		return nil, fmt.Errorf("comment not found")
	case Postgres:
		// Fetch comment by ID from Postgres database
		query := `
			SELECT id, content, author, created_at, version
			FROM comments
			WHERE id = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		var c types.Comment
		err := db.context.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Content, &c.Author, &c.CreatedAt, &c.Version)
		if err != nil {
			return nil, err
		}
		return &c, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

func (db *Database) ModifyComment(commentID int, comment types.Comment) error {
	// @completed Implement modifying a comment in the database
	switch db.dbType {
	case InMemory:
		for i, c := range InMemoryComments {
			if c.ID == commentID {
				InMemoryComments[i] = comment
				return nil
			}
		}
		return fmt.Errorf("comment not found")
	case Postgres:
		// Modify comment in Postgres database
		query := `
			UPDATE comments
			SET content = $1, author = $2, created_at = $3, version = version + 1
			WHERE id = $4
			RETURNING version
		`
		args := []any{comment.Content, comment.Author, time.Now(), commentID}
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		err := db.context.QueryRowContext(ctx, query, args...).Scan(&comment.Version)
		if err != nil {
			return err
		}

		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

func ValidateComment(comment types.Comment) error {
	if comment.Author == "" {
		return fmt.Errorf("Field 'Author' missing")
	}

	if comment.Content == "" {
		return fmt.Errorf("Field 'Content' missing")
	}
	return nil
}

func (db *Database) DeleteComment(id int) error {
	// @completed Implement deleting a comment from the database
	switch db.dbType {
	case InMemory:
		for i, c := range InMemoryComments {
			if c.ID == id {
				InMemoryComments = append(InMemoryComments[:i], InMemoryComments[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("comment not found")
	case Postgres:
		// Delete comment from Postgres database
		query := `
			DELETE FROM comments
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
			return fmt.Errorf("comment not found")
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}
