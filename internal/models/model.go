package models

import (
	"cmp"
	"fmt"
	"slices"
	"time"

	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/helpers"
)

type Player struct {
	ID       int64
	Name     string
	Handicap float64
	Clubs    []Club
	Rounds   []Round
}

type Club struct {
	Player         Player
	ClubName       string
	AddedDate      time.Time
	RemovedDate    time.Time
	CarryAvg       float64
	CarryReliable  float64
	CarryMax       float64
	DispersionAvgM float64
	DispersionBias string
	SampleSize     int64
	CalculatedAt   time.Time
}

func (this *Club) Load(c db.PlayerClub, p Player) {
	this.Player = p
	this.ClubName = c.ClubName
	this.CarryAvg = helpers.Float64(c.CarryAvg)
	this.CarryReliable = helpers.Float64(c.CarryReliable)
	this.CarryMax = helpers.Float64(c.CarryMax)
	this.DispersionAvgM = helpers.Float64(c.DispersionAvgM)
	this.DispersionBias = helpers.String(c.DispersionBias)
}

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

// Finds the last hole
func (this *Round) CurrentHole() (*PlayedHole, error) {

	if !this.RoundCompleted {

		incompleteHoles := helpers.Filter(this.PlayedHoles, func(ph PlayedHole) bool {
			return ph.Completed == false
		})

		if len(incompleteHoles) == 0 {
			return nil, fmt.Errorf("Round %v is not marked as completed but has no incomplete holes", this.ID)
		} else if len(incompleteHoles) > 1 {
			return nil, fmt.Errorf("Round %v has more than 1 incomplete hole", this.ID)
		}

		return &incompleteHoles[0], nil
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

type PlayedHole struct {
	Hole Hole

	FlagPosition FlagPosition

	Score  int64
	Points int64

	NumberOfPutts     int64
	FairwayHit        bool
	GreenInRegulation bool
	ScrambleSave      bool
	Penalty           bool

	PreShotRecommendation string
	Completed             bool

	ShotsTaken []Shot
}

func (hole *PlayedHole) LastShot() *Shot {

	if hole.ShotsTaken == nil || len(hole.ShotsTaken) == 0 {
		return nil
	}
	lastShot := slices.MaxFunc(hole.ShotsTaken, func(a, b Shot) int {
		return cmp.Compare(a.ShotNumber, b.ShotNumber)
	})
	return &lastShot
}

type Location string
type ShotType string
type ShotResult string
type StrikeQuality string
type CompetitionType string
type RoundType string
type FlagPosition string

const (
	CompetitionTypeStableford CompetitionType = "stableford"
	CompetitionTypeStroke     CompetitionType = "stroke"
	CompetitionTypeOther      CompetitionType = "other"

	RoundTypeSocial      RoundType = "social"
	RoundTypeCompetition RoundType = "competition"
	RoundTypePractice    RoundType = "practice"

	LocationTee           Location = "tee"
	LocationFairway       Location = "fairway"
	LocationRough         Location = "rough"
	LocationBunker        Location = "bunker"
	LocationGreen         Location = "green"
	LocationHazard        Location = "hazard"
	LocationHoleCompleted Location = "hole completed"
	LocationOutOfBounds   Location = "ob"
	LocationLostBall      Location = "lost"

	ShotTypeTee      ShotType = "tee"
	ShotTypeApproach ShotType = "approach"
	ShotTypeLayup    ShotType = "layup"
	ShotTypeChip     ShotType = "chip"
	ShotTypePitch    ShotType = "pitch"
	ShotTypeBunker   ShotType = "bunker"
	ShotTypeRecord   ShotType = "recovery"

	ShotResultLeft  ShotResult = "left"
	ShotResultRight ShotResult = "righ"
	ShotResultShort ShotResult = "short"
	ShotResultLong  ShotResult = "long"

	StrikeQualityClean StrikeQuality = "clean"
	StrikeQualityFat   StrikeQuality = "fat"
	StrikeQualityThin  StrikeQuality = "thin"
	StrikeQualityShank StrikeQuality = "shank"

	FlagPositionFrontLeft    FlagPosition = "front_left"
	FlagPositionFrontCentre  FlagPosition = "front_centre"
	FlagPositionFrontRight   FlagPosition = "front_right"
	FlagPositionMiddleLeft   FlagPosition = "front_left"
	FlagPositionMiddleCentre FlagPosition = "front_centre"
	FlagPositionMiddleRight  FlagPosition = "front_right"
	FlagPositionBackLeft     FlagPosition = "front_left"
	FlagPositionBackCentre   FlagPosition = "front_centre"
	FlagPositionBackRight    FlagPosition = "front_right"
)

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

	var lastValidShot *Shot = nil
	for _, shot := range hole.ShotsTaken {
		if shot.ValidShot() {
			lastValidShot = &shot
		}
	}

	return lastValidShot
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

func (hole *PlayedHole) RecordShot(shotType ShotType, clubUsed Club, result Location, missDirection ShotResult, strike StrikeQuality, agentRecommendation string) (*Shot, error) {

	if hole.Completed {
		return nil, fmt.Errorf("Hole %v already completed", hole.Hole.HoleNumber)
	}

	shotNumber := 1
	lastShot := hole.LastShot()
	if lastShot != nil {
		shotNumber = int(lastShot.ShotNumber) + 1
	}

	shot := Shot{
		Hole:                  *hole,
		ShotNumber:            int64(shotNumber),
		ShotType:              shotType,
		Club:                  clubUsed.ClubName,
		Result:                result,
		Miss:                  missDirection,
		StrikeQuality:         strike,
		PreShotRecommendation: agentRecommendation,
		Completed:             true,
	}

	hole.ShotsTaken = append(hole.ShotsTaken, shot)

	return &shot, nil
}

type Shot struct {
	Hole                  PlayedHole
	ShotNumber            int64 // 1, 2,3,4
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

type Course struct {
	Id   int64
	Name string
	Tees []Tee
}

type Position struct {
	latitude  float64
	longitude float64
}

type Tee struct {
	ID           int64
	Name         string
	SlopeIndex   int64
	CourseRating float64
	TeeCentre    Position
	Holes        []Hole
}

func (this *Course) FindTee(teeName string) *Tee {
	idx := slices.IndexFunc(this.Tees, func(u Tee) bool {
		return u.Name == teeName
	})
	if idx != -1 {
		return &this.Tees[idx]
	}
	return nil
}

func (this *Tee) GetHole(holeNumber int64) *Hole {
	idx := slices.IndexFunc(this.Holes, func(u Hole) bool {
		return u.HoleNumber == holeNumber
	})
	if idx != -1 {
		return &this.Holes[idx]
	}
	return nil
}

type Hole struct {
	Tee         Tee
	HoleNumber  int64
	Distance    int64
	Par         int64
	StrokeIndex int64
	GreenCentre Position
}

func ConvertCourse(in db.Course) Course {
	return Course{
		Id:   in.ID,
		Name: in.Name,
	}
}

func ConvertTee(in db.Tee) Tee {
	return Tee{
		ID:           in.ID,
		Name:         in.Name,
		SlopeIndex:   helpers.Int64(in.SlopeRating),
		CourseRating: helpers.Float64(in.CourseRating),
	}
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

func ConvertHole(in db.GetHolesByCourseRow, t Tee) Hole {
	return Hole{
		Tee:         t,
		HoleNumber:  in.HoleNumber,
		Distance:    in.Distance,
		Par:         in.Par,
		StrokeIndex: helpers.Int64(in.StrokeIndex),
	}
}

func LoadCourse(c db.Course, t []db.Tee, h []db.GetHolesByCourseRow) *Course {
	course := ConvertCourse(c)

	course.Tees = make([]Tee, len(t))
	for i, tee := range t {
		teeModel := ConvertTee(tee)

		teeHoles := helpers.Filter(h, func(hh db.GetHolesByCourseRow) bool {
			return hh.Teename == tee.Name
		})
		teeModel.Holes = make([]Hole, len(teeHoles))

		for j, teeHole := range teeHoles {
			teeModel.Holes[j] = ConvertHole(teeHole, teeModel)
		}
		course.Tees[i] = teeModel
	}

	return &course
}

func ConvertPlayedHole(in db.Hole, h Hole) PlayedHole {
	return PlayedHole{
		Hole:              h,
		Score:             helpers.Int64(in.Score),
		Points:            helpers.Int64(in.Points),
		NumberOfPutts:     helpers.Int64(in.Putts),
		FairwayHit:        helpers.Bool(in.FairwayHit),
		GreenInRegulation: helpers.Bool(in.Gir),
		ScrambleSave:      helpers.Bool(in.ScrambleSave),
		Penalty:           helpers.Bool(in.Penalty),
		Completed:         helpers.Bool(in.Completed),
	}
}
