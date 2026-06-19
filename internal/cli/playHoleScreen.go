package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"charm.land/huh/v2"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
	"github.com/sbedford/agentic-caddie/internal/services"
)

func renderPuttOutForm(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole) error {
	if currentHole == nil {
		return fmt.Errorf("currentHole cannot be nil for showPuttOutForm")
	}

	action := ""
	numberOfPutts := ""

	holeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Hole Number").
				Description(strconv.FormatInt(currentHole.Hole.HoleNumber, 10)).Height(1),
			huh.NewNote().
				Title("Par").
				Description(strconv.FormatInt(currentHole.Hole.Par, 10)),
			huh.NewNote().
				Title("Shots Given").
				Description(strconv.FormatInt(currentHole.ShotsGiven(), 10)),
			huh.NewNote().
				Title("Current Location").
				Description(string(currentHole.CurrentLocation())),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("Distance (m)").
				Description(strconv.FormatInt(currentHole.DistanceToPin, 10)),
			huh.NewInput().
				Title("Number of Putts").
				Value(&numberOfPutts).
				// Validating fields is easy. The form will mark erroneous fields
				// and display error messages accordingly.
				Validate(func(str string) error {
					d, err := strconv.Atoi(numberOfPutts)
					if err != nil || d < 0 {
						return errors.New("Please enter a number >= 0")
					}
					return nil
				}),
			huh.NewSelect[string]().
				Title("Action").
				Options(
					huh.NewOption("Save", "save"),
					huh.NewOption("Back", "back"),
				).
				Value(&action),
		),
	).WithLayout(huh.LayoutGrid(2, 3))

	if err := holeForm.Run(); err != nil {
		return err
	}

	if action == "save" {
		putts, _ := strconv.ParseInt(numberOfPutts, 10, 64)
		currentHole.CompleteHole(putts)

		roundService := services.GetRoundsService(ctx, q)
		roundService.PersistRound(currentHole.Round)

		return RenderPlayHoleRoundForm(ctx, q, currentHole.Round)
	} else if action == "back" {
		return RenderPlayHoleRoundForm(ctx, q, currentHole.Round)
	}

	return nil
}

func showHoleDetails(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole) error {

	if currentHole == nil {
		return fmt.Errorf("currentHole cannot be nil for showHoleDetails")
	}

	action := ""

	holeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Hole Number").
				Description(strconv.FormatInt(currentHole.Hole.HoleNumber, 10)),
			huh.NewNote().
				Title("Par").
				Description(strconv.FormatInt(currentHole.Hole.Par, 10)),
			huh.NewNote().
				Title("Shots Given").
				Description(strconv.FormatInt(currentHole.ShotsGiven(), 10)),
			huh.NewNote().
				Title("Current Location").
				Description(string(currentHole.CurrentLocation())),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("Distance (m)").
				Description(strconv.FormatInt(currentHole.DistanceToPin, 10)),
			huh.NewNote().
				Title("Flag Position").
				Description(string(currentHole.FlagPosition)),
			huh.NewNote().
				Title("Shots Taken").
				Description(strconv.Itoa(len(currentHole.ShotsTaken))),
			huh.NewSelect[string]().
				Title("Action").
				OptionsFunc(func() []huh.Option[string] {

					if currentHole.CurrentLocation() == models.LocationGreen {
						return []huh.Option[string]{
							huh.NewOption("Set Flag Position", "flag"),
							huh.NewOption("Putt Out", "putt-out"),
							huh.NewOption("Wipe Hole", "wipe"),
						}
					} else {
						return []huh.Option[string]{
							huh.NewOption("Set Flag Position", "flag"),
							huh.NewOption("Set Distance from Pin", "distance"),
							huh.NewOption("Record Shot", "record"),
							huh.NewOption("Wipe Hole", "wipe"),
						}
					}

				}, &action).
				Value(&action),
		),
	).WithLayout(huh.LayoutGrid(2, 3))

	if err := holeForm.Run(); err != nil {
		return err
	}

	if action == "record" {
		return RenderRecordShotScreen(ctx, q, currentHole)
	} else if action == "flag" {
		return RenderSetFlagPositionScreen(ctx, q, currentHole)
	} else if action == "distance" {
		return RenderSetDistanceFromHoleScreen(ctx, q, currentHole)
	} else if action == "putt-out" {
		return renderPuttOutForm(ctx, q, currentHole)
	} else if action == "wipe" {
		currentHole.RecordWipe()

		roundService := services.GetRoundsService(ctx, q)
		roundService.PersistRound(currentHole.Round)

		return RenderPlayHoleRoundForm(ctx, q, currentHole.Round)
	}

	return nil
}

func RenderPlayHoleRoundForm(ctx context.Context, q *db.Queries, round *models.Round) error {

	if round == nil {
		return fmt.Errorf("Round is nil")
	}

	currentHole, err := round.CurrentHole()
	if err != nil {
		return err
	}

	return showHoleDetails(ctx, q, currentHole)

}
