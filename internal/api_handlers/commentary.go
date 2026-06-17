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
)

type CommentaryHandler struct {
	Queries *db.Queries
}

type commentaryResponse struct {
	ID          int64     `json:"id"`
	Scope       string    `json:"scope"`
	ScopeID     int64     `json:"scope_id"`
	Content     string    `json:"content"`
	GeneratedAt time.Time `json:"generated_at"`
}

type createCommentaryRequest struct {
	Scope   string `json:"scope"    example:"hole"`
	ScopeID int64  `json:"scope_id" example:"1"`
	Content string `json:"content"  example:"A strong par on a tricky hole."`
}

func toCommentaryResponse(c db.Commentary) commentaryResponse {
	return commentaryResponse{
		ID:          c.ID,
		Scope:       c.Scope,
		ScopeID:     c.ScopeID,
		Content:     c.Content,
		GeneratedAt: c.GeneratedAt,
	}
}

// ListCommentaryByScope godoc
// @Summary     List commentary entries for a hole or round
// @Tags        commentary
// @Produce     json
// @Param       scope   path     string true "Scope (hole or round)"
// @Param       scopeId path     int    true "Scope ID"
// @Success     200     {array}  commentaryResponse
// @Failure     400     {string} string "Invalid scope ID"
// @Failure     500     {string} string "Internal server error"
// @Router      /commentary/scope/{scope}/{scopeId} [get]
func (h CommentaryHandler) ListCommentaryByScope(w http.ResponseWriter, r *http.Request) {
	scopeID, err := strconv.ParseInt(chi.URLParam(r, "scopeId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid scope ID", http.StatusBadRequest)
		return
	}

	entries, err := h.Queries.ListCommentaryByScope(r.Context(), db.ListCommentaryByScopeParams{
		Scope:   chi.URLParam(r, "scope"),
		ScopeID: scopeID,
	})
	if err != nil {
		log.Printf("failed to list commentary: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]commentaryResponse, len(entries))
	for i, c := range entries {
		resp[i] = toCommentaryResponse(c)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetLatestCommentaryByScope godoc
// @Summary     Get the most recent commentary for a hole or round
// @Tags        commentary
// @Produce     json
// @Param       scope   path     string true "Scope (hole or round)"
// @Param       scopeId path     int    true "Scope ID"
// @Success     200     {object} commentaryResponse
// @Failure     400     {string} string "Invalid scope ID"
// @Failure     404     {string} string "No commentary found"
// @Failure     500     {string} string "Internal server error"
// @Router      /commentary/scope/{scope}/{scopeId}/latest [get]
func (h CommentaryHandler) GetLatestCommentaryByScope(w http.ResponseWriter, r *http.Request) {
	scopeID, err := strconv.ParseInt(chi.URLParam(r, "scopeId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid scope ID", http.StatusBadRequest)
		return
	}

	entry, err := h.Queries.GetLatestCommentaryByScope(r.Context(), db.GetLatestCommentaryByScopeParams{
		Scope:   chi.URLParam(r, "scope"),
		ScopeID: scopeID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "No commentary found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get latest commentary: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toCommentaryResponse(entry)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetCommentaryByID godoc
// @Summary     Get a commentary entry by ID
// @Tags        commentary
// @Produce     json
// @Param       id  path     int true "Commentary ID"
// @Success     200 {object} commentaryResponse
// @Failure     400 {string} string "Invalid ID"
// @Failure     404 {string} string "Commentary not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /commentary/{id} [get]
func (h CommentaryHandler) GetCommentaryByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	entry, err := h.Queries.GetCommentaryByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Commentary not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get commentary: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toCommentaryResponse(entry)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreateCommentary godoc
// @Summary     Store a commentary entry
// @Tags        commentary
// @Accept      json
// @Produce     json
// @Param       commentary body     createCommentaryRequest true "Commentary to store"
// @Success     201        {object} commentaryResponse
// @Failure     400        {string} string "Invalid request body"
// @Failure     500        {string} string "Internal server error"
// @Router      /commentary [post]
func (h CommentaryHandler) CreateCommentary(w http.ResponseWriter, r *http.Request) {
	var req createCommentaryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Scope == "" || req.ScopeID == 0 || req.Content == "" {
		http.Error(w, "scope, scope_id and content are required", http.StatusBadRequest)
		return
	}

	result, err := h.Queries.CreateCommentary(r.Context(), db.CreateCommentaryParams{
		Scope:   req.Scope,
		ScopeID: req.ScopeID,
		Content: req.Content,
	})
	if err != nil {
		log.Printf("failed to create commentary: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	entry, err := h.Queries.GetCommentaryByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to fetch created commentary: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toCommentaryResponse(entry))
}

// DeleteCommentaryByScope godoc
// @Summary     Delete all commentary for a hole or round
// @Tags        commentary
// @Param       scope   path string true "Scope (hole or round)"
// @Param       scopeId path int    true "Scope ID"
// @Success     204
// @Failure     400 {string} string "Invalid scope ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /commentary/scope/{scope}/{scopeId} [delete]
func (h CommentaryHandler) DeleteCommentaryByScope(w http.ResponseWriter, r *http.Request) {
	scopeID, err := strconv.ParseInt(chi.URLParam(r, "scopeId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid scope ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeleteCommentaryByScope(r.Context(), db.DeleteCommentaryByScopeParams{
		Scope:   chi.URLParam(r, "scope"),
		ScopeID: scopeID,
	}); err != nil {
		log.Printf("failed to delete commentary by scope: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteCommentary godoc
// @Summary     Delete a commentary entry by ID
// @Tags        commentary
// @Param       id path int true "Commentary ID"
// @Success     204
// @Failure     400 {string} string "Invalid ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /commentary/{id} [delete]
func (h CommentaryHandler) DeleteCommentary(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeleteCommentary(r.Context(), id); err != nil {
		log.Printf("failed to delete commentary: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
