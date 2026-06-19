package models

import (
	"fmt"
	"slices"
	"time"

	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/helpers"
)

/// --------------------------------------
/// --- Round Entities and Extensions ----
/// --------------------------------------

type Round struct {
	ID int64

	Course Course
	Golfer Player
	Tee    Tee

	RoundDate time.Time

	DailyHandicap int64

	RoundType       RoundType
	CompetitionType CompetitionType

	TotalScore  int64
	TotalPoints int64
	TotalPutts  int64

	RoundCompleted bool

	PlayedHoles []PlayedHole
}

func (this *Round) IsStableford() bool {
	return this.CompetitionType == CompetitionTypeStableford
}

func (this *Round) IsStroke() bool {
	return this.CompetitionType == CompetitionTypeStroke
}

func (this *Round) PointsBehind() int {
	if this.IsStableford() {

		completedHoles := helpers.Filter(this.PlayedHoles, func(ph PlayedHole) bool {
			return ph.Completed
		})

		return (2 * len(completedHoles)) - int(this.TotalPoints)
	}
	return 0
}

// Finds the last completed hole
func (this *Round) LastCompletedHole() (*PlayedHole, error) {

	if !this.RoundCompleted {

		completedHoles := helpers.Filter(this.PlayedHoles, func(ph PlayedHole) bool {
			return ph.Completed == true
		})

		if len(completedHoles) == 0 {
			return nil, fmt.Errorf("Round %v has no completed holes", this.ID)
		}

		return (completedHoles[len(completedHoles)-1]), nil
	}
	return nil, nil
}

// Finds the current active hole
func (this *Round) CurrentHole() (*PlayedHole, error) {

	if !this.RoundCompleted {

		var incompleteHole *PlayedHole = nil
		for i, hole := range this.PlayedHoles {
			if !hole.Completed {
				if incompleteHole != nil {
					return nil, fmt.Errorf("Round %v has more than 1 incomplete hole", this.ID)
				}
				incompleteHole = &this.PlayedHoles[i]
			}
		}

		if incompleteHole == nil {
			var err error
			incompleteHole, err = this.ProgressHole()
			if err != nil {
				return nil, err
			}

			if incompleteHole == nil {
				return nil, fmt.Errorf("Round %v is not marked as complete but has no incomplete holes", this.ID)
			}
		}

		return incompleteHole, nil
	}
	return nil, nil
}

func (this *Round) ProgressHole() (*PlayedHole, error) {
	if !this.RoundCompleted {

		// Fresh round, build a new record for Hole 1
		if this.PlayedHoles == nil || len(this.PlayedHoles) == 0 {

			th := this.Tee.GetHole(1)
			if th != nil {

				this.PlayedHoles = append(this.PlayedHoles, PlayedHole{
					Round:         this,
					Hole:          *th,
					DistanceToPin: *&th.Distance,
				})
				return &this.PlayedHoles[len(this.PlayedHoles)-1], nil

			}
			return nil, fmt.Errorf("Unable to retrieve Hole 1 - Course/Tee not initialised?")
		} else {
			lastHole, err := this.LastCompletedHole()
			if err != nil {
				return nil, err
			}

			if lastHole.Hole.HoleNumber < 18 {
				th := this.Tee.GetHole(lastHole.Hole.HoleNumber + 1)
				if th != nil {

					this.PlayedHoles = append(this.PlayedHoles, PlayedHole{
						Round:         this,
						Hole:          *th,
						DistanceToPin: *&th.Distance,
					})
					return &this.PlayedHoles[len(this.PlayedHoles)-1], nil
				}
			}
			return nil, fmt.Errorf("18 holes Completed")
		}
	}
	return nil, nil
}

func (this *Round) StrokesAbovePar() int {
	if this.IsStableford() {

		completedHoles := helpers.Filter(this.PlayedHoles, func(ph PlayedHole) bool {
			return ph.Completed
		})

		par := 0
		for _, ch := range completedHoles {

			// needs to factor in shots given on a hole, but data model doesnt support yet.

			par += int(ch.Hole.Par)
		}

		return (-1 * (int(this.TotalScore) - par))
	}
	return 0
}

func ConvertRound(in db.Round, c Course, p Player, t Tee) Round {
	return Round{
		ID:              in.ID,
		Course:          c,
		Tee:             t,
		Golfer:          p,
		RoundDate:       in.PlayedAt,
		DailyHandicap:   in.DailyHandicap,
		RoundType:       RoundType(in.RoundType),
		CompetitionType: CompetitionType(helpers.String(in.CompetitionType)),

		RoundCompleted: in.Completed,

		TotalScore:  helpers.Int64(in.TotalScore),
		TotalPoints: helpers.Int64(in.TotalPoints),
		TotalPutts:  helpers.Int64(in.TotalPutts),
	}
}

func (this *Round) ensurePointsConsistency() {

	var strokes int64 = 0
	var points int64 = 0
	var putts int64 = 0

	for _, h := range this.PlayedHoles {
		if h.Completed {
			strokes += h.Score
			points += h.Points
			putts += h.NumberOfPutts
		}
	}

	this.TotalScore = strokes
	this.TotalPoints = points
	this.TotalPutts = putts

}

func InitialiseRound(c Course, p Player, t Tee, roundDate time.Time, dailyHandicap int64, roundType RoundType, competitionnType CompetitionType) (*Round, error) {
	round := Round{
		Course:          c,
		Tee:             t,
		Golfer:          p,
		RoundDate:       roundDate,
		DailyHandicap:   dailyHandicap,
		RoundType:       roundType,
		CompetitionType: competitionnType,

		RoundCompleted: false,

		TotalScore:  0,
		TotalPoints: 0,
		TotalPutts:  0,
	}

	// setup the first hole
	_, err := round.ProgressHole()
	if err != nil {
		return nil, err
	}

	return &round, nil
}

/// --------------------------------------
/// --- Played Hole Entities and Extensions ----
/// --------------------------------------

type PlayedHole struct {
	Round *Round
	ID    int64

	Hole Hole

	DistanceToPin int64
	FlagPosition  FlagPosition

	Score  int64
	Points int64

	NumberOfPutts     int64
	FairwayHit        bool
	GreenInRegulation bool
	ScrambleSave      bool
	Penalty           bool
	PenaltyStrokes    int64

	Completed bool
	Wiped     bool

	ShotsTaken []Shot
}

func (hole *PlayedHole) LastShot() *Shot {

	if hole.ShotsTaken == nil || len(hole.ShotsTaken) == 0 {
		return nil
	}

	maxIdx := 0
	for i := range hole.ShotsTaken {
		if hole.ShotsTaken[i].ShotNumber > hole.ShotsTaken[maxIdx].ShotNumber {
			maxIdx = i
		}
	}
	return &hole.ShotsTaken[maxIdx]
}

func (hole *PlayedHole) ShotsGiven() int64 {
	if hole.Round != nil && hole.Hole.StrokeIndex <= hole.Round.DailyHandicap {
		return 1
	}
	return 0
}

func (hole *PlayedHole) CurrentLocation() Location {

	if hole.Completed {
		return LocationHoleCompleted
	}

	if hole.ShotsTaken == nil || len(hole.ShotsTaken) == 0 {
		return LocationTee
	}

	lastShot := hole.GetLastValidShot()
	if lastShot == nil {
		return LocationTee
	}

	switch lastShot.Result {
	case "fairway":
		return LocationFairway
	case "green":
		return LocationGreen
	case "rough":
		return LocationRough
	case "bunker":
		return LocationBunker
	case "hazard":
		return LocationHazard
	}
	return LocationTee
}

func (hole *PlayedHole) GetLastValidShot() *Shot {

	if hole.ShotsTaken == nil || len(hole.ShotsTaken) == 0 {
		return nil
	}

	lastIdx := -1
	for i := range hole.ShotsTaken {
		if hole.ShotsTaken[i].ValidShot() {
			lastIdx = i
		}
	}

	if lastIdx == -1 {
		return nil
	}
	return &hole.ShotsTaken[lastIdx]
}

func (hole *PlayedHole) LookupShot(shotNumber int) (*Shot, error) {
	idx := slices.IndexFunc(hole.ShotsTaken, func(s Shot) bool {
		return shotNumber == int(s.ShotNumber)
	})

	if idx < 0 {
		return nil, fmt.Errorf("Shot with number %v doesnt exist", shotNumber)
	}

	return &hole.ShotsTaken[idx], nil
}

func (hole *PlayedHole) RecordShot(DistanceToPin int64, shotType ShotType, clubUsed Club, result Location, missDirection ShotResult, strike StrikeQuality, agentRecommendation string) (*Shot, error) {

	if hole.Completed {
		return nil, fmt.Errorf("Hole %v already completed", hole.Hole.HoleNumber)
	}

	shotNumber := 1
	lastShot := hole.LastShot()
	if lastShot != nil {
		shotNumber = int(lastShot.ShotNumber) + 1
	}

	hole.FairwayHit = (shotNumber == 1 && result == LocationFairway)

	hole.GreenInRegulation = (result == LocationGreen && (shotNumber <= int(hole.Hole.Par)-2))

	hole.ShotsTaken = append(hole.ShotsTaken, Shot{
		Hole:                  hole,
		DistanceToPin:         DistanceToPin,
		ShotNumber:            int64(shotNumber),
		ShotType:              shotType,
		Club:                  clubUsed.ClubName,
		Result:                result,
		Miss:                  missDirection,
		StrikeQuality:         strike,
		PreShotRecommendation: agentRecommendation,
		Completed:             true,
	})

	return &hole.ShotsTaken[len(hole.ShotsTaken)-1], nil
}

func (hole *PlayedHole) CompleteHole(numberOfPutts int64) {

	hole.NumberOfPutts = numberOfPutts
	hole.Score = int64(len(hole.ShotsTaken)) + numberOfPutts + int64(hole.PenaltyStrokes)
	hole.Completed = true

	hole.CalculatePoints()
}

func (hole *PlayedHole) CalculatePoints() {
	hole.Points = 0
	if hole.Round != nil && hole.Round.IsStableford() {
		revisedPar := hole.Hole.Par

		if hole.Hole.StrokeIndex <= hole.Round.DailyHandicap {
			revisedPar += 1
		}

		hole.Points = (revisedPar - hole.Score) + 2
		if hole.Points < 0 {
			hole.Points = 0
		}
	}

	hole.Round.ensurePointsConsistency()
}

func (hole *PlayedHole) RecordPenalty(penaltyStrokes int64) {
	hole.Penalty = true
	hole.PenaltyStrokes += penaltyStrokes
}

func (hole *PlayedHole) RecordWipe() {
	hole.Wiped = true
	hole.Completed = true
}

func ConvertPlayedHole(in db.Hole, h Hole, r *Round, existingShots []db.Shot) PlayedHole {

	dd := in.DailyDistance
	if dd == 0 {
		dd = h.Distance
	}
	hole := PlayedHole{
		ID:                in.ID,
		Round:             r,
		Hole:              h,
		FlagPosition:      FlagPosition(helpers.String(in.FlagPosition)),
		DistanceToPin:     dd,
		Score:             helpers.Int64(in.Score),
		Points:            helpers.Int64(in.Points),
		NumberOfPutts:     helpers.Int64(in.Putts),
		FairwayHit:        helpers.Bool(in.FairwayHit),
		GreenInRegulation: helpers.Bool(in.Gir),
		ScrambleSave:      helpers.Bool(in.ScrambleSave),
		Penalty:           helpers.Bool(in.Penalty),
		Completed:         helpers.Bool(in.Completed),
	}

	hole.ShotsTaken = make([]Shot, len(existingShots))
	for i, s := range existingShots {

		// club, err := r.Golfer.GetClub(helpers.String(s.Club))
		// if err != nil {
		// 	return nil, err
		// }

		shot := Shot{
			ID:                    s.ID,
			Hole:                  &hole,
			ShotNumber:            s.ShotNumber,
			DistanceToPin:         s.DistanceToPin,
			ShotType:              ShotType(s.ShotType),
			Club:                  helpers.String(s.Club),
			Result:                Location(helpers.String(s.Result)),
			Miss:                  ShotResult(helpers.String(s.Miss)),
			StrikeQuality:         StrikeQuality(helpers.String(s.StrikeQuality)),
			PreShotRecommendation: helpers.String(s.PreShotRecommendation),
			Completed:             helpers.Bool(s.Completed),
			Source:                s.Source,
		}
		hole.ShotsTaken[i] = shot
	}

	return hole
}

/// --------------------------------------
/// --- Shot Entities and Extensions ----
/// --------------------------------------

type Shot struct {
	ID                    int64
	Hole                  *PlayedHole
	ShotNumber            int64 // 1, 2,3,4
	DistanceToPin         int64
	ShotType              ShotType
	Club                  string
	Result                Location
	Miss                  ShotResult
	StrikeQuality         StrikeQuality
	PreShotRecommendation string
	Completed             bool
	Source                string
}

func (shot *Shot) ValidShot() bool {
	return (shot.Result == LocationFairway ||
		shot.Result == LocationRough ||
		shot.Result == LocationBunker ||
		shot.Result == LocationHazard ||
		shot.Result == LocationGreen)
}
