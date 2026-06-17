package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sbedford/agentic-caddie/internal/agent"
	"github.com/sbedford/agentic-caddie/internal/config"
	"github.com/sbedford/agentic-caddie/internal/db"

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
	p, err := queries.GetPlayerByID(context.Background(), 1)
	if err != nil {
		log.Println("error getting player")
		return
	}

	c, err := queries.ListActiveClubsByPlayer(context.Background(), 1)
	if err != nil {
		log.Println("error getting clubs")
		return
	}
	r, err := queries.ListRoundsByPlayer(context.Background(), p.ID)
	if err != nil {
		log.Println("error getting rounds")
		return
	}

	req := agent.GetAdviceRequest{
		Queries:      queries,
		Player:       p,
		Clubs:        c,
		Rounds:       r,
		CurrentRound: r[len(r)-1],
	}

	out, err := json.Marshal(req)
	if err != nil {
		fmt.Errorf("marshal failed: %w", err)
	}
	log.Printf(string(out))

	/*result, err := agent.GetAdvice(context.Background(), req)
	if err != nil {
		log.Fatalf("agent error: %v", err)
	}
	log.Println(result)
	*/
}
