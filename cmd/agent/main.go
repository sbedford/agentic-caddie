package main

import (
	"context"
	"database/sql"
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
		log.Println("LoadPLayer failed: ", err)
		return
	}

	previousRounds, err := roundService.GetRoundsByPlayer(*golfer)
	if err != nil {
		log.Println("GetRoundsByPlayer failed: ", err)
		return
	}

	var currentRoundId int64 = 32
	var nextHoleNumber int64 = 2

	currentRound, err := roundService.GetRoundById(currentRoundId)
	nextHole := currentRound.Tee.GetHole(nextHoleNumber)

	req := agent.GetHoleStrategyRequest{
		Queries:      queries,
		Player:       *golfer,
		Rounds:       previousRounds,
		CurrentRound: *currentRound,
		ScopeForAdvice: models.PlayedHole{
			Hole:         *nextHole,
			FlagPosition: "",
		},
	}

	response, err := agent.GetHoleStrategy(context.Background(), req)
	if err != nil {
		log.Println("agent error: ", err)
		return
	}

	log.Println(response.Strategy)
	log.Println("------------------------")
	log.Println("Total Input Tokens:", response.Usage.TotalInputTokens)
	log.Println("Total Output Tokens:", response.Usage.TotalOutputTokens)
	log.Println("Total Cache Create Tokens:", response.Usage.TotalCacheCreationInputTokens)
	log.Println("Total Cache Read Tokens:", response.Usage.TotalCacheReadInputTokens)
	log.Println("------------------------")
}

func PlayRound() {

}
