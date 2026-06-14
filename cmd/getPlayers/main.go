package main

import (
	"database/sql"
	"log"
	"time"

	"context"

	"github.com/sbedford/agentic-caddie/internal/config"
	"github.com/sbedford/agentic-caddie/internal/db"

	_ "modernc.org/sqlite"
)

var database *sql.DB

func main() {
	cfg := config.Load()

	var err error
	database, err = sql.Open("sqlite", cfg.DSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	queries := db.New(database)
	players, err := queries.ListPlayers(ctx)
	if err != nil {
		log.Fatalf("failed to list players: %v", err)
	}
	log.Println(players)
}
