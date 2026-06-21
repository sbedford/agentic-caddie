package main

import (
	cli "github.com/sbedford/agentic-caddie/internal/cli"
	"database/sql"
	"log"
	"context"

	"github.com/sbedford/agentic-caddie/internal/config"
	"github.com/sbedford/agentic-caddie/internal/db"
	_ "modernc.org/sqlite"
)

func main() {
    
	cfg := config.Load()
	database, err := sql.Open("sqlite", cfg.DSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	queries := db.New(database)

	cli.RenderForm(context.Background(), queries)
}