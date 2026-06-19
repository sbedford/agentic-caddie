package cli

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"charm.land/huh/v2"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/models"
	"github.com/sbedford/agentic-caddie/internal/services"
)

func RenderLoadRoundScreen(ctx context.Context, q *db.Queries, player *models.Player) error {

	roundServices := services.GetRoundsService(ctx, q)
	rounds, err := roundServices.GetActiveRoundsByPlayer(*player)
	if err != nil {
		return err
	}

	roundOptions := make([]huh.Option[string], len(rounds))
	for i, u := range rounds {
		roundOptions[i] = huh.NewOption(fmt.Sprintf("%v round at %v", u.CompetitionType, u.Course.Name, u.RoundDate), strconv.FormatInt(int64(u.ID), 10))
	}

	var selectedRound string
	roundForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select from the folllowing active rounds").
				Options(roundOptions...).
				Value(&selectedRound).
				WithHeight(5),
		),
	)

	log.Printf("SelectedRound [%v]", selectedRound)

	if err := roundForm.Run(); err != nil {
		return err
	}

	selectedRoundID, err := strconv.ParseInt(selectedRound, 10, 64)
	if err != nil {
		return err
	}

	var round *models.Round
	for i, _ := range rounds {
		if (rounds[i].ID) == selectedRoundID {
			round = &rounds[i]
		}
	}

	return RenderPlayHoleRoundForm(ctx, q, round)
}
