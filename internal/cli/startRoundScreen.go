package cli
import (

    "fmt"
    "strconv"
    "context"
	"charm.land/huh/v2"
	"github.com/sbedford/agentic-caddie/internal/db"
)

func RenderStartRoundForm(ctx context.Context, q *db.Queries) error {
    courses, err := q.ListCourses(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch courses: %w", err)
	}

	courseOptions := make([]huh.Option[string], len(courses))
	for i, u := range courses {
		// FIX: Use strconv.FormatInt if u.ID is an integer type!
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
		return err
	}

	// --- STEP 3: Tees Selection Form ---
	// At this point, selectedCourse is guaranteed to be a populated string
	courseID, err := strconv.ParseInt(selectedCourse, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid course ID string: %w", err)
	}

	tees, err := q.GetTeesByCourse(ctx, courseID)
	if err != nil {
		return fmt.Errorf("failed to fetch tees: %w", err)
	}

	teeOptions := make([]huh.Option[string], len(tees))
	for i, t := range tees {
		teeOptions[i] = huh.NewOption(t.Name, strconv.FormatInt(int64(t.ID), 10))
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
		return err
	}

	// Success! Continue program with clean state data
	fmt.Printf("Setup Complete! Course ID: %d, Tee ID: %s\n", courseID, selectedTee)
	return nil
}
