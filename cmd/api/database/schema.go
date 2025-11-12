package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"qotd/cmd/api/types"
	"time"

	_ "github.com/lib/pq"
)

var InMemoryQuotes []types.Quote
var InMemoryComments []types.Comment

type DatabaseType int

const (
	InMemory DatabaseType = iota
	Postgres
)

type Database struct {
	connectionString string
	dbType           DatabaseType
	context          *sql.DB
	queryTimeout     time.Duration
}

// Create a logger
var logger = slog.New(slog.NewTextHandler(os.Stderr, nil))

func NewDatabase(dbType DatabaseType, connectionString *string) *Database {
	// @completed Database connection logic
	if connectionString == nil {
		connectionString = new(string)
		*connectionString = ""
	}
	return &Database{
		connectionString: *connectionString,
		dbType:           dbType,
		queryTimeout:     3 * time.Second,
	}
}

func (ctx *Database) Connect() error {
	// @todo Implement database connection logic
	switch ctx.dbType {
	case InMemory:
		// Connect to in-memory database
		return nil
	case Postgres:
		// Connect to Postgres database
		db, err := openDB(ctx.connectionString)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}
		// Assign the context
		ctx.context = db

		logger.Info("database connection pool established")

		return nil
	}
	return fmt.Errorf("unsupported database type")
}

func (db *Database) Disconnect() error {
	// todo-complete Implement database disconnection logic
	switch db.dbType {
	case InMemory:
		// Disconnect from in-memory database
		// Flush the in-memory quotes
		InMemoryQuotes = nil
		return nil
	case Postgres:
		// Disconnect from Postgres database
		db.context.Close()
		return nil
	}
	return fmt.Errorf("unsupported database type")
}
