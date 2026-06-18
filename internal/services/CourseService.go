package services

import (
	"context"
	"fmt"

	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
)

type CourseService struct {
	q   *db.Queries
	ctx context.Context
}

func GetCourseService(context context.Context, db *db.Queries) CourseService {
	return CourseService{
		q:   db,
		ctx: context,
	}
}

func (ps CourseService) GetCourse(courseId int64) (*models.Course, error) {
	course, err := ps.q.GetCourseByID(ps.ctx, courseId)
	if err != nil {
		fmt.Errorf("Error in GetCourseByID(%V)", courseId)
		return nil, err
	}

	tees, err := ps.q.GetTeesByCourse(ps.ctx, courseId)
	if err != nil {
		fmt.Errorf("Error in GetTeesByCourse(%V)", courseId)
		return nil, err
	}

	holes, err := ps.q.GetHolesByCourse(ps.ctx, courseId)
	if err != nil {
		fmt.Errorf("Error in GetHolesByCourse(%V)", courseId)
		return nil, err
	}

	model := models.LoadCourse(course, tees, holes)
	if err != nil {
		fmt.Errorf("Error calling LoadCourse")
		return nil, err
	}
	return model, nil
}
