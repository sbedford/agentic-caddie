package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"charm.land/huh/v2"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
	"github.com/sbedford/agentic-caddie/internal/services"
)

func renderCourseForm(ctx context.Context, q *db.Queries) (int64, error) {
	courses, err := q.ListCourses(ctx)
	if err != nil {
		return -1, fmt.Errorf("failed to fetch courses: %w", err)
	}

	courseOptions := make([]huh.Option[string], len(courses))
	for i, u := range courses {
		courseOptions[i] = huh.NewOption(u.Name, strconv.FormatInt(int64(u.ID), 10))
	}

	var selectedCourse string
	courseForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your course").
				Options(courseOptions...).
				Value(&selectedCourse).
				WithHeight(5),
		),
	)

	if err := courseForm.Run(); err != nil {
		return -1, err
	}

	courseID, err := strconv.ParseInt(selectedCourse, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("invalid course ID string: %w", err)
	}
	return courseID, nil
}

func renderTeeForm(ctx context.Context, q *db.Queries, courseID int64) (string, error) {
	tees, err := q.GetTeesByCourse(ctx, courseID)
	if err != nil {
		return "", fmt.Errorf("failed to fetch tees: %w", err)
	}

	teeOptions := make([]huh.Option[string], len(tees))
	for i, t := range tees {
		teeOptions[i] = huh.NewOption(t.Name, t.Name)
	}

	var selectedTee string
	teeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your Tees").
				Options(teeOptions...).
				Value(&selectedTee).
				WithHeight(3),
		),
	)
	if err := teeForm.Run(); err != nil {
		return "", err
	}
	return selectedTee, nil
}

func renderRoundTypeForm(ctx context.Context, q *db.Queries) (string, error) {

	var roundType string
	teeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Round Type").
				Options(
					huh.NewOption("Practice", string(models.RoundTypePractice)),
					huh.NewOption("Competition", string(models.RoundTypeCompetition)),
					huh.NewOption("Social", string(models.RoundTypeSocial)),
				).
				Value(&roundType).
				WithHeight(3),
		),
	)
	if err := teeForm.Run(); err != nil {
		return "", err
	}
	return roundType, nil
}

func renderCompetitionTypeForm(ctx context.Context, q *db.Queries) (string, error) {

	var compType string
	teeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Competition Type").
				Options(
					huh.NewOption("Stableford", string(models.CompetitionTypeStableford)),
					huh.NewOption("Stroke", string(models.CompetitionTypeStroke)),
				).
				Value(&compType).
				WithHeight(3),
		),
	)
	if err := teeForm.Run(); err != nil {
		return "", err
	}
	return compType, nil
}

func RenderStartRoundForm(ctx context.Context, q *db.Queries, player *models.Player) error {

	courseID, err := renderCourseForm(ctx, q)

	if err != nil {
		return err
	}

	teeName, err := renderTeeForm(ctx, q, courseID)
	if err != nil {
		return err
	}

	courseService := services.GetCourseService(ctx, q)
	course, err := courseService.GetCourse(courseID)
	if err != nil {
		return err
	}

	tee := course.FindTee(teeName)
	if tee == nil {
		return fmt.Errorf("Error retrieving Tee [%v] for Course [%v]", teeName, course.ID)
	}

	roundType, err := renderRoundTypeForm(ctx, q)
	if err != nil {
		return err
	}

	competitionType := ""
	if roundType == string(models.RoundTypeCompetition) {
		competitionType, err = renderCompetitionTypeForm(ctx, q)
		if err != nil {
			return err
		}
	}

	round, err := models.InitialiseRound(*course, *player, *tee, time.Now(), int64(player.Handicap), models.RoundType(roundType), models.CompetitionType(competitionType))
	if err != nil {
		return err
	}

	roundService := services.GetRoundsService(ctx, q)

	err = roundService.PersistRound(round)
	if err != nil {
		return err
	}

	return RenderPlayHoleRoundForm(ctx, q, round)
}
