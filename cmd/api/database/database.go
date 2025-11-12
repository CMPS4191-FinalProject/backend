package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func openDB(dsn string) (*sql.DB, error) {
	fmt.Println("Opening database with DSN:", dsn)
	// open a connection pool
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// set a context to ensure DB operations don't take too long
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// let's test if the connection pool was created
	// we trying pinging it with a 5-second timeout
	err = sqlDB.PingContext(ctx)
	if err != nil {
		sqlDB.Close()
		return nil, err
	}

	// return the connection pool (sql.DB)
	return sqlDB, nil

}
