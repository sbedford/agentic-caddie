package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sbedford/agentic-caddie/internal/db"
	"github.com/sbedford/agentic-caddie/internal/helpers"
)

type RoundsHandler struct {
	Queries *db.Queries
}

type roundResponse struct {
	ID              int64     `json:"id"`
	PlayerID        int64     `json:"player_id"`
	CourseID        int64     `json:"course_id"`
	PlayedAt        time.Time `json:"played_at"`
	Tees            string    `json:"tees"`
	DailyHandicap   int64     `json:"daily_handicap"`
	RoundType       string    `json:"round_type"`
	CompetitionType *string   `json:"competition_type"`
	TotalScore      *int64    `json:"total_score"`
	TotalPoints     *int64    `json:"total_points"`
	TotalPutts      *int64    `json:"total_putts"`
	CreatedAt       time.Time `json:"created_at"`
}

type createRoundRequest struct {
	PlayerID        int64  `json:"player_id"       example:"1"`
	CourseID        int64  `json:"course_id"       example:"1"`
	PlayedAt        string `json:"played_at"       example:"2024-06-01"`
	Tees            string `json:"tees"            example:"white"`
	DailyHandicap   int64  `json:"daily_handicap"  example:"18"`
	RoundType       string `json:"round_type"      example:"social"`
	CompetitionType string `json:"competition_type" example:"stableford"`
}

func toRoundResponse(r db.Round) roundResponse {
	return roundResponse{
		ID:              r.ID,
		PlayerID:        r.PlayerID,
		CourseID:        r.CourseID,
		PlayedAt:        r.PlayedAt,
		Tees:            r.Tees,
		DailyHandicap:   r.DailyHandicap,
		RoundType:       r.RoundType,
		CompetitionType: helpers.NullableString(r.CompetitionType),
		TotalScore:      helpers.NullableInt64(r.TotalScore),
		TotalPoints:     helpers.NullableInt64(r.TotalPoints),
		TotalPutts:      helpers.NullableInt64(r.TotalPutts),
		CreatedAt:       r.CreatedAt,
	}
}

// ListRounds godoc
// @Summary     List all rounds
// @Tags        rounds
// @Produce     json
// @Success     200      {array}  roundResponse
// @Failure     400      {string} string "Invalid player ID"
// @Failure     500      {string} string "Internal server error"
// @Router      /rounds [get]
func (h RoundsHandler) ListRounds(w http.ResponseWriter, r *http.Request) {

	rounds, err := h.Queries.ListRounds(r.Context())
	if err != nil {
		log.Printf("failed to list rounds: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]roundResponse, len(rounds))
	for i, rnd := range rounds {
		resp[i] = toRoundResponse(rnd)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// ListRoundsByPlayer godoc
// @Summary     List rounds for a player
// @Tags        rounds
// @Produce     json
// @Param       playerId path     int true "Player ID"
// @Success     200      {array}  roundResponse
// @Failure     400      {string} string "Invalid player ID"
// @Failure     500      {string} string "Internal server error"
// @Router      /rounds/player/{playerId} [get]
func (h RoundsHandler) ListRoundsByPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.ParseInt(chi.URLParam(r, "playerId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	rounds, err := h.Queries.ListRoundsByPlayer(r.Context(), playerID)
	if err != nil {
		log.Printf("failed to list rounds: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]roundResponse, len(rounds))
	for i, rnd := range rounds {
		resp[i] = toRoundResponse(rnd)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// ListRoundsByPlayerAndCourse godoc
// @Summary     List rounds for a player at a specific course
// @Tags        rounds
// @Produce     json
// @Param       playerId path     int true "Player ID"
// @Param       courseId path     int true "Course ID"
// @Success     200      {array}  roundResponse
// @Failure     400      {string} string "Invalid parameters"
// @Failure     500      {string} string "Internal server error"
// @Router      /rounds/player/{playerId}/course/{courseId} [get]
func (h RoundsHandler) ListRoundsByPlayerAndCourse(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.ParseInt(chi.URLParam(r, "playerId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	rounds, err := h.Queries.ListRoundsByPlayerAndCourse(r.Context(), db.ListRoundsByPlayerAndCourseParams{
		PlayerID: playerID,
		CourseID: courseID,
	})
	if err != nil {
		log.Printf("failed to list rounds: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]roundResponse, len(rounds))
	for i, rnd := range rounds {
		resp[i] = toRoundResponse(rnd)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetRoundByID godoc
// @Summary     Get a round by ID
// @Tags        rounds
// @Produce     json
// @Param       id  path     int true "Round ID"
// @Success     200 {object} roundResponse
// @Failure     400 {string} string "Invalid ID"
// @Failure     404 {string} string "Round not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /rounds/{id} [get]
func (h RoundsHandler) GetRoundByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	round, err := h.Queries.GetRoundByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Round not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get round: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toRoundResponse(round)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetRoundByPlayerAndDate godoc
// @Summary     Get a round by player and date played
// @Tags        rounds
// @Produce     json
// @Param       playerId path     int    true "Player ID"
// @Param       date     path     string true "Date (YYYY-MM-DD)"
// @Success     200      {object} roundResponse
// @Failure     400      {string} string "Invalid parameters"
// @Failure     404      {string} string "Round not found"
// @Failure     500      {string} string "Internal server error"
// @Router      /rounds/player/{playerId}/date/{date} [get]
func (h RoundsHandler) GetRoundByPlayerAndDate(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.ParseInt(chi.URLParam(r, "playerId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}
	playedAt, err := time.Parse("2006-01-02", chi.URLParam(r, "date"))
	if err != nil {
		http.Error(w, "Invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	round, err := h.Queries.GetRoundByPlayerAndDate(r.Context(), db.GetRoundByPlayerAndDateParams{
		PlayerID: playerID,
		PlayedAt: playedAt,
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Round not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get round: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toRoundResponse(round)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreateRound godoc
// @Summary     Create a round
// @Tags        rounds
// @Accept      json
// @Produce     json
// @Param       round body     createRoundRequest true "Round to create"
// @Success     201   {object} roundResponse
// @Failure     400   {string} string "Invalid request body"
// @Failure     500   {string} string "Internal server error"
// @Router      /rounds [post]
func (h RoundsHandler) CreateRound(w http.ResponseWriter, r *http.Request) {
	var req createRoundRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.PlayerID == 0 || req.CourseID == 0 || req.PlayedAt == "" || req.Tees == "" || req.RoundType == "" {
		http.Error(w, "player_id, course_id, played_at, tees and round_type are required", http.StatusBadRequest)
		return
	}
	if req.DailyHandicap < 0 {
		http.Error(w, "daily_handicap must be 0 or greater", http.StatusBadRequest)
		return
	}

	playedAt, err := time.Parse("2006-01-02", req.PlayedAt)
	if err != nil {
		http.Error(w, "Invalid played_at format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	if !helpers.CheckVocab(w, r.Context(), h.Queries, "round_type", req.RoundType) {
		return
	}
	if req.CompetitionType != "" && !helpers.CheckVocab(w, r.Context(), h.Queries, "competition_type", req.CompetitionType) {
		return
	}

	params := db.CreateRoundParams{
		PlayerID:      req.PlayerID,
		CourseID:      req.CourseID,
		PlayedAt:      playedAt,
		Tees:          req.Tees,
		DailyHandicap: req.DailyHandicap,
		RoundType:     req.RoundType,
	}
	if req.CompetitionType != "" {
		params.CompetitionType = sql.NullString{String: req.CompetitionType, Valid: true}
	}

	result, err := h.Queries.CreateRound(r.Context(), params)
	if err != nil {
		log.Printf("failed to create round: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	round, err := h.Queries.GetRoundByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to fetch created round: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toRoundResponse(round))
}

// UpdateRoundTotals godoc
// @Summary     Recalculate and store round totals from hole data
// @Tags        rounds
// @Param       id path int true "Round ID"
// @Success     204
// @Failure     400 {string} string "Invalid ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /rounds/{id}/totals [patch]
func (h RoundsHandler) UpdateRoundTotals(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.UpdateRoundTotals(r.Context(), id); err != nil {
		log.Printf("failed to update round totals: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteRound godoc
// @Summary     Delete a round
// @Tags        rounds
// @Param       id path int true "Round ID"
// @Success     204
// @Failure     400 {string} string "Invalid ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /rounds/{id} [delete]
func (h RoundsHandler) DeleteRound(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeleteRound(r.Context(), id); err != nil {
		log.Printf("failed to delete round: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
