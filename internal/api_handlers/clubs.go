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

type ClubsHandler struct {
	Queries *db.Queries
}

type clubResponse struct {
	ID             int64      `json:"id"`
	PlayerID       int64      `json:"player_id"`
	ClubName       string     `json:"club_name"`
	AddedDate      time.Time  `json:"added_date"`
	RemovedDate    *time.Time `json:"removed_date"`
	CarryAvg       *float64   `json:"carry_avg"`
	CarryReliable  *float64   `json:"carry_reliable"`
	CarryMax       *float64   `json:"carry_max"`
	DispersionAvgM *float64   `json:"dispersion_avg_m"`
	DispersionBias *string    `json:"dispersion_bias"`
	SampleSize     int64      `json:"sample_size"`
	CalculatedAt   *time.Time `json:"calculated_at"`
}

type createClubRequest struct {
	PlayerID       int64    `json:"player_id"        example:"1"`
	ClubName       string   `json:"club_name"        example:"7i"`
	AddedDate      string   `json:"added_date"       example:"2024-01-01"`
	CarryAvg       *float64 `json:"carry_avg"        example:"155.0"`
	CarryReliable  *float64 `json:"carry_reliable"   example:"148.0"`
	CarryMax       *float64 `json:"carry_max"        example:"165.0"`
	DispersionAvgM *float64 `json:"dispersion_avg_m" example:"8.5"`
	DispersionBias *string  `json:"dispersion_bias"  example:"straight"`
	SampleSize     int64    `json:"sample_size"      example:"0"`
}

type retireClubRequest struct {
	RemovedDate string `json:"removed_date" example:"2024-06-01"`
}

type updateClubDistancesRequest struct {
	CarryAvg       *float64 `json:"carry_avg"        example:"155.0"`
	CarryReliable  *float64 `json:"carry_reliable"   example:"148.0"`
	CarryMax       *float64 `json:"carry_max"        example:"165.0"`
	DispersionAvgM *float64 `json:"dispersion_avg_m" example:"8.5"`
	DispersionBias *string  `json:"dispersion_bias"  example:"straight"`
	SampleSize     int64    `json:"sample_size"      example:"12"`
}

func toClubResponse(c db.PlayerClub) clubResponse {
	return clubResponse{
		ID:             c.ID,
		PlayerID:       c.PlayerID,
		ClubName:       c.ClubName,
		AddedDate:      c.AddedDate,
		RemovedDate:    helpers.NullableTime(c.RemovedDate),
		CarryAvg:       helpers.NullableFloat64(c.CarryAvg),
		CarryReliable:  helpers.NullableFloat64(c.CarryReliable),
		CarryMax:       helpers.NullableFloat64(c.CarryMax),
		DispersionAvgM: helpers.NullableFloat64(c.DispersionAvgM),
		DispersionBias: helpers.NullableString(c.DispersionBias),
		SampleSize:     c.SampleSize,
		CalculatedAt:   helpers.NullableTime(c.CalculatedAt),
	}
}

// ListClubsByPlayer godoc
// @Summary     List all clubs for a player
// @Tags        clubs
// @Produce     json
// @Param       playerId path     int true "Player ID"
// @Success     200      {array}  clubResponse
// @Failure     400      {string} string "Invalid player ID"
// @Failure     500      {string} string "Internal server error"
// @Router      /clubs/player/{playerId} [get]
func (h ClubsHandler) ListClubsByPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.ParseInt(chi.URLParam(r, "playerId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	clubs, err := h.Queries.ListClubsByPlayer(r.Context(), playerID)
	if err != nil {
		log.Printf("failed to list clubs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]clubResponse, len(clubs))
	for i, c := range clubs {
		resp[i] = toClubResponse(c)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// ListActiveClubsByPlayer godoc
// @Summary     List active (in-bag) clubs for a player
// @Tags        clubs
// @Produce     json
// @Param       playerId path     int true "Player ID"
// @Success     200      {array}  clubResponse
// @Failure     400      {string} string "Invalid player ID"
// @Failure     500      {string} string "Internal server error"
// @Router      /clubs/player/{playerId}/active [get]
func (h ClubsHandler) ListActiveClubsByPlayer(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.ParseInt(chi.URLParam(r, "playerId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	clubs, err := h.Queries.ListActiveClubsByPlayer(r.Context(), playerID)
	if err != nil {
		log.Printf("failed to list active clubs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]clubResponse, len(clubs))
	for i, c := range clubs {
		resp[i] = toClubResponse(c)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetClubByID godoc
// @Summary     Get a club by ID
// @Tags        clubs
// @Produce     json
// @Param       id  path     int true "Club ID"
// @Success     200 {object} clubResponse
// @Failure     400 {string} string "Invalid ID"
// @Failure     404 {string} string "Club not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /clubs/{id} [get]
func (h ClubsHandler) GetClubByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	club, err := h.Queries.GetClubByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Club not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get club: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toClubResponse(club)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetClubByPlayerAndName godoc
// @Summary     Get the active club for a player by club name
// @Tags        clubs
// @Produce     json
// @Param       playerId  path     int    true "Player ID"
// @Param       clubName  path     string true "Club name"
// @Success     200       {object} clubResponse
// @Failure     400       {string} string "Invalid player ID"
// @Failure     404       {string} string "Club not found"
// @Failure     500       {string} string "Internal server error"
// @Router      /clubs/player/{playerId}/name/{clubName} [get]
func (h ClubsHandler) GetClubByPlayerAndName(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.ParseInt(chi.URLParam(r, "playerId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	club, err := h.Queries.GetClubByPlayerAndName(r.Context(), db.GetClubByPlayerAndNameParams{
		PlayerID: playerID,
		ClubName: chi.URLParam(r, "clubName"),
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Club not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get club: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toClubResponse(club)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetClubByPlayerNameAndDate godoc
// @Summary     Get a club as it existed on a specific date
// @Tags        clubs
// @Produce     json
// @Param       playerId  path     int    true "Player ID"
// @Param       clubName  path     string true "Club name"
// @Param       date      path     string true "Date (YYYY-MM-DD)"
// @Success     200       {object} clubResponse
// @Failure     400       {string} string "Invalid parameters"
// @Failure     404       {string} string "Club not found"
// @Failure     500       {string} string "Internal server error"
// @Router      /clubs/player/{playerId}/name/{clubName}/date/{date} [get]
func (h ClubsHandler) GetClubByPlayerNameAndDate(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.ParseInt(chi.URLParam(r, "playerId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}
	addedDate, err := time.Parse("2006-01-02", chi.URLParam(r, "date"))
	if err != nil {
		http.Error(w, "Invalid date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	club, err := h.Queries.GetClubByPlayerNameAndDate(r.Context(), db.GetClubByPlayerNameAndDateParams{
		PlayerID:  playerID,
		ClubName:  chi.URLParam(r, "clubName"),
		AddedDate: addedDate,
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Club not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get club: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toClubResponse(club)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreateClub godoc
// @Summary     Add a club to a player's bag
// @Tags        clubs
// @Accept      json
// @Produce     json
// @Param       club body     createClubRequest true "Club to create"
// @Success     201  {object} clubResponse
// @Failure     400  {string} string "Invalid request body"
// @Failure     500  {string} string "Internal server error"
// @Router      /clubs [post]
func (h ClubsHandler) CreateClub(w http.ResponseWriter, r *http.Request) {
	var req createClubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.PlayerID == 0 || req.ClubName == "" || req.AddedDate == "" {
		http.Error(w, "player_id, club_name and added_date are required", http.StatusBadRequest)
		return
	}

	addedDate, err := time.Parse("2006-01-02", req.AddedDate)
	if err != nil {
		http.Error(w, "Invalid added_date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "dispersion_bias", req.DispersionBias) {
		return
	}

	params := db.CreateClubParams{
		PlayerID:   req.PlayerID,
		ClubName:   req.ClubName,
		AddedDate:  addedDate,
		SampleSize: req.SampleSize,
	}
	if req.CarryAvg != nil {
		params.CarryAvg = sql.NullFloat64{Float64: *req.CarryAvg, Valid: true}
	}
	if req.CarryReliable != nil {
		params.CarryReliable = sql.NullFloat64{Float64: *req.CarryReliable, Valid: true}
	}
	if req.CarryMax != nil {
		params.CarryMax = sql.NullFloat64{Float64: *req.CarryMax, Valid: true}
	}
	if req.DispersionAvgM != nil {
		params.DispersionAvgM = sql.NullFloat64{Float64: *req.DispersionAvgM, Valid: true}
	}
	if req.DispersionBias != nil {
		params.DispersionBias = sql.NullString{String: *req.DispersionBias, Valid: true}
	}

	result, err := h.Queries.CreateClub(r.Context(), params)
	if err != nil {
		log.Printf("failed to create club: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	club, err := h.Queries.GetClubByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to fetch created club: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toClubResponse(club))
}

// RetireClub godoc
// @Summary     Retire a club (remove from bag)
// @Tags        clubs
// @Accept      json
// @Param       playerId  path int              true "Player ID"
// @Param       clubName  path string           true "Club name"
// @Param       body      body retireClubRequest true "Removal date"
// @Success     204
// @Failure     400 {string} string "Invalid request"
// @Failure     500 {string} string "Internal server error"
// @Router      /clubs/player/{playerId}/name/{clubName}/retire [patch]
func (h ClubsHandler) RetireClub(w http.ResponseWriter, r *http.Request) {
	playerID, err := strconv.ParseInt(chi.URLParam(r, "playerId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	var req retireClubRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	removedDate, err := time.Parse("2006-01-02", req.RemovedDate)
	if err != nil {
		http.Error(w, "Invalid removed_date format, expected YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	if err = h.Queries.RetireClub(r.Context(), db.RetireClubParams{
		PlayerID:    playerID,
		ClubName:    chi.URLParam(r, "clubName"),
		RemovedDate: sql.NullTime{Time: removedDate, Valid: true},
	}); err != nil {
		log.Printf("failed to retire club: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateClubDistances godoc
// @Summary     Update distance model for a club
// @Tags        clubs
// @Accept      json
// @Param       id   path int                       true "Club ID"
// @Param       body body updateClubDistancesRequest true "Distance model"
// @Success     204
// @Failure     400 {string} string "Invalid request"
// @Failure     500 {string} string "Internal server error"
// @Router      /clubs/{id}/distances [patch]
func (h ClubsHandler) UpdateClubDistances(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req updateClubDistancesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "dispersion_bias", req.DispersionBias) {
		return
	}

	params := db.UpdateClubDistancesParams{ID: id, SampleSize: req.SampleSize}
	if req.CarryAvg != nil {
		params.CarryAvg = sql.NullFloat64{Float64: *req.CarryAvg, Valid: true}
	}
	if req.CarryReliable != nil {
		params.CarryReliable = sql.NullFloat64{Float64: *req.CarryReliable, Valid: true}
	}
	if req.CarryMax != nil {
		params.CarryMax = sql.NullFloat64{Float64: *req.CarryMax, Valid: true}
	}
	if req.DispersionAvgM != nil {
		params.DispersionAvgM = sql.NullFloat64{Float64: *req.DispersionAvgM, Valid: true}
	}
	if req.DispersionBias != nil {
		params.DispersionBias = sql.NullString{String: *req.DispersionBias, Valid: true}
	}

	if err = h.Queries.UpdateClubDistances(r.Context(), params); err != nil {
		log.Printf("failed to update club distances: ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteClub godoc
// @Summary     Delete a club record
// @Tags        clubs
// @Param       id path int true "Club ID"
// @Success     204
// @Failure     400 {string} string "Invalid ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /clubs/{id} [delete]
func (h ClubsHandler) DeleteClub(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeleteClub(r.Context(), id); err != nil {
		log.Printf("failed to delete club: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
