package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/helpers"
)

var GetHoleLayoutToolDef = anthropic.ToolUnionParam{
	OfTool: &anthropic.ToolParam{
		Name:        "get_hole_layout",
		Description: anthropic.String("Returns data points for a hole on a specific course: par, stroke index, distance and a series of hole Layout. Use this to look for specific details on the hole to base your advice."),
		InputSchema: anthropic.ToolInputSchemaParam{
			Properties: map[string]any{
				"course_id": map[string]any{
					"type":        "integer",
					"description": "The course identifier",
				},
				"tee_name": map[string]any{
					"type":        "string",
					"description": "Tee name, typically Black or White",
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

type getHoleLayoutInput struct {
	CourseID int64  `json:"course_id"`
	TeeName  string `json:"tee_name"`
	HoleNum  int64  `json:"hole_num"`
}

type getHoleLayoutFeatures struct {
	PoiType        string
	Side           string
	ReferencePoint string
	DistanceStart  float64
	DistanceEnd    float64
	Label          string
}

func buildListPOIsByHoleAndTeeParams(pois []db.HolePointsOfInterest) []getHoleLayoutFeatures {

	output := make([]getHoleLayoutFeatures, len(pois))

	for i, poi := range pois {
		output[i] = getHoleLayoutFeatures{
			PoiType:        poi.PoiType,
			Side:           helpers.String(poi.Side),
			ReferencePoint: helpers.String(poi.ReferencePoint),
			DistanceStart:  helpers.Float64(poi.DistanceStart),
			DistanceEnd:    helpers.Float64(poi.DistanceEnd),
			Label:          poi.Label,
		}
	}
	return output
}

type getHoleLayoutOutput struct {
	Par         int64                   `json:"par"`
	StrokeIndex int64                   `json:"stroke_index"`
	Distance    int64                   `json:"distance"`
	Features    []getHoleLayoutFeatures `json:"features"`
}

type holeLayoutHandler struct {
	queries *db.Queries
}

// type ToolHandler func(ctx context.Context, input json.RawMessage) (string, error)
func (h *holeLayoutHandler) handle(ctx context.Context, raw json.RawMessage) (string, error) {
	var in getHoleLayoutInput
	if err := json.Unmarshal(raw, &in); err != nil {
		return "", fmt.Errorf("invalid input: %w", err)
	}
	if in.HoleNum < 1 || in.HoleNum > 18 {
		return "", fmt.Errorf("hole_num must be between 1 and 18, got %d", in.HoleNum)
	}

	hole, err := h.queries.GetTeeHoleByCourseIdAndHoleAndTeeName(ctx, db.GetTeeHoleByCourseIdAndHoleAndTeeNameParams{
		Courseid:   in.CourseID,
		Holenumber: in.HoleNum,
		Teename:    in.TeeName,
	})

	if err != nil {
		return "", fmt.Errorf("Error in db.GetTeeHoleByCourseIdAndHoleAndTeeNameParams: %w", err)
	}

	features, err := h.queries.ListPOIsByHoleAndTee(ctx, db.ListPOIsByHoleAndTeeParams{
		CourseHoleID: hole.CourseHoleID,
		SpecificTee: sql.NullString{
			String: in.TeeName,
		},
	})

	if err != nil {
		return "", fmt.Errorf("Error in db.ListPOIsByHoleAndTeeParams: %w", err)
	}

	layout := getHoleLayoutOutput{
		Par:         hole.Par,
		StrokeIndex: hole.StrokeIndex.Int64,
		Distance:    hole.Distance,
		Features:    buildListPOIsByHoleAndTeeParams(features),
	}

	// Return raw rows as JSON — let the model do the reasoning over them,
	// don't average/aggregate here.
	out, err := json.Marshal(layout)
	if err != nil {
		return "", fmt.Errorf("marshal failed: %w", err)
	}
	return string(out), nil
}

// in RoundHistoryTool.go (or wherever the round history tool lives)
func NewHoleLayoutHandler(q *db.Queries) func(context.Context, json.RawMessage) (string, error) {
	h := &holeLayoutHandler{queries: q}
	return h.handle
}
