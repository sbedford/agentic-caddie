package services

import (
	"context"
	"log"

	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
)

type PlayerService struct {
	q   *db.Queries
	ctx context.Context
}

func GetPlayerService(context context.Context, db *db.Queries) PlayerService {
	return PlayerService{
		q:   db,
		ctx: context,
	}
}

func (ps PlayerService) GetPlayer(playerId int64) (*models.Player, error) {
	player := models.Player{}

	p, err := ps.q.GetPlayerByID(ps.ctx, playerId)
	if err != nil {
		log.Printf("error getting player")
		return nil, err
	}

	player.ID = p.ID
	player.Name = p.Name
	player.Handicap = p.Handicap.Float64

	clubs, err := ps.q.ListActiveClubsByPlayer(ps.ctx, 1)
	if err != nil {
		log.Println("error getting clubs")
		return &player, err
	}

	player.Clubs = make([]models.Club, len(clubs))
	for i, club := range clubs {

		player.Clubs[i] = models.Club{}
		player.Clubs[i].Load(club, player)
	}

	roundsService := GetRoundsService(ps.ctx, ps.q)

	player.Rounds, err = roundsService.GetRoundsByPlayer(player)
	if err != nil {
		log.Println("error getting rounds")
		return &player, err
	}

	return &player, nil
}
