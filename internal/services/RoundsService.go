package services

import (
	"context"
	"fmt"

	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
)

type RoundsService struct {
	q   *db.Queries
	ctx context.Context
}

func GetRoundsService(context context.Context, db *db.Queries) RoundsService {
	return RoundsService{
		q:   db,
		ctx: context,
	}
}

// Loads all data on a round - down to the shot level
func (ps RoundsService) GetRoundById(roundId int64) (*models.Round, error) {

	round, err := ps.q.GetRoundByID(ps.ctx, roundId)

	if err != nil {
		fmt.Errorf("Error in GetRoundById(%V)", roundId)
		return nil, err
	}

	courseService := GetCourseService(ps.ctx, ps.q)
	course, err := courseService.GetCourse(round.CourseID)
	if err != nil {
		fmt.Errorf("Error in courseService.GetCourse(%V)", round.CourseID)
		return nil, err
	}

	playerService := GetPlayerService(ps.ctx, ps.q)
	player, err := playerService.GetPlayer(round.PlayerID)
	if err != nil {
		fmt.Errorf("Error in courseService.GetCourse(%V)", round.CourseID)
		return nil, err
	}

	teesPlayed := *course.FindTee(round.Tees)

	output := models.ConvertRound(round, *course, *player, teesPlayed)

	playedHoles, err := ps.q.ListHolesByRound(ps.ctx, roundId)
	//shotsTaken, err := ps.q.GetShotsByRound(ps.ctx, roundId)

	output.PlayedHoles = make([]models.PlayedHole, len(playedHoles))
	for j, playedHole := range playedHoles {
		if playedHole.HoleNumber <= 9 { // dirty hack
			output.PlayedHoles[j] = models.ConvertPlayedHole(playedHole, *teesPlayed.GetHole(playedHole.HoleNumber))
		}
	}

	return &output, nil

}

func (ps RoundsService) GetRoundsByPlayerId(playerId int64) ([]models.Round, error) {

	playerService := GetPlayerService(ps.ctx, ps.q)

	player, err := playerService.GetPlayer(playerId)

	if err != nil {
		fmt.Errorf("Error in GetRounds / GetPlayer(%V)", playerId)
		return nil, err
	}

	return ps.GetRoundsByPlayer(*player)
}

func (ps RoundsService) GetRoundsByPlayer(player models.Player) ([]models.Round, error) {

	savedRounds, err := ps.q.ListCompletedRoundsByPlayer(ps.ctx, player.ID)
	if err != nil {
		fmt.Errorf("Error in ListRoundsByPlayer(%V)", player.ID)
		return nil, err
	}

	courseService := GetCourseService(ps.ctx, ps.q)

	courses := make(map[int64]models.Course)

	resp := make([]models.Round, len(savedRounds))
	for i, round := range savedRounds {

		playedHoles, err := ps.q.ListHolesByRound(ps.ctx, round.ID)
		if err != nil {
			fmt.Errorf("Error in ListHolesByRound(%V)", round.ID)
			return nil, err
		}

		course, ok := courses[round.CourseID]
		if !ok {
			cPtr, err := courseService.GetCourse(round.CourseID)
			if err != nil || cPtr == nil {
				fmt.Errorf("Error in courseService.GetCourse(%V)", round.CourseID)
				return nil, err
			}
			courses[round.CourseID] = *cPtr
			course = *cPtr
		}

		teesPlayed := *course.FindTee(round.Tees)
		resp[i] = models.ConvertRound(round, course, player, teesPlayed)

		//resp[i].PlayedHoles = make([]models.PlayedHole, len(playedHoles))
		for j, playedHole := range playedHoles {
			if playedHole.HoleNumber <= 9 { // dirty hack
				resp[i].PlayedHoles[j] = models.ConvertPlayedHole(playedHole, *teesPlayed.GetHole(playedHole.HoleNumber))
			}
		}
	}

	return resp, nil
}
