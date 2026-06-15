package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbedford/agentic-caddie/internal/db"
)

type ShotsHandler struct {
	Queries *db.Queries
}

type shotResponse struct {
	ID            int64   `json:"id"`
	HoleID        int64   `json:"hole_id"`
	ShotNumber    int64   `json:"shot_number"`
	ShotType      string  `json:"shot_type"`
	Club          *string `json:"club"`
	Result        *string `json:"result"`
	Miss          *string `json:"miss"`
	StrikeQuality *string `json:"strike_quality"`
	Source        string  `json:"source"`
}

type createShotRequest struct {
	HoleID        int64   `json:"hole_id"        example:"1"`
	ShotNumber    int64   `json:"shot_number"    example:"1"`
	ShotType      string  `json:"shot_type"      example:"tee"`
	Club          *string `json:"club"           example:"driver"`
	Result        *string `json:"result"         example:"fairway"`
	Miss          *string `json:"miss"`
	StrikeQuality *string `json:"strike_quality" example:"clean"`
	Source        string  `json:"source"         example:"manual"`
}

type updateShotRequest struct {
	ShotType      string  `json:"shot_type"      example:"tee"`
	Club          *string `json:"club"           example:"driver"`
	Result        *string `json:"result"         example:"fairway"`
	Miss          *string `json:"miss"`
	StrikeQuality *string `json:"strike_quality" example:"clean"`
	Source        string  `json:"source"         example:"manual"`
}

func toShotResponse(s db.Shot) shotResponse {
	return shotResponse{
		ID:            s.ID,
		HoleID:        s.HoleID,
		ShotNumber:    s.ShotNumber,
		ShotType:      s.ShotType,
		Club:          nullableString(s.Club),
		Result:        nullableString(s.Result),
		Miss:          nullableString(s.Miss),
		StrikeQuality: nullableString(s.StrikeQuality),
		Source:        s.Source,
	}
}

// ListShotsByHole godoc
// @Summary     List shots for a hole
// @Tags        shots
// @Produce     json
// @Param       holeId path     int true "Hole ID"
// @Success     200    {array}  shotResponse
// @Failure     400    {string} string "Invalid hole ID"
// @Failure     500    {string} string "Internal server error"
// @Router      /shots/hole/{holeId} [get]
func (h ShotsHandler) ListShotsByHole(w http.ResponseWriter, r *http.Request) {
	holeID, err := strconv.ParseInt(chi.URLParam(r, "holeId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid hole ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	shots, err := h.Queries.ListShotsByHole(ctx, holeID)
	if err != nil {
		log.Printf("failed to list shots: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]shotResponse, len(shots))
	for i, s := range shots {
		resp[i] = toShotResponse(s)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// ListShotsByHoleAndType godoc
// @Summary     List shots for a hole filtered by shot type
// @Tags        shots
// @Produce     json
// @Param       holeId   path     int    true "Hole ID"
// @Param       shotType path     string true "Shot type"
// @Success     200      {array}  shotResponse
// @Failure     400      {string} string "Invalid hole ID"
// @Failure     500      {string} string "Internal server error"
// @Router      /shots/hole/{holeId}/type/{shotType} [get]
func (h ShotsHandler) ListShotsByHoleAndType(w http.ResponseWriter, r *http.Request) {
	holeID, err := strconv.ParseInt(chi.URLParam(r, "holeId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid hole ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	shots, err := h.Queries.ListShotsByHoleAndType(ctx, db.ListShotsByHoleAndTypeParams{
		HoleID:   holeID,
		ShotType: chi.URLParam(r, "shotType"),
	})
	if err != nil {
		log.Printf("failed to list shots by type: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]shotResponse, len(shots))
	for i, s := range shots {
		resp[i] = toShotResponse(s)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetShotByID godoc
// @Summary     Get a shot by ID
// @Tags        shots
// @Produce     json
// @Param       id  path     int true "Shot ID"
// @Success     200 {object} shotResponse
// @Failure     400 {string} string "Invalid ID"
// @Failure     404 {string} string "Shot not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /shots/{id} [get]
func (h ShotsHandler) GetShotByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	shot, err := h.Queries.GetShotByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Shot not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get shot: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toShotResponse(shot)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetShotByHoleAndNumber godoc
// @Summary     Get a shot by hole and shot number
// @Tags        shots
// @Produce     json
// @Param       holeId     path     int true "Hole ID"
// @Param       shotNumber path     int true "Shot number"
// @Success     200        {object} shotResponse
// @Failure     400        {string} string "Invalid parameters"
// @Failure     404        {string} string "Shot not found"
// @Failure     500        {string} string "Internal server error"
// @Router      /shots/hole/{holeId}/number/{shotNumber} [get]
func (h ShotsHandler) GetShotByHoleAndNumber(w http.ResponseWriter, r *http.Request) {
	holeID, err := strconv.ParseInt(chi.URLParam(r, "holeId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid hole ID", http.StatusBadRequest)
		return
	}
	shotNumber, err := strconv.ParseInt(chi.URLParam(r, "shotNumber"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid shot number", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	shot, err := h.Queries.GetShotByHoleAndNumber(ctx, db.GetShotByHoleAndNumberParams{
		HoleID:     holeID,
		ShotNumber: shotNumber,
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Shot not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get shot: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toShotResponse(shot)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreateShot godoc
// @Summary     Record a shot
// @Tags        shots
// @Accept      json
// @Produce     json
// @Param       shot body     createShotRequest true "Shot to record"
// @Success     201  {object} shotResponse
// @Failure     400  {string} string "Invalid request body"
// @Failure     500  {string} string "Internal server error"
// @Router      /shots [post]
func (h ShotsHandler) CreateShot(w http.ResponseWriter, r *http.Request) {
	var req createShotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.HoleID == 0 || req.ShotNumber == 0 || req.ShotType == "" {
		http.Error(w, "hole_id, shot_number and shot_type are required", http.StatusBadRequest)
		return
	}
	if req.Source == "" {
		req.Source = "manual"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	params := db.CreateShotParams{
		HoleID:     req.HoleID,
		ShotNumber: req.ShotNumber,
		ShotType:   req.ShotType,
		Source:     req.Source,
	}
	if req.Club != nil {
		params.Club = sql.NullString{String: *req.Club, Valid: true}
	}
	if req.Result != nil {
		params.Result = sql.NullString{String: *req.Result, Valid: true}
	}
	if req.Miss != nil {
		params.Miss = sql.NullString{String: *req.Miss, Valid: true}
	}
	if req.StrikeQuality != nil {
		params.StrikeQuality = sql.NullString{String: *req.StrikeQuality, Valid: true}
	}

	result, err := h.Queries.CreateShot(ctx, params)
	if err != nil {
		log.Printf("failed to create shot: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	shot, err := h.Queries.GetShotByID(ctx, id)
	if err != nil {
		log.Printf("failed to fetch created shot: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toShotResponse(shot))
}

// UpdateShot godoc
// @Summary     Update a shot
// @Tags        shots
// @Accept      json
// @Param       id   path int             true "Shot ID"
// @Param       shot body updateShotRequest true "Updated fields"
// @Success     204
// @Failure     400 {string} string "Invalid request"
// @Failure     500 {string} string "Internal server error"
// @Router      /shots/{id} [put]
func (h ShotsHandler) UpdateShot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req updateShotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	params := db.UpdateShotParams{ID: id, ShotType: req.ShotType, Source: req.Source}
	if req.Club != nil {
		params.Club = sql.NullString{String: *req.Club, Valid: true}
	}
	if req.Result != nil {
		params.Result = sql.NullString{String: *req.Result, Valid: true}
	}
	if req.Miss != nil {
		params.Miss = sql.NullString{String: *req.Miss, Valid: true}
	}
	if req.StrikeQuality != nil {
		params.StrikeQuality = sql.NullString{String: *req.StrikeQuality, Valid: true}
	}

	if err = h.Queries.UpdateShot(ctx, params); err != nil {
		log.Printf("failed to update shot: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteShot godoc
// @Summary     Delete a shot
// @Tags        shots
// @Param       id path int true "Shot ID"
// @Success     204
// @Failure     400 {string} string "Invalid ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /shots/{id} [delete]
func (h ShotsHandler) DeleteShot(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err = h.Queries.DeleteShot(ctx, id); err != nil {
		log.Printf("failed to delete shot: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteShotsByHole godoc
// @Summary     Delete all shots for a hole
// @Tags        shots
// @Param       holeId path int true "Hole ID"
// @Success     204
// @Failure     400 {string} string "Invalid hole ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /shots/hole/{holeId} [delete]
func (h ShotsHandler) DeleteShotsByHole(w http.ResponseWriter, r *http.Request) {
	holeID, err := strconv.ParseInt(chi.URLParam(r, "holeId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid hole ID", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if err = h.Queries.DeleteShotsByHole(ctx, holeID); err != nil {
		log.Printf("failed to delete shots: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
