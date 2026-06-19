package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/sbedford/agentic-caddie/internal/agent"
	"github.com/sbedford/agentic-caddie/internal/config"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
	"github.com/sbedford/agentic-caddie/internal/services"

	_ "modernc.org/sqlite"
)

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

	previousRounds, err := roundService.GetRoundsByPlayer(*golfer)
	if err != nil {
		fmt.Errorf("GetRoundsByPlayer failed: %w", err)
		return
	}

	var currentRoundId int64 = 32
	var nextHoleNumber int64 = 2

	currentRound, err := roundService.GetRoundById(currentRoundId)
	nextHole := currentRound.Tee.GetHole(nextHoleNumber)

	fmt.Printf("Tee [%v]", currentRound.Tee.Name)

	req := agent.GetAdviceRequest{
		Queries:      queries,
		Player:       *golfer,
		Rounds:       previousRounds,
		CurrentRound: *currentRound,
		ScopeForAdvice: models.PlayedHole{
			Hole:         *nextHole,
			FlagPosition: "front-left",
		},
	}

	result, err := agent.GetAdvice(context.Background(), req)
	if err != nil {
		log.Fatalf("agent error: %v", err)
	}
	log.Println(result)
}
