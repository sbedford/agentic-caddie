package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/sbedford/agentic-caddie/internal/config"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/services"

	_ "modernc.org/sqlite"
)

func buildContextBlock() string {

	/*cfg := config.Load()
	database, err := sql.Open("sqlite", cfg.DSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	queries := db.New(database)
	player, err := queries.GetPlayerByID(context.Background(), 1)
	if err != nil {
		log.Println("error getting player")
		return
	}
	clubs, err := queries.ListActiveClubsByPlayer(context.Background(), 1)
	if err != nil {
		log.Println("error getting clubs")
		return
	}*/

	return ""
}

func main() {
	log.Println("Starting Agent")

	cfg := config.Load()
	database, err := sql.Open("sqlite", cfg.DSN())
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer database.Close()

	queries := db.New(database)

	playerService := services.GetPlayerService(context.Background(), queries)
	roundService := services.GetRoundsService(context.Background(), queries)

	golfer, err := playerService.GetPlayer(1)
	if err != nil {
		fmt.Errorf("LoadPLayer failed: %w", err)
		return
	}

	rounds, err := roundService.GetRoundsByPlayer(golfer)
	if err != nil {
		fmt.Errorf("GetRoundsByPlayer failed: %w", err)
		return
	}

	/*
			Input:
				HoleId



		req := agent.GetAdviceRequest{
			Queries:      queries,
			Player:       p,
			Clubs:        c,
			Rounds:       r,
			CurrentRound: r[len(r)-1],
		}

	*/

	/*result, err := agent.GetAdvice(context.Background(), req)
	if err != nil {
		log.Fatalf("agent error: %v", err)
	}
	log.Println(result)
	*/
}
