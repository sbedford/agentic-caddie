package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/sbedford/agentic-caddie/internal/db"
)

var GetHoleStatsToolDef = anthropic.ToolUnionParam{
	OfTool: &anthropic.ToolParam{
		Name:        "get_hole_stats",
		Description: anthropic.String("Returns historical per-round results for a specific player on a specific hole at a specific course: date, score, points, putts, fairway hit, and GIR for every recorded round on that hole. Use this to look for patterns on individual holes rather than whole rounds."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"player_id": map[string]any{
					"type":        "integer",
					"description": "The player identifier",
				},
				"course_id": map[string]any{
					"type":        "integer",
					"description": "The course identifier, as returned by get_round_history or get_course_info.",
				},
				"tee_name": map[string]any{
					"type":        "string",
					"description": "The name of the tee's being played",
				},
				"hole_num": map[string]any{
					"type":        "integer",
					"description": "Hole number, 1-18.",
				},
			},
			Required: []string{"course_id", "tee", "hole_num"},
		},
	},
}

// --- Handler ---

type getHoleStatsInput struct {
	CourseID int64  `json:"course_id"`
	TeeName  string `json:"tee_name"`
	HoleNum  int64  `json:"hole_num"`
	PlayerID int64  `json:"player_id"`
}

type holeStatsHandler struct {
	queries *db.Queries
}

// type ToolHandler func(ctx context.Context, input json.RawMessage) (string, error)
func (h *holeStatsHandler) handle(ctx context.Context, raw json.RawMessage) (string, error) {
	var in getHoleStatsInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if in.HoleNum < 1 || in.HoleNum > 18 {
		return "", fmt.Errorf("hole_num must be between 1 and 18, got %d", in.HoleNum)
	}

	rows, err := h.queries.GetHoleStats(ctx, db.GetHoleStatsParams{
		Courseid:   in.CourseID,
		Holenumber: in.HoleNum,
		Teename:    in.TeeName,
		Playerid:   in.PlayerID,
	})
	if err != nil {
		return "", fmt.Errorf("query failed: %w", err)
	}

	if len(rows) == 0 {
		return fmt.Sprintf("No recorded rounds found for Played %d on hole %d at course %s.", in.PlayerID, in.HoleNum, in.CourseID), nil
	}

	out, err := json.Marshal(rows)
	if err != nil {
		return "", fmt.Errorf("marshal failed: %w", err)
	}
	return string(out), nil
}

func NewHoleStatsHandler(q *db.Queries) func(context.Context, json.RawMessage) (string, error) {
	h := &holeStatsHandler{queries: q}
	return h.handle
}
