package database

import (
	"context"
	"fmt"
	"qotd/cmd/api/types"
	"sort"
)

// sortQuotes sorts quotes based on the given field and order
func sortQuotes(quotes []types.Quote, sortBy, sortOrder string) {
	switch sortBy {
	case "id":
		if sortOrder == "desc" {
			sort.Slice(quotes, func(i, j int) bool { return quotes[i].ID > quotes[j].ID })
		} else {
			sort.Slice(quotes, func(i, j int) bool { return quotes[i].ID < quotes[j].ID })
		}
	case "author":
		if sortOrder == "desc" {
			sort.Slice(quotes, func(i, j int) bool { return quotes[i].Author > quotes[j].Author })
		} else {
			sort.Slice(quotes, func(i, j int) bool { return quotes[i].Author < quotes[j].Author })
		}
	case "text":
		if sortOrder == "desc" {
			sort.Slice(quotes, func(i, j int) bool { return quotes[i].Text > quotes[j].Text })
		} else {
			sort.Slice(quotes, func(i, j int) bool { return quotes[i].Text < quotes[j].Text })
		}
	case "created_at":
		if sortOrder == "desc" {
			sort.Slice(quotes, func(i, j int) bool { return quotes[i].CreatedAt.After(quotes[j].CreatedAt) })
		} else {
			sort.Slice(quotes, func(i, j int) bool { return quotes[i].CreatedAt.Before(quotes[j].CreatedAt) })
		}
	}
}

// Fetching quotes from the database
func (db *Database) GetQuotes() ([]types.Quote, error) {
	return db.GetQuotesWithPagination(0, 0, "", "")
}

// Fetching quotes from the database with pagination and sorting
func (db *Database) GetQuotesWithPagination(limit, offset int, sortBy, sortOrder string) ([]types.Quote, error) {
	switch db.dbType {
	case InMemory:
		quotes := make([]types.Quote, len(InMemoryQuotes))
		copy(quotes, InMemoryQuotes)

		// Apply sorting
		if sortBy != "" {
			sortQuotes(quotes, sortBy, sortOrder)
		}

		// Apply pagination
		if limit > 0 {
			start := offset
			if start > len(quotes) {
				return []types.Quote{}, nil
			}
			end := start + limit
			if end > len(quotes) {
				end = len(quotes)
			}
			return quotes[start:end], nil
		}
		return quotes, nil
	case Postgres:
		// Build the query with sorting and pagination
		query := `
			SELECT id, text, author, created_at
			FROM quotes
		`

		// Add ORDER BY clause
		if sortBy != "" {
			orderBy := "created_at" // default
			switch sortBy {
			case "id":
				orderBy = "id"
			case "author":
				orderBy = "author"
			case "text":
				orderBy = "text"
			case "created_at":
				orderBy = "created_at"
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

		var quotes []types.Quote
		for rows.Next() {
			var q types.Quote
			if err := rows.Scan(&q.ID, &q.Text, &q.Author, &q.CreatedAt); err != nil {
				return nil, err
			}
			quotes = append(quotes, q)
		}
		return quotes, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// Writing quotes to the database
func (db *Database) WriteQuote(quote types.Quote) error {
	/*
	 * Already completed
	 * ===============
	 * 1. In-memory database writing
	 *
	 * Not_completed_Todos
	 * ==============
	 * 1. PostgreSQL writing
	 */
	switch db.dbType {
	case InMemory:
		// Find the last quote's ID and assign the next ID
		lastID := 0
		for _, q := range InMemoryQuotes {
			if q.ID > lastID {
				lastID = q.ID
			}
		}
		quote.ID = lastID + 1
		InMemoryQuotes = append(InMemoryQuotes, quote)
		return nil
	case Postgres:
		// Write quote to Postgres database
		query := `
			INSERT INTO quotes (text, author)
			VALUES ($1, $2)
			RETURNING id, created_at
		`
		args := []any{quote.Text, quote.Author}
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		err := db.context.QueryRowContext(ctx, query, args...).Scan(&quote.ID, &quote.CreatedAt)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

// Fetching a single, specific quote by ID from the database
func (db *Database) GetQuoteByID(id int) (*types.Quote, error) {
	switch db.dbType {
	case InMemory:
		for _, quote := range InMemoryQuotes {
			if quote.ID == id {
				return &quote, nil
			}
		}
		return nil, fmt.Errorf("quote not found")
	case Postgres:
		// Fetch quote by ID from Postgres database
		query := `
			SELECT id, text, author, created_at
			FROM quotes
			WHERE id = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		var q types.Quote
		err := db.context.QueryRowContext(ctx, query, id).Scan(&q.ID, &q.Text, &q.Author, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		return &q, nil
	}
	return nil, fmt.Errorf(DATABASE_UNSUPPORTED)
}

// Modifying a quote in the database
func (db *Database) ModifyQuote(quoteID int, quote types.Quote) error {
	switch db.dbType {
	case InMemory:
		for i, q := range InMemoryQuotes {
			if q.ID == quoteID {
				InMemoryQuotes[i] = quote
				return nil
			}
		}
		return fmt.Errorf("quote not found")
	case Postgres:
		// Modify quote in Postgres database
		query := `
			UPDATE quotes
			SET text = $1, author = $2
			WHERE id = $3
		`
		args := []any{quote.Text, quote.Author, quoteID}
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

// Deleting a quote from the database
func (db *Database) DeleteQuote(id int) error {
	switch db.dbType {
	case InMemory:
		for i, quote := range InMemoryQuotes {
			if quote.ID == id {
				InMemoryQuotes = append(InMemoryQuotes[:i], InMemoryQuotes[i+1:]...)
				return nil
			}
		}
		return fmt.Errorf("quote not found")
	case Postgres:
		// Delete quote from Postgres database
		query := `

			DELETE FROM quotes
			WHERE id = $1
		`
		ctx, cancel := context.WithTimeout(context.Background(), db.queryTimeout)
		defer cancel()

		_, err := db.context.ExecContext(ctx, query, id)
		if err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf(DATABASE_UNSUPPORTED)
}

func ValidateQuote(quote types.Quote) error {
	if quote.Author == "" {
		return fmt.Errorf("Field 'Author' missing")
	}

	if quote.Text == "" {
		return fmt.Errorf("Field 'Text' missing")
	}
	return nil
}
