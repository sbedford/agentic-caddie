package main

import (
	"database/sql"
	"log"
	"time"

	"context"
	"os"

	"github.com/sbedford/agentic-caddie/internal/config"

	_ "modernc.org/sqlite"
)

var db *sql.DB

func main() {
	cfg := config.Load()

	var err error
	db, err = sql.Open("sqlite", cfg.DSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := applySchema(ctx); err != nil {
		log.Fatalf("Failed to apply schema: %v", err)
	}

	if err := bootstrapData(ctx); err != nil {
		log.Fatalf("Failed to bootstrap data: %v", err)
	}
}

func applySchema(ctx context.Context) error {

	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='courses'").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("Schema already exists, no action needed")
		return nil
	}

	sqlFile, err := os.ReadFile("data/schema/schema.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(sqlFile))
	if err != nil {
		return err
	}

	return err
}

func bootstrapData(ctx context.Context) error {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM players").Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		log.Println("Data already exists, no action needed")
		return nil
	}

	sqlFile, err := os.ReadFile("data/scripts/data.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(sqlFile))
	if err != nil {
		return err
	}

	return err
}
