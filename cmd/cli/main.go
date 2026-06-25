package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	cli "github.com/sbedford/agentic-caddie/internal/cli"

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

	file, err := os.OpenFile("data/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close() // Ensure the file handles close properly when main exits

	// Redirect all standard log package outputs to our file
	log.SetOutput(file)

	err = cli.RenderForm(context.Background(), queries)
	if err != nil {
		log.Printf("Got Error - %v", err.Error())
	}
}
