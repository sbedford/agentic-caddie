package services

import (
	"context"
	"fmt"

	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/helpers"
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
			output.PlayedHoles[j] = models.ConvertPlayedHole(playedHole, *teesPlayed.GetHole(playedHole.HoleNumber), output)
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
				resp[i].PlayedHoles[j] = models.ConvertPlayedHole(playedHole, *teesPlayed.GetHole(playedHole.HoleNumber), resp[i])
			}
		}
	}

	return resp, nil
}

func (ps RoundsService) PersistRound(round *models.Round) error {

	if round.ID == 0 {

		result, err := ps.q.CreateRound(ps.ctx, db.CreateRoundParams{
			PlayerID:        round.Golfer.ID,
			CourseID:        round.Course.ID,
			PlayedAt:        round.RoundDate,
			Tees:            round.Tee.Name,
			DailyHandicap:   round.DailyHandicap,
			RoundType:       string(round.RoundType),
			CompetitionType: helpers.ToNullString(string(round.CompetitionType)),
			TotalScore:      helpers.ToNullInt64(round.TotalScore),
			TotalPoints:     helpers.ToNullInt64(round.TotalPoints),
			TotalPutts:      helpers.ToNullInt64(round.TotalPutts),
			Completed:       round.RoundCompleted,
		})

		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		round.ID = id
	} else {
		err := ps.q.UpdateRound(ps.ctx, db.UpdateRoundParams{
			ID:          round.ID,
			TotalScore:  helpers.ToNullInt64(round.TotalScore),
			TotalPoints: helpers.ToNullInt64(round.TotalPoints),
			TotalPutts:  helpers.ToNullInt64(round.TotalPutts),
			Completed:   round.RoundCompleted,
		})
		if err != nil {
			return err
		}
	}

	for _, hole := range round.PlayedHoles {
		err := ps.PersistHole(&hole)
		if err != nil {
			return err
		}
	}

	return nil
}

func (ps RoundsService) PersistHole(ph *models.PlayedHole) error {
	if ph.ID == 0 {
		result, err := ps.q.CreateHole(ps.ctx, db.CreateHoleParams{
			RoundID:        ph.Round.ID,
			CourseHoleID:   ph.Hole.CourseHoleID,
			HoleNumber:     ph.Hole.HoleNumber,
			FlagPosition:   helpers.ToNullString(string(ph.FlagPosition)),
			Score:          helpers.ToNullInt64(ph.Score),
			Points:         helpers.ToNullInt64(ph.Points),
			Putts:          helpers.ToNullInt64(ph.NumberOfPutts),
			Gir:            helpers.ToNullBool(ph.GreenInRegulation),
			ScrambleSave:   helpers.ToNullBool(ph.ScrambleSave),
			Penalty:        helpers.ToNullBool(ph.Penalty),
			PenaltyStrokes: helpers.ToNullInt64(ph.PenaltyStrokes),
			Wiped:          helpers.ToNullBool(ph.Wiped),
			Completed:      helpers.ToNullBool(ph.Completed),
		})
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		ph.ID = id
	} else {
		err := ps.q.UpdateHole(ps.ctx, db.UpdateHoleParams{
			FlagPosition:   helpers.ToNullString(string(ph.FlagPosition)),
			Score:          helpers.ToNullInt64(ph.Score),
			Points:         helpers.ToNullInt64(ph.Points),
			Putts:          helpers.ToNullInt64(ph.NumberOfPutts),
			Gir:            helpers.ToNullBool(ph.GreenInRegulation),
			ScrambleSave:   helpers.ToNullBool(ph.ScrambleSave),
			Penalty:        helpers.ToNullBool(ph.Penalty),
			PenaltyStrokes: helpers.ToNullInt64(ph.PenaltyStrokes),
			Wiped:          helpers.ToNullBool(ph.Wiped),
			Completed:      helpers.ToNullBool(ph.Completed),
			ID:             ph.ID,
		})
		if err != nil {
			return err
		}
	}

	for _, shot := range ph.ShotsTaken {
		err := ps.PersistShot(&shot)
		if err != nil {
			return err
		}
	}
	return nil

}

func (ps RoundsService) PersistShot(sh *models.Shot) error {
	if sh.ID == 0 {
		result, err := ps.q.CreateShot(ps.ctx, db.CreateShotParams{
			HoleID:                sh.Hole.ID,
			ShotNumber:            sh.ShotNumber,
			ShotType:              string(sh.ShotType),
			Club:                  helpers.ToNullString(string(sh.Club)),
			Result:                helpers.ToNullString(string(sh.Result)),
			Miss:                  helpers.ToNullString(string(sh.Miss)),
			StrikeQuality:         helpers.ToNullString(string(sh.StrikeQuality)),
			Source:                "",
			PreShotRecommendation: helpers.ToNullString(sh.PreShotRecommendation),
			Completed:             helpers.ToNullBool(sh.Completed),
		})
		if err != nil {
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			return err
		}
		sh.ID = id
	} else {
		err := ps.q.UpdateShot(ps.ctx, db.UpdateShotParams{
			ShotType:              string(sh.ShotType),
			Club:                  helpers.ToNullString(string(sh.Club)),
			Result:                helpers.ToNullString(string(sh.Result)),
			Miss:                  helpers.ToNullString(string(sh.Miss)),
			StrikeQuality:         helpers.ToNullString(string(sh.StrikeQuality)),
			Source:                "",
			PreShotRecommendation: helpers.ToNullString(sh.PreShotRecommendation),
			Completed:             helpers.ToNullBool(sh.Completed),
			ID:                    sh.ID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
