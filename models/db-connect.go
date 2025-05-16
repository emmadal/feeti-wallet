package models

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
	"sync"
	"time"
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
				log.Println("Connected to database")
			}

			// Assign connection to global DB variable before using it
			DB = conn

			// Ping the database to ensure the connection is valid
			if err := DB.Ping(context.Background()); err != nil {
				log.Printf("Unable to ping database: %v\n", err)
			} else {
				log.Println("Pinged database successfully")
			}

			// Create tables after a connection is established and assigned to DB
			if err := createTables(); err != nil {
				log.Printf("Unable to create tables: %v\n", err)
			}
		},
	)
}
