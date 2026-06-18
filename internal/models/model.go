package models

import (
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

	RoundType       string
	CompetitionType string

	TotalScore  int64
	TotalPoints int64
	TotalPutts  int64

	PlayedHoles []PlayedHole
}

type PlayedHole struct {
	Hole Hole

	FlagPosition string

	Score  int64
	Points int64

	NumberOfPutts     int64
	FairwayHit        bool
	GreenInRegulation bool
	ScrambleSave      bool
	Penalty           bool

	PreShotRecommendation string
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
		RoundType:       in.RoundType,
		CompetitionType: helpers.String(in.CompetitionType),

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
		Hole:                  h,
		Score:                 helpers.Int64(in.Score),
		Points:                helpers.Int64(in.Points),
		NumberOfPutts:         helpers.Int64(in.Putts),
		FairwayHit:            helpers.Bool(in.FairwayHit),
		GreenInRegulation:     helpers.Bool(in.Gir),
		ScrambleSave:          helpers.Bool(in.ScrambleSave),
		Penalty:               helpers.Bool(in.Penalty),
		PreShotRecommendation: "",
	}

}
