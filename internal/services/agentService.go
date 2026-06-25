package services

import (
	"context"

	"github.com/sbedford/agentic-caddie/internal/agent"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
)

type AgentService struct {
	q   *db.Queries
	ctx context.Context
}

func GetAgentService(context context.Context, db *db.Queries) AgentService {
	return AgentService{
		q:   db,
		ctx: context,
	}
}

func (this *AgentService) GetRecommendation(hole models.PlayedHole, distanceFromPin int64) (*agent.GetHoleStrategyResponse, error) {

	roundService := GetRoundsService(context.Background(), this.q)

	previousRounds, err := roundService.GetRoundsByPlayer(hole.Round.Golfer)
	if err != nil {
		return nil, err
	}

	location := hole.CurrentLocation()
	var lastValidShot *models.Shot = nil
	var miss *models.ShotResult = nil

	if location != models.LocationTee {
		lastValidShot = hole.GetLastValidShot()
	}

	if lastValidShot != nil {
		miss = &lastValidShot.Miss
	}

	req := agent.GetHoleStrategyRequest{
		Queries:      this.q,
		Player:       *&hole.Round.Golfer,
		Rounds:       previousRounds,
		CurrentRound: *hole.Round,
		ScopeForAdvice: agent.CurrentSituation{
			HoleNumber:            hole.Hole.HoleNumber,
			ShotNumber:            len(hole.ShotsTaken) + 1,
			CurrentLocation:       location,
			DistanceToThePin:      distanceFromPin,
			LastShotMissDirection: miss,
			Flag:                  &hole.FlagPosition,
		},
	}

	return agent.GetHoleStrategy(context.Background(), req)
}
