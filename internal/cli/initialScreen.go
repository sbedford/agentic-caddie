package cli

import (
	"context"
	"fmt"
	"strconv"

	"charm.land/huh/v2"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/services"
)

type FormState struct {
	Option string
	Course string
	Tee    string
}

func RenderForm(ctx context.Context, q *db.Queries) error {
	var selectedOption string = ""
	var selectedPlayer string = ""

	players, err := q.ListPlayers(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch players: %w", err)
	}

	playerOptions := make([]huh.Option[string], len(players))
	for i, u := range players {
		playerOptions[i] = huh.NewOption(u.Name, strconv.FormatInt(int64(u.ID), 10))
	}

	welcomeForm := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select your player").
				Options(playerOptions...).
				Value(&selectedPlayer).
				WithHeight(5),
		),
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Welcome").
				Options(
					huh.NewOption("Start a new round", "start"),
					huh.NewOption("Load an existing round", "load"),
				).
				Value(&selectedOption),
		),
	)
	if err := welcomeForm.Run(); err != nil {
		return err
	}

	playerService := services.GetPlayerService(ctx, q)
	playerId, err := strconv.ParseInt(selectedPlayer, 10, 64)
	if err != nil {
		return err
	}

	player, err := playerService.GetPlayer(playerId)
	if err != nil {
		return err
	}

	if selectedOption == "load" {
		return RenderLoadRoundScreen(ctx, q, player)
	}

	return RenderStartRoundForm(ctx, q, player)
}
