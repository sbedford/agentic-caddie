package cli

import (
	"context"
	"fmt"
	"strconv"

	"charm.land/huh/v2"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
)

func showSetFlagPositionForm(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole) error {

	if currentHole == nil {
		return fmt.Errorf("currentHole cannot be nil for showHoleDetails")
	}

	action := ""
	pinPosition := ""

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
				Description(strconv.FormatInt(currentHole.Hole.Distance, 10)),
			huh.NewSelect[string]().
				Title("Pin Position").
				Options(
					huh.NewOption("Front Left", string(models.FlagPositionFrontLeft)),
					huh.NewOption("Front Center", string(models.FlagPositionFrontCentre)),
					huh.NewOption("Front Right", string(models.FlagPositionFrontRight)),
					huh.NewOption("Middle Left", string(models.FlagPositionMiddleLeft)),
					huh.NewOption("Middle Center", string(models.FlagPositionMiddleCentre)),
					huh.NewOption("Middle Right", string(models.FlagPositionMiddleRight)),
					huh.NewOption("Back Left", string(models.FlagPositionBackLeft)),
					huh.NewOption("Back Center", string(models.FlagPositionBackCentre)),
					huh.NewOption("Back Right", string(models.FlagPositionBackRight)),
				).
				Value(&pinPosition),
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
		currentHole.FlagPosition = models.FlagPosition(pinPosition)
		return RenderPlayHoleRoundForm(ctx, q, currentHole.Round)
	} else if action == "back" {
		return RenderPlayHoleRoundForm(ctx, q, currentHole.Round)
	}

	return nil
}

func RenderSetFlagPositionScreen(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole) error {
	return showSetFlagPositionForm(ctx, q, currentHole)
}
