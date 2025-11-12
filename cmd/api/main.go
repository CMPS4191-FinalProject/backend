package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"qotd/cmd/api/database"
	"syscall"
	"time"

	"github.com/julienschmidt/httprouter"
)

type serverConfig struct {
	port    int
	env     string
	logger  *slog.Logger
	version string
	db      *database.Database
	router  *httprouter.Router
}

func main() {
	// Initialize the server configuration
	config := serverConfig{
		port:    getEnvAsInt("PORT", 8080),
		env:     getEnvAsString("ENVIRONMENT", "development"),
		version: getEnvAsString("API_VERSION", "v1"),
		router:  httprouter.New(),
	}

	dbDsn := getEnvAsString("DB_DSN", "")
	dbType := getEnvAsString("DB_TYPE", "IN_MEMORY")

	// Read in the database type
	flag.StringVar(&dbType, "db-type", dbType, "Database type (IN_MEMORY or POSTGRES)")

	if dbType == "POSTGRES" {
		flag.StringVar(&dbDsn, "db-dsn", dbDsn, "PostgreSQL DSN")
	}

	fmt.Println(dbDsn)
	fmt.Println(dbType)

	flag.IntVar(&config.port, "port", config.port, "API server port")

	if dbType != "IN_MEMORY" && dbType != "POSTGRES" {
		fmt.Print("Error: Unsupported database type. Use IN_MEMORY or POSTGRES.\n")
		os.Exit(1)
	}

	if dbType == "POSTGRES" && dbDsn == "" {
		fmt.Print("Error: DSN must be provided for PostgreSQL database type.\n")
		os.Exit(1)
	}

	flag.Parse()
	// Initialize the database connection
	if dbType == "POSTGRES" {
		config.db = database.NewDatabase(database.Postgres, &dbDsn)
	} else {
		config.db = database.NewDatabase(database.InMemory, nil)
	}

	// Create a logger
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config.logger = logger

	fmt.Println("Listening on port " + fmt.Sprint(config.port))
	fmt.Println("Environment: " + config.env)
	router := config.routes()
	config.db.Connect()

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + fmt.Sprint(config.port),
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Channel to listen for interrupt signal
	shutdownError := make(chan error)

	// Start a goroutine to listen for interrupt signals
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		config.logger.Info("shutting down server", "signal", s.String())

		// Create a context with a timeout for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		// Attempt graceful shutdown
		err := server.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
		}

		config.logger.Info("completing background tasks", "addr", server.Addr)

		// Close database connection
		config.db.Disconnect()

		config.logger.Info("stopped server", "addr", server.Addr)
		close(shutdownError)
	}()

	config.logger.Info("starting server", "addr", server.Addr, "env", config.env)

	// Start the server
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		config.logger.Error("server error", "error", err)
		os.Exit(1)
	}

	// Wait for shutdown to complete or timeout
	err = <-shutdownError
	if err != nil {
		config.logger.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	config.logger.Info("stopped server", "addr", server.Addr)
}
