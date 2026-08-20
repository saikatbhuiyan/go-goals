package postgres

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func ConnectFromEnv() *sql.DB {
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		log.Fatal("POSTGRES_URL environment variable must be set")
	}

	db, err := sql.Open("postgres", postgresURL)
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to the database")
	return db
}
