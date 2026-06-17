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

type PlayersHandler struct {
	Queries *db.Queries
}

type playerResponse struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	Handicap float64 `json:"handicap"`
}

type createPlayerRequest struct {
	Name     string  `json:"name"     example:"Alice"`
	Handicap float64 `json:"handicap" example:"12.5"`
}

type updateHandicapRequest struct {
	Handicap float64 `json:"handicap" example:"10.2"`
}

// ListPlayers godoc
// @Summary     List all players
// @Tags        players
// @Produce     json
// @Success     200 {array}  playerResponse
// @Failure     500 {string} string "Internal server error"
// @Router      /players [get]
func (b PlayersHandler) ListPlayers(w http.ResponseWriter, r *http.Request) {

	players, err := b.Queries.ListPlayers(r.Context())
	if err != nil {
		log.Printf("failed to list players: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	responseData := make([]playerResponse, len(players))
	for i, p := range players {
		responseData[i] = playerResponse{
			ID:       p.ID,
			Name:     p.Name,
			Handicap: p.Handicap.Float64,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(responseData); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetPlayer godoc
// @Summary     Get a player by ID
// @Tags        players
// @Produce     json
// @Param       id  path     int  true  "Player ID"
// @Success     200 {object} playerResponse
// @Failure     400 {string} string "Invalid player ID"
// @Failure     404 {string} string "Player not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /players/{id} [get]
func (b PlayersHandler) GetPlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	player, err := b.Queries.GetPlayerByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get player by id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(playerResponse{ID: player.ID, Name: player.Name, Handicap: player.Handicap.Float64}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetPlayerByName godoc
// @Summary     Get a player by name
// @Tags        players
// @Produce     json
// @Param       name path     string true "Player name"
// @Success     200  {object} playerResponse
// @Failure     404  {string} string "Player not found"
// @Failure     500  {string} string "Internal server error"
// @Router      /players/name/{name} [get]
func (b PlayersHandler) GetPlayerByName(w http.ResponseWriter, r *http.Request) {

	player, err := b.Queries.GetPlayerByName(r.Context(), chi.URLParam(r, "name"))
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get player by name: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(playerResponse{ID: player.ID, Name: player.Name, Handicap: player.Handicap.Float64}); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreatePlayer godoc
// @Summary     Create a new player
// @Tags        players
// @Accept      json
// @Produce     json
// @Param       player body     createPlayerRequest true "Player to create"
// @Success     201    {object} playerResponse
// @Failure     400    {string} string "Invalid request body"
// @Failure     500    {string} string "Internal server error"
// @Router      /players [post]
func (b PlayersHandler) CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var req createPlayerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	result, err := b.Queries.CreatePlayer(r.Context(), db.CreatePlayerParams{
		Name:     req.Name,
		Handicap: sql.NullFloat64{Float64: req.Handicap, Valid: true},
	})
	if err != nil {
		log.Printf("failed to create player: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	player, err := b.Queries.GetPlayerByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to get created player: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(playerResponse{
		ID:       player.ID,
		Name:     player.Name,
		Handicap: player.Handicap.Float64,
	})
}

// UpdateHandicap godoc
// @Summary     Update a player's handicap
// @Tags        players
// @Accept      json
// @Param       id       path int                  true "Player ID"
// @Param       handicap body updateHandicapRequest true "New handicap value"
// @Success     204
// @Failure     400 {string} string "Invalid player ID or request body"
// @Failure     500 {string} string "Internal server error"
// @Router      /players/{id}/handicap [patch]
func (b PlayersHandler) UpdateHandicap(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	var req updateHandicapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = b.Queries.UpdatePlayerHandicap(r.Context(), db.UpdatePlayerHandicapParams{
		ID:       id,
		Handicap: sql.NullFloat64{Float64: req.Handicap, Valid: true},
	})
	if err != nil {
		log.Printf("failed to update player handicap: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePlayer godoc
// @Summary     Delete a player
// @Tags        players
// @Param       id  path int true "Player ID"
// @Success     204
// @Failure     400 {string} string "Invalid player ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /players/{id} [delete]
func (b PlayersHandler) DeletePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid player ID", http.StatusBadRequest)
		return
	}

	if err = b.Queries.DeletePlayer(r.Context(), id); err != nil {
		log.Printf("failed to delete player by id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
