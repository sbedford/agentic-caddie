package services

import (
	"context"
	"fmt"
	"log"
	"strconv"

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
	shotsTaken, err := ps.q.ListShotsByRound(ps.ctx, roundId)

	output.PlayedHoles = make([]models.PlayedHole, len(playedHoles))
	for j, playedHole := range playedHoles {
		if th := teesPlayed.GetHole(playedHole.HoleNumber); th != nil {
			output.PlayedHoles[j] = models.ConvertPlayedHole(playedHole, *th, &output, shotsTaken)
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

func (ps RoundsService) populateRounds(player models.Player, savedRounds []db.Round) ([]models.Round, error) {
	courseService := GetCourseService(ps.ctx, ps.q)

	courses := make(map[int64]models.Course)

	resp := make([]models.Round, len(savedRounds))
	for i, round := range savedRounds {

		playedHoles, err := ps.q.ListHolesByRound(ps.ctx, round.ID)
		if err != nil {
			fmt.Errorf("Error in ListHolesByRound(%V)", round.ID)
			return nil, err
		}

		// all shots in this round!
		playedShots, err := ps.q.ListShotsByRound(ps.ctx, round.ID)
		if err != nil {
			fmt.Errorf("Error in ListShotsByRound(%V)", round.ID)
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

		if len(playedHoles) > 0 {

			resp[i].PlayedHoles = make([]models.PlayedHole, len(playedHoles))
			for j, playedHole := range playedHoles {
				if th := teesPlayed.GetHole(playedHole.HoleNumber); th != nil {

					var shotsForThisHole []db.Shot

					for _, s := range playedShots {
						if s.HoleID == playedHole.ID {
							shotsForThisHole = append(shotsForThisHole, s)
						}
					}

					resp[i].PlayedHoles[j] = models.ConvertPlayedHole(playedHole, *th, &resp[i], shotsForThisHole)
				}
			}
		}
	}

	return resp, nil
}

func (ps RoundsService) GetActiveRoundsByPlayer(player models.Player) ([]models.Round, error) {
	savedRounds, err := ps.q.ListActiveRoundsByPlayer(ps.ctx, player.ID)
	if err != nil {
		fmt.Errorf("Error in ListActiveRoundsByPlayer(%V)", player.ID)
		return nil, err
	}
	return ps.populateRounds(player, savedRounds)
}

func (ps RoundsService) GetRoundsByPlayer(player models.Player) ([]models.Round, error) {

	savedRounds, err := ps.q.ListCompletedRoundsByPlayer(ps.ctx, player.ID)
	if err != nil {
		fmt.Errorf("Error in ListRoundsByPlayer(%V)", player.ID)
		return nil, err
	}
	return ps.populateRounds(player, savedRounds)

}

func (ps RoundsService) PersistRound(round *models.Round) error {

	// tx, err := ps.q.Begin()
	// if err != nil {
	// 	return err
	// }
	// defer tx.Rollback()
	// qtx := queries.WithTx(tx)

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

	for i := range round.PlayedHoles {
		err := ps.PersistHole(&round.PlayedHoles[i])
		if err != nil {
			return err
		}
	}

	return nil
}

func (ps RoundsService) PersistHole(ph *models.PlayedHole) error {

	if ph.ID == 0 {

		if ph.Round == nil || ph.Round.ID == 0 {
			return fmt.Errorf("PersistHole - Missing RoundId")
		}

		if ph.Hole.ID == 0 {
			return fmt.Errorf("PersistHole - Missing Hole ID")
		}

		if ph.Hole.CourseHoleID == 0 {
			return fmt.Errorf("PersistHole - Missing CourseHoleID")
		}

		log.Printf("PlayedHole - round_id [%v] hole_number [%v] course_hole_id [%v]", strconv.FormatInt(ph.Round.ID, 10), strconv.FormatInt(ph.Hole.HoleNumber, 10), strconv.FormatInt(ph.Hole.CourseHoleID, 10))

		result, err := ps.q.CreateHole(ps.ctx, db.CreateHoleParams{
			RoundID:        ph.Round.ID,
			CourseHoleID:   ph.Hole.CourseHoleID,
			HoleNumber:     ph.Hole.HoleNumber,
			DailyDistance:  ph.DistanceToPin,
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
			DailyDistance:  ph.DistanceToPin,
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
			DistanceToPin:         sh.DistanceToPin,
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
			DistanceToPin:         sh.DistanceToPin,
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
