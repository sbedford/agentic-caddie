package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strconv"

	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"

	"github.com/sbedford/agentic-caddie/internal/agent"
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

func captureDistanceFromPin(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole) (int64, error) {

	if currentHole == nil {
		return -1, fmt.Errorf("currentHole cannot be nil for showHoleDetails")
	}

	distance := ""

	holeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Distance to the Pin").
				Value(&distance).
				Validate(func(str string) error {
					d, err := strconv.Atoi(str)
					if err != nil || d <= 10 {
						return errors.New("Please enter a correct distance to the pin greater than 10m")
					}
					return nil
				}),
		),
	)

	if err := holeForm.Run(); err != nil {
		return -1, err
	}

	d, _ := strconv.ParseInt(distance, 10, 64)

	return d, nil
}

func renderCaddyAdviceForm(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole) error {

	// HoleNumber            int64                // Hole 4
	// ShotNumber            int64                // Second Shot
	// CurrentLocation       models.Location      // Rough
	// LastShotMissDirection *models.ShotResult   // Right
	// Flag                  *models.FlagPosition // Front-Left
	// DistanceToThePin      int64                // 160m

	if currentHole == nil {
		return fmt.Errorf("currentHole cannot be nil for renderCaddyAdviceForm")
	}

	agentService := services.GetAgentService(ctx, q)
	var agentResponse *agent.GetHoleStrategyResponse

	// agentResponse := agent.GetHoleStrategyResponse{
	// 	Strategy:      "Hit 3-iron left of centre to avoid right OB and long bunker. Leave 100-110m approach.",
	// 	ClubSelection: "3i",
	// 	Reasoning:     "Stroke index 4 with 303 yard distance means par is the target. OB right eliminates driver/3-hybrid. 3-iron RC190 leaves ~110m in. History shows you score when hitting fairway (50% of pars vs 20% of bogeys from rough). Left miss into scrub is recoverable; right OB is catastrophic.",
	// }

	distance := currentHole.DistanceToPin

	// we need to know the distance
	if currentHole.CurrentLocation() != models.LocationTee {
		distance, _ = captureDistanceFromPin(ctx, q, currentHole)
	}

	err := spinner.New().
		Context(ctx).
		Title("Asking your caddy").
		ActionWithErr(func(ctx context.Context) error {
			var e error
			agentResponse, e = agentService.GetRecommendation(*currentHole, distance)
			return e
		}).
		Run()

	if err != nil {
		log.Fatalln(err)
	}

	holeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Hole Number").
				Description(agentResponse.Strategy).WithHeight(6),
			huh.NewNote().
				Title("Club Recommenddation").
				Description(agentResponse.ClubSelection).WithHeight(1),
		),
		huh.NewGroup(
			huh.NewNote().
				Title("Reason").
				Description(agentResponse.Reasoning).WithHeight(7),
		),
	).WithLayout(huh.LayoutGrid(2, 3))

	if err := holeForm.Run(); err != nil {
		return err
	}

	return showHoleDetails(ctx, q, currentHole)
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
							huh.NewOption("Ask the Caddy", "caddy"),
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
	} else if action == "caddy" {
		return renderCaddyAdviceForm(ctx, q, currentHole)
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
