package models

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"os"
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
			conn, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
			if err != nil {
				log.Fatalf("Unable to connect to database: %v\n", err)
			} else {
				fmt.Println("Connected to database")
			}

			// Assign connection to global DB variable before using it
			DB = conn

			// Ping the database to ensure the connection is valid
			if err := DB.Ping(context.Background()); err != nil {
				log.Fatalf("Unable to ping database: %v\n", err)
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
