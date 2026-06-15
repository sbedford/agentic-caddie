package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sbedford/agentic-caddie/internal/db"
)

type HolesHandler struct {
	Queries *db.Queries
}

type holeResponse struct {
	ID           int64   `json:"id"`
	RoundID      int64   `json:"round_id"`
	CourseHoleID int64   `json:"course_hole_id"`
	HoleNumber   int64   `json:"hole_number"`
	FlagPosition *string `json:"flag_position"`
	Score        *int64  `json:"score"`
	Points       *int64  `json:"points"`
	Putts        *int64  `json:"putts"`
	GIR          *bool   `json:"gir"`
	ScrambleSave *bool   `json:"scramble_save"`
	Penalty      *bool   `json:"penalty"`
}

type createHoleRequest struct {
	RoundID      int64   `json:"round_id"       example:"1"`
	CourseHoleID int64   `json:"course_hole_id" example:"1"`
	HoleNumber   int64   `json:"hole_number"    example:"1"`
	FlagPosition *string `json:"flag_position"  example:"middle_centre"`
	Score        *int64  `json:"score"          example:"4"`
	Points       *int64  `json:"points"         example:"2"`
	Putts        *int64  `json:"putts"          example:"2"`
	GIR          *bool   `json:"gir"            example:"true"`
	ScrambleSave *bool   `json:"scramble_save"  example:"false"`
	Penalty      *bool   `json:"penalty"        example:"false"`
}

type updateHoleRequest struct {
	FlagPosition *string `json:"flag_position" example:"middle_centre"`
	Score        *int64  `json:"score"         example:"4"`
	Points       *int64  `json:"points"        example:"2"`
	Putts        *int64  `json:"putts"         example:"2"`
	GIR          *bool   `json:"gir"           example:"true"`
	ScrambleSave *bool   `json:"scramble_save" example:"false"`
	Penalty      *bool   `json:"penalty"       example:"false"`
}

func toHoleResponse(h db.Hole) holeResponse {
	return holeResponse{
		ID:           h.ID,
		RoundID:      h.RoundID,
		CourseHoleID: h.CourseHoleID,
		HoleNumber:   h.HoleNumber,
		FlagPosition: nullableString(h.FlagPosition),
		Score:        nullableInt64(h.Score),
		Points:       nullableInt64(h.Points),
		Putts:        nullableInt64(h.Putts),
		GIR:          nullableBool(h.Gir),
		ScrambleSave: nullableBool(h.ScrambleSave),
		Penalty:      nullableBool(h.Penalty),
	}
}

// ListHolesByRound godoc
// @Summary     List holes for a round
// @Tags        holes
// @Produce     json
// @Param       roundId path     int true "Round ID"
// @Success     200     {array}  holeResponse
// @Failure     400     {string} string "Invalid round ID"
// @Failure     500     {string} string "Internal server error"
// @Router      /holes/round/{roundId} [get]
func (h HolesHandler) ListHolesByRound(w http.ResponseWriter, r *http.Request) {
	roundID, err := strconv.ParseInt(chi.URLParam(r, "roundId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid round ID", http.StatusBadRequest)
		return
	}

	holes, err := h.Queries.ListHolesByRound(r.Context(), roundID)
	if err != nil {
		log.Printf("failed to list holes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]holeResponse, len(holes))
	for i, hole := range holes {
		resp[i] = toHoleResponse(hole)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetHoleByID godoc
// @Summary     Get a hole by ID
// @Tags        holes
// @Produce     json
// @Param       id  path     int true "Hole ID"
// @Success     200 {object} holeResponse
// @Failure     400 {string} string "Invalid ID"
// @Failure     404 {string} string "Hole not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /holes/{id} [get]
func (h HolesHandler) GetHoleByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	hole, err := h.Queries.GetHoleByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Hole not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toHoleResponse(hole)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetHoleByRoundAndNumber godoc
// @Summary     Get a hole by round and hole number
// @Tags        holes
// @Produce     json
// @Param       roundId    path     int true "Round ID"
// @Param       holeNumber path     int true "Hole number"
// @Success     200        {object} holeResponse
// @Failure     400        {string} string "Invalid parameters"
// @Failure     404        {string} string "Hole not found"
// @Failure     500        {string} string "Internal server error"
// @Router      /holes/round/{roundId}/number/{holeNumber} [get]
func (h HolesHandler) GetHoleByRoundAndNumber(w http.ResponseWriter, r *http.Request) {
	roundID, err := strconv.ParseInt(chi.URLParam(r, "roundId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid round ID", http.StatusBadRequest)
		return
	}
	holeNumber, err := strconv.ParseInt(chi.URLParam(r, "holeNumber"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid hole number", http.StatusBadRequest)
		return
	}

	hole, err := h.Queries.GetHoleByRoundAndNumber(r.Context(), db.GetHoleByRoundAndNumberParams{
		RoundID:    roundID,
		HoleNumber: holeNumber,
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Hole not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toHoleResponse(hole)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreateHole godoc
// @Summary     Record a hole in a round
// @Tags        holes
// @Accept      json
// @Produce     json
// @Param       hole body     createHoleRequest true "Hole to record"
// @Success     201  {object} holeResponse
// @Failure     400  {string} string "Invalid request body"
// @Failure     500  {string} string "Internal server error"
// @Router      /holes [post]
func (h HolesHandler) CreateHole(w http.ResponseWriter, r *http.Request) {
	var req createHoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.RoundID == 0 || req.CourseHoleID == 0 || req.HoleNumber == 0 {
		http.Error(w, "round_id, course_hole_id and hole_number are required", http.StatusBadRequest)
		return
	}

	if !checkOptVocab(w, r.Context(), h.Queries, "flag_position", req.FlagPosition) {
		return
	}

	params := db.CreateHoleParams{
		RoundID:      req.RoundID,
		CourseHoleID: req.CourseHoleID,
		HoleNumber:   req.HoleNumber,
	}
	if req.FlagPosition != nil {
		params.FlagPosition = sql.NullString{String: *req.FlagPosition, Valid: true}
	}
	if req.Score != nil {
		params.Score = sql.NullInt64{Int64: *req.Score, Valid: true}
	}
	if req.Points != nil {
		params.Points = sql.NullInt64{Int64: *req.Points, Valid: true}
	}
	if req.Putts != nil {
		params.Putts = sql.NullInt64{Int64: *req.Putts, Valid: true}
	}
	if req.GIR != nil {
		params.Gir = sql.NullBool{Bool: *req.GIR, Valid: true}
	}
	if req.ScrambleSave != nil {
		params.ScrambleSave = sql.NullBool{Bool: *req.ScrambleSave, Valid: true}
	}
	if req.Penalty != nil {
		params.Penalty = sql.NullBool{Bool: *req.Penalty, Valid: true}
	}

	result, err := h.Queries.CreateHole(r.Context(), params)
	if err != nil {
		log.Printf("failed to create hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	hole, err := h.Queries.GetHoleByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to fetch created hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toHoleResponse(hole))
}

// UpdateHole godoc
// @Summary     Update a hole's scorecard data
// @Tags        holes
// @Accept      json
// @Param       id   path int             true "Hole ID"
// @Param       hole body updateHoleRequest true "Updated fields"
// @Success     204
// @Failure     400 {string} string "Invalid request"
// @Failure     500 {string} string "Internal server error"
// @Router      /holes/{id} [put]
func (h HolesHandler) UpdateHole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req updateHoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !checkOptVocab(w, r.Context(), h.Queries, "flag_position", req.FlagPosition) {
		return
	}

	params := db.UpdateHoleParams{ID: id}
	if req.FlagPosition != nil {
		params.FlagPosition = sql.NullString{String: *req.FlagPosition, Valid: true}
	}
	if req.Score != nil {
		params.Score = sql.NullInt64{Int64: *req.Score, Valid: true}
	}
	if req.Points != nil {
		params.Points = sql.NullInt64{Int64: *req.Points, Valid: true}
	}
	if req.Putts != nil {
		params.Putts = sql.NullInt64{Int64: *req.Putts, Valid: true}
	}
	if req.GIR != nil {
		params.Gir = sql.NullBool{Bool: *req.GIR, Valid: true}
	}
	if req.ScrambleSave != nil {
		params.ScrambleSave = sql.NullBool{Bool: *req.ScrambleSave, Valid: true}
	}
	if req.Penalty != nil {
		params.Penalty = sql.NullBool{Bool: *req.Penalty, Valid: true}
	}

	if err = h.Queries.UpdateHole(r.Context(), params); err != nil {
		log.Printf("failed to update hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteHole godoc
// @Summary     Delete a hole record
// @Tags        holes
// @Param       id path int true "Hole ID"
// @Success     204
// @Failure     400 {string} string "Invalid ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /holes/{id} [delete]
func (h HolesHandler) DeleteHole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeleteHole(r.Context(), id); err != nil {
		log.Printf("failed to delete hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
