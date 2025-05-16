package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"strings"
	"sync"
)

var (
	DB   *pgxpool.Pool
	once sync.Once
)

func DBConnect() {
	once.Do(
		func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Get a connection string and log it (without a password)
			dbURL := os.Getenv("DATABASE_URL")
			dbURLRedacted := redactPasswordFromConnString(dbURL)
			log.Printf("Connecting to database: %s\n", dbURLRedacted)

			// Ensure the host parameter is present for TCP connection
			if !strings.Contains(dbURL, "host=") && strings.HasPrefix(dbURL, "postgres://") {
				log.Println("Connection string appears to be in URL format, proceeding normally")
			}

			conn, err := pgxpool.New(ctx, dbURL)
			if err != nil {
				log.Fatalf("Unable to connect to database: %v\n", err)
			} else {
				fmt.Println("Connected to database")
			}

			// Assign connection to global DB variable before using it
			DB = conn

			// Ping the database to ensure the connection is valid
			if err := DB.Ping(context.Background()); err != nil {
				log.Printf("Unable to ping database: %v\n", err)
			} else {
				fmt.Println("Successfully pinged database")
			}
			// Create tables after a connection is established and assigned to DB
			if err := createTables(); err != nil {
				log.Fatalf("Unable to create tables: %v\n", err)
			}
		},
	)
}

// Helper function to redact password from a connection string for logging
func redactPasswordFromConnString(connString string) string {
	if connString == "" {
		return ""
	}

	// Handle URL format
	if strings.HasPrefix(connString, "postgres://") {
		parts := strings.Split(connString, "@")
		if len(parts) >= 2 {
			userPart := strings.Split(parts[0], ":")
			if len(userPart) >= 2 {
				return userPart[0] + ":****@" + strings.Join(parts[1:], "@")
			}
		}
	}

	// Handle key=value format
	if strings.Contains(connString, "password=") {
		result := connString
		passwordStart := strings.Index(result, "password=")
		if passwordStart != -1 {
			passwordEnd := strings.Index(result[passwordStart:], " ")
			if passwordEnd == -1 {
				// Password is at the end of the string
				result = result[:passwordStart] + "password=****"
			} else {
				// Password is in the middle of the string
				result = result[:passwordStart] + "password=****" + result[passwordStart+passwordEnd:]
			}
		}
		return result
	}

	// If we can't identify the format, just return a placeholder
	return "CONNECTION_STRING_REDACTED"
}
