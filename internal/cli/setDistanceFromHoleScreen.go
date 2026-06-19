package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"charm.land/huh/v2"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
)

func showRecordDistanceFromHoleForm(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole) error {

	if currentHole == nil {
		return fmt.Errorf("currentHole cannot be nil for showHoleDetails")
	}

	action := ""
	distance := ""

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
				Title("Distance from the Pin").
				Value(&distance).
				// Validating fields is easy. The form will mark erroneous fields
				// and display error messages accordingly.
				Validate(func(str string) error {
					d, err := strconv.Atoi(str)
					if err != nil || d <= 10 {
						return errors.New("Please enter a correct distance from the pin greater than 10m")
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
		dd, _ := strconv.ParseInt(distance, 10, 64)
		currentHole.DistanceToPin = dd
		return RenderPlayHoleRoundForm(ctx, q, currentHole.Round)
	} else if action == "back" {
		return RenderPlayHoleRoundForm(ctx, q, currentHole.Round)
	}

	return nil
}

func RenderSetDistanceFromHoleScreen(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole) error {
	return showRecordDistanceFromHoleForm(ctx, q, currentHole)
}
