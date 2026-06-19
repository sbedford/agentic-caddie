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
	"github.com/sbedford/agentic-caddie/internal/helpers"
)

type ShotsHandler struct {
	Queries *db.Queries
	DB      *sql.DB
}

type reorderShotsRequest struct {
	IDs []int64 `json:"ids"`
}

type shotResponse struct {
	ID                    int64   `json:"id"`
	HoleID                int64   `json:"hole_id"`
	ShotNumber            int64   `json:"shot_number"`
	ShotType              string  `json:"shot_type"`
	Club                  *string `json:"club"`
	Result                *string `json:"result"`
	Miss                  *string `json:"miss"`
	StrikeQuality         *string `json:"strike_quality"`
	Source                string  `json:"source"`
	PreShotRecommendation *string `json:"pre_shot_recommendation"`
	Completed             *bool   `json:"completed"`
}

type createShotRequest struct {
	HoleID                int64   `json:"hole_id"                 example:"1"`
	ShotNumber            int64   `json:"shot_number"             example:"1"`
	ShotType              string  `json:"shot_type"               example:"tee"`
	Club                  *string `json:"club"                    example:"driver"`
	Result                *string `json:"result"                  example:"fairway"`
	Miss                  *string `json:"miss"`
	StrikeQuality         *string `json:"strike_quality"          example:"clean"`
	Source                string  `json:"source"                  example:"manual"`
	PreShotRecommendation *string `json:"pre_shot_recommendation"`
	Completed             *bool   `json:"completed"`
}

type updateShotRequest struct {
	ShotType              string  `json:"shot_type"               example:"tee"`
	Club                  *string `json:"club"                    example:"driver"`
	Result                *string `json:"result"                  example:"fairway"`
	Miss                  *string `json:"miss"`
	StrikeQuality         *string `json:"strike_quality"          example:"clean"`
	Source                string  `json:"source"                  example:"manual"`
	PreShotRecommendation *string `json:"pre_shot_recommendation"`
	Completed             *bool   `json:"completed"`
}

func toShotResponse(s db.Shot) shotResponse {
	return shotResponse{
		ID:                    s.ID,
		HoleID:                s.HoleID,
		ShotNumber:            s.ShotNumber,
		ShotType:              s.ShotType,
		Club:                  helpers.NullableString(s.Club),
		Result:                helpers.NullableString(s.Result),
		Miss:                  helpers.NullableString(s.Miss),
		StrikeQuality:         helpers.NullableString(s.StrikeQuality),
		Source:                s.Source,
		PreShotRecommendation: helpers.NullableString(s.PreShotRecommendation),
		Completed:             helpers.NullableBool(s.Completed),
	}
}

func toShotResponseFromHoleAndNumber(s db.Shot) shotResponse {
	return shotResponse{
		ID:                    s.ID,
		HoleID:                s.HoleID,
		ShotNumber:            s.ShotNumber,
		ShotType:              s.ShotType,
		Club:                  helpers.NullableString(s.Club),
		Result:                helpers.NullableString(s.Result),
		Miss:                  helpers.NullableString(s.Miss),
		StrikeQuality:         helpers.NullableString(s.StrikeQuality),
		Source:                s.Source,
		PreShotRecommendation: helpers.NullableString(s.PreShotRecommendation),
		Completed:             helpers.NullableBool(s.Completed),
	}
}

func toShotResponseFromListRow(s db.Shot) shotResponse {
	return shotResponse{
		ID:                    s.ID,
		HoleID:                s.HoleID,
		ShotNumber:            s.ShotNumber,
		ShotType:              s.ShotType,
		Club:                  helpers.NullableString(s.Club),
		Result:                helpers.NullableString(s.Result),
		Miss:                  helpers.NullableString(s.Miss),
		StrikeQuality:         helpers.NullableString(s.StrikeQuality),
		Source:                s.Source,
		PreShotRecommendation: helpers.NullableString(s.PreShotRecommendation),
		Completed:             helpers.NullableBool(s.Completed),
	}
}

func toShotResponseFromListTypeRow(s db.Shot) shotResponse {
	return shotResponse{
		ID:                    s.ID,
		HoleID:                s.HoleID,
		ShotNumber:            s.ShotNumber,
		ShotType:              s.ShotType,
		Club:                  helpers.NullableString(s.Club),
		Result:                helpers.NullableString(s.Result),
		Miss:                  helpers.NullableString(s.Miss),
		StrikeQuality:         helpers.NullableString(s.StrikeQuality),
		Source:                s.Source,
		PreShotRecommendation: helpers.NullableString(s.PreShotRecommendation),
		Completed:             helpers.NullableBool(s.Completed),
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

	shots, err := h.Queries.ListShotsByHole(r.Context(), holeID)
	if err != nil {
		log.Printf("failed to list shots: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]shotResponse, len(shots))
	for i, s := range shots {
		resp[i] = toShotResponseFromListRow(s)
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

	shots, err := h.Queries.ListShotsByHoleAndType(r.Context(), db.ListShotsByHoleAndTypeParams{
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
		resp[i] = toShotResponseFromListTypeRow(s)
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

	shot, err := h.Queries.GetShotByID(r.Context(), id)
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

	shot, err := h.Queries.GetShotByHoleAndNumber(r.Context(), db.GetShotByHoleAndNumberParams{
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
	if err = json.NewEncoder(w).Encode(toShotResponseFromHoleAndNumber(shot)); err != nil {
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

	if !helpers.CheckVocab(w, r.Context(), h.Queries, "shot_type", req.ShotType) {
		return
	}
	if !helpers.CheckVocab(w, r.Context(), h.Queries, "shot_source", req.Source) {
		return
	}
	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "shot_result", req.Result) {
		return
	}
	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "shot_miss", req.Miss) {
		return
	}
	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "shot_strike", req.StrikeQuality) {
		return
	}

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
	if req.PreShotRecommendation != nil {
		params.PreShotRecommendation = sql.NullString{String: *req.PreShotRecommendation, Valid: true}
	}
	if req.Completed != nil {
		params.Completed = sql.NullBool{Bool: *req.Completed, Valid: true}
	}

	result, err := h.Queries.CreateShot(r.Context(), params)
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

	shot, err := h.Queries.GetShotByID(r.Context(), id)
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

	if !helpers.CheckVocab(w, r.Context(), h.Queries, "shot_type", req.ShotType) {
		return
	}
	if !helpers.CheckVocab(w, r.Context(), h.Queries, "shot_source", req.Source) {
		return
	}
	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "shot_result", req.Result) {
		return
	}
	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "shot_miss", req.Miss) {
		return
	}
	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "shot_strike", req.StrikeQuality) {
		return
	}

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
	if req.PreShotRecommendation != nil {
		params.PreShotRecommendation = sql.NullString{String: *req.PreShotRecommendation, Valid: true}
	}
	if req.Completed != nil {
		params.Completed = sql.NullBool{Bool: *req.Completed, Valid: true}
	}

	if err = h.Queries.UpdateShot(r.Context(), params); err != nil {
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

	if err = h.Queries.DeleteShot(r.Context(), id); err != nil {
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

	if err = h.Queries.DeleteShotsByHole(r.Context(), holeID); err != nil {
		log.Printf("failed to delete shots: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ReorderShots godoc
// @Summary     Reorder shots for a hole
// @Tags        shots
// @Accept      json
// @Param       holeId path int                true "Hole ID"
// @Param       body   body reorderShotsRequest true "Ordered shot IDs"
// @Success     204
// @Failure     400 {string} string "Invalid request"
// @Failure     500 {string} string "Internal server error"
// @Router      /shots/hole/{holeId}/reorder [post]
func (h ShotsHandler) ReorderShots(w http.ResponseWriter, r *http.Request) {
	holeID, err := strconv.ParseInt(chi.URLParam(r, "holeId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid hole ID", http.StatusBadRequest)
		return
	}

	var req reorderShotsRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.IDs) == 0 {
		http.Error(w, "ids must not be empty", http.StatusBadRequest)
		return
	}

	tx, err := h.DB.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("failed to begin transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback() //nolint:errcheck

	// Shift all shot_numbers to a high range to avoid unique constraint conflicts
	if _, err = tx.ExecContext(r.Context(),
		"UPDATE shots SET shot_number = shot_number + 10000 WHERE hole_id = ?", holeID); err != nil {
		log.Printf("failed to shift shot numbers: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Assign final positions in the requested order
	for i, id := range req.IDs {
		if _, err = tx.ExecContext(r.Context(),
			"UPDATE shots SET shot_number = ? WHERE id = ? AND hole_id = ?",
			i+1, id, holeID); err != nil {
			log.Printf("failed to assign shot number: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("failed to commit reorder transaction: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
