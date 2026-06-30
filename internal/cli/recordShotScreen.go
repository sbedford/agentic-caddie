package cli

import (
	"context"
	"fmt"
	"strconv"

	"charm.land/huh/v2"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
	"github.com/sbedford/agentic-caddie/internal/services"
)

type shotInfo struct {
	Club                models.Club
	Distance            int64
	Type                models.ShotType
	Location            models.Location
	missDirection       models.ShotResult
	strike              models.StrikeQuality
	agentRecommendation string
}

func showSelectClubForm(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole, info *shotInfo) (string, error) {

	clubs := currentHole.Round.Golfer.Clubs

	clubOptions := make([]huh.Option[string], len(clubs))
	for i, t := range clubs {
		clubOptions[i] = huh.NewOption(t.ClubName, t.ClubName)
	}

	var selectedClub string
	clubForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your Club").
				Options(clubOptions...).
				Value(&selectedClub).
				WithHeight(10),
		),
	)
	if err := clubForm.Run(); err != nil {
		return "", err
	}

	for i, c := range clubs {
		if c.ClubName == selectedClub {
			info.Club = currentHole.Round.Golfer.Clubs[i]
		}
	}

	return selectedClub, nil

}

func showSelectShotTypeForm(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole, info *shotInfo) (string, error) {

	var selectedShotType string
	clubForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What type of shot are you playing?").
				Options(
					huh.NewOption("Tee Shot", string(models.ShotTypeTee)),
					huh.NewOption("Approach", string(models.ShotTypeApproach)),
					huh.NewOption("Layup", string(models.ShotTypeLayup)),
					huh.NewOption("Bunker", string(models.ShotTypeBunker)),
					huh.NewOption("Chip", string(models.ShotTypeChip)),
					huh.NewOption("Pitch", string(models.ShotTypePitch)),
				).
				Value(&selectedShotType).
				WithHeight(6),
		),
	)
	if err := clubForm.Run(); err != nil {
		return "", err
	}

	info.Type = models.ShotType(selectedShotType)

	return selectedShotType, nil

}

func showSetDistanceFromPinForm(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole, info *shotInfo) (int64, error) {

	var lastDistance int64 = -1
	title := "How far from the pin are you?"

	lastShot := currentHole.GetLastValidShot()
	if lastShot != nil {
		lastDistance = lastShot.DistanceToPin
		title = title + fmt.Sprintf(" (Last Shot Distance: %vm)", lastDistance)
	} else if currentHole.CurrentLocation() == models.LocationTee {
		lastDistance = currentHole.DistanceToPin
		title = title + fmt.Sprintf(" (Tee Distance: %vm)", lastDistance)
	}

	var distance string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(title).
				Value(&distance).
				WithHeight(2),
		),
	)
	if err := form.Run(); err != nil {
		return -1, err
	}

	d, err := strconv.ParseInt(distance, 10, 64)
	if err != nil {
		return -1, err
	}

	info.Distance = d

	return info.Distance, nil
}

func showSelectLocationForm(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole, info *shotInfo) (string, error) {

	var selectedLocation string
	clubForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Where did your shot finish?").
				Options(
					huh.NewOption("Fairway", string(models.LocationFairway)),
					huh.NewOption("Rough", string(models.LocationRough)),
					huh.NewOption("Green", string(models.LocationGreen)),
					huh.NewOption("Bunker", string(models.LocationBunker)),
					huh.NewOption("Hazard", string(models.LocationHazard)),
					huh.NewOption("Out of Bounds", string(models.LocationOutOfBounds)),
					huh.NewOption("Lost Ball", string(models.LocationLostBall)),
				).
				Value(&selectedLocation).
				WithHeight(7),
		),
	)
	if err := clubForm.Run(); err != nil {
		return "", err
	}

	info.Location = models.Location(selectedLocation)

	return selectedLocation, nil

}

func showMissLocationForm(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole, info *shotInfo) (string, error) {

	var selectedLocation string
	clubForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("What was the direction of your shot?").
				Options(
					huh.NewOption("Left", string(models.ShotResultLeft)),
					huh.NewOption("Right", string(models.ShotResultRight)),
					huh.NewOption("Short", string(models.ShotResultShort)),
					huh.NewOption("Long", string(models.ShotResultLong)),
				).
				Value(&selectedLocation).
				WithHeight(4),
		),
	)
	if err := clubForm.Run(); err != nil {
		return "", err
	}

	info.missDirection = models.ShotResult(selectedLocation)

	return selectedLocation, nil

}

func showStrikeQualityForm(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole, info *shotInfo) (string, error) {

	var selectedLocation string
	clubForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("How was the strike?").
				Options(
					huh.NewOption("Clean", string(models.StrikeQualityClean)),
					huh.NewOption("Fat", string(models.StrikeQualityFat)),
					huh.NewOption("Thin", string(models.StrikeQualityThin)),
					huh.NewOption("Shank", string(models.StrikeQualityShank)),
				).
				Value(&selectedLocation).
				WithHeight(4),
		),
	)
	if err := clubForm.Run(); err != nil {
		return "", err
	}

	info.strike = models.StrikeQuality(selectedLocation)

	return selectedLocation, nil

}

func showRecordShotScreen(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole, info *shotInfo) error {

	if currentHole == nil {
		return fmt.Errorf("currentHole cannot be nil for showHoleDetails")
	}

	action := ""

	holeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewNote().
				Title("Hole Number").
				Description(strconv.FormatInt(currentHole.Hole.HoleNumber, 10)).Height(1),
			huh.NewNote().
				Title("Par").
				Description(strconv.FormatInt(currentHole.Hole.Par, 10)),
			huh.NewNote().
				Title("Club").
				Description(info.Club.ClubName),
			huh.NewNote().
				Title("Location").
				Description(string(info.Location)),
			huh.NewNote().
				Title("Miss Location").
				Description(string(info.missDirection)),
			huh.NewNote().
				Title("Strike").
				Description(string(info.strike)),
			huh.NewSelect[string]().
				Title("Action").
				Options(
					huh.NewOption("Save", "save"),
					huh.NewOption("Cancel", "cancel"),
				).
				Value(&action),
		),
	)

	if err := holeForm.Run(); err != nil {
		return err
	}

	if action == "save" {
		_, err := currentHole.RecordShot(info.Distance, info.Type, info.Club, info.Location, info.missDirection, info.strike, info.agentRecommendation)

		if err != nil {
			return err
		}

		roundService := services.GetRoundsService(ctx, q)
		roundService.PersistRound(currentHole.Round)
	}

	return RenderPlayHoleRoundForm(ctx, q, currentHole.Round)
}

func RenderRecordShotScreen(ctx context.Context, q *db.Queries, currentHole *models.PlayedHole) error {

	shotInfo := shotInfo{}

	_, err := showSetDistanceFromPinForm(ctx, q, currentHole, &shotInfo)
	if err != nil {
		return err
	}

	_, err = showSelectClubForm(ctx, q, currentHole, &shotInfo)
	if err != nil {
		return err
	}

	if currentHole.CurrentLocation() != models.LocationTee {
		_, err = showSelectShotTypeForm(ctx, q, currentHole, &shotInfo)
		if err != nil {
			return err
		}
	} else {
		shotInfo.Type = models.ShotTypeTee
	}

	_, err = showSelectLocationForm(ctx, q, currentHole, &shotInfo)
	if err != nil {
		return err
	}

	if shotInfo.Location != models.LocationFairway && shotInfo.Location != models.LocationGreen {
		_, err = showMissLocationForm(ctx, q, currentHole, &shotInfo)
		if err != nil {
			return err
		}
	}

	_, err = showStrikeQualityForm(ctx, q, currentHole, &shotInfo)
	if err != nil {
		return err
	}

	return showRecordShotScreen(ctx, q, currentHole, &shotInfo)
}
