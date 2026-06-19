package models

import (
	"slices"

	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/helpers"
)

type Course struct {
	ID   int64
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

// Maps to TeeHole
type Hole struct {
	ID           int64 // TeeHoleID
	CourseHoleID int64
	Tee          Tee
	HoleNumber   int64
	Distance     int64
	Par          int64
	StrokeIndex  int64
	GreenCentre  Position
}

func ConvertCourse(in db.Course) Course {
	return Course{
		ID:   in.ID,
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
			teeModel.Holes[j] = ConvertHole(*teeHole, teeModel)
		}
		course.Tees[i] = teeModel
	}

	return &course
}

func ConvertHole(in db.GetHolesByCourseRow, t Tee) Hole {
	return Hole{
		ID:           in.Teeholeid,
		CourseHoleID: in.Courseholeid,
		Tee:          t,
		HoleNumber:   in.HoleNumber,
		Distance:     in.Distance,
		Par:          in.Par,
		StrokeIndex:  helpers.Int64(in.StrokeIndex),
	}
}
