package cli

import (
    "context"
	"charm.land/huh/v2"
	"github.com/sbedford/agentic-caddie/internal/db"
)


type FormState struct {
    Option string
    Course string
    Tee    string
}

func RenderForm(ctx context.Context, q *db.Queries) error {
	// --- STEP 1: Welcome Form ---
	var selectedOption string = ""
	welcomeForm := huh.NewForm(
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

	// Exit early or branch logic if they choose something else
	if selectedOption == "load" {
		// Handle load round sequence here...
		return nil
	} 
    
    return RenderStartRoundForm(ctx,q)
}