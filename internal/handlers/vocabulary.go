package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sbedford/agentic-caddie/internal/db"
)

type VocabularyHandler struct {
	Queries *db.Queries
}

type vocabularyEntryResponse struct {
	Domain    string `json:"domain"`
	Value     string `json:"value"`
	Label     string `json:"label"`
	SortOrder int64  `json:"sort_order"`
}

type createVocabularyEntryRequest struct {
	Domain    string `json:"domain"     example:"shot_type"`
	Value     string `json:"value"      example:"punch"`
	Label     string `json:"label"      example:"Punch"`
	SortOrder int64  `json:"sort_order" example:"9"`
}

type updateVocabularyEntryRequest struct {
	Label     string `json:"label"      example:"Punch Out"`
	SortOrder int64  `json:"sort_order" example:"9"`
}

func toVocabResponse(v db.Vocabulary) vocabularyEntryResponse {
	return vocabularyEntryResponse{
		Domain:    v.Domain,
		Value:     v.Value,
		Label:     v.Label,
		SortOrder: v.SortOrder,
	}
}

// GetAllVocabulary godoc
// @Summary     List all vocabulary entries grouped by domain
// @Tags        vocabulary
// @Produce     json
// @Success     200 {array}  vocabularyEntryResponse
// @Failure     500 {string} string "Internal server error"
// @Router      /vocabulary [get]
func (h VocabularyHandler) GetAllVocabulary(w http.ResponseWriter, r *http.Request) {

	entries, err := h.Queries.GetAllVocabulary(r.Context())
	if err != nil {
		log.Printf("failed to list vocabulary: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]vocabularyEntryResponse, len(entries))
	for i, e := range entries {
		resp[i] = toVocabResponse(e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// GetVocabularyByDomain godoc
// @Summary     List vocabulary entries for a domain
// @Tags        vocabulary
// @Produce     json
// @Param       domain path     string true "Domain name"
// @Success     200    {array}  vocabularyEntryResponse
// @Failure     500    {string} string "Internal server error"
// @Router      /vocabulary/{domain} [get]
func (h VocabularyHandler) GetVocabularyByDomain(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")

	entries, err := h.Queries.GetVocabularyByDomain(r.Context(), domain)
	if err != nil {
		log.Printf("failed to list vocabulary for domain %q: %v", domain, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]vocabularyEntryResponse, len(entries))
	for i, e := range entries {
		resp[i] = toVocabResponse(e)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// CreateVocabularyEntry godoc
// @Summary     Create a new vocabulary entry
// @Tags        vocabulary
// @Accept      json
// @Produce     json
// @Param       entry body     createVocabularyEntryRequest true "Entry to create"
// @Success     201   {object} vocabularyEntryResponse
// @Failure     400   {string} string "Invalid request body"
// @Failure     409   {string} string "Entry already exists"
// @Failure     500   {string} string "Internal server error"
// @Router      /vocabulary [post]
func (h VocabularyHandler) CreateVocabularyEntry(w http.ResponseWriter, r *http.Request) {
	var req createVocabularyEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Domain == "" || req.Value == "" || req.Label == "" {
		http.Error(w, "domain, value and label are required", http.StatusBadRequest)
		return
	}

	err := h.Queries.CreateVocabularyEntry(r.Context(), db.CreateVocabularyEntryParams{
		Domain:    req.Domain,
		Value:     req.Value,
		Label:     req.Label,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		if isUniqueConstraintError(err) {
			http.Error(w, "entry already exists for that domain+value", http.StatusConflict)
			return
		}
		log.Printf("failed to create vocabulary entry: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vocabularyEntryResponse{
		Domain:    req.Domain,
		Value:     req.Value,
		Label:     req.Label,
		SortOrder: req.SortOrder,
	})
}

// UpdateVocabularyEntry godoc
// @Summary     Update label and sort order for a vocabulary entry
// @Tags        vocabulary
// @Accept      json
// @Param       domain path     string                       true "Domain name"
// @Param       value  path     string                       true "Value"
// @Param       entry  body     updateVocabularyEntryRequest true "Fields to update"
// @Success     204
// @Failure     400 {string} string "Invalid request body"
// @Failure     404 {string} string "Entry not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /vocabulary/{domain}/{value} [put]
func (h VocabularyHandler) UpdateVocabularyEntry(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	value := chi.URLParam(r, "value")

	var req updateVocabularyEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Label == "" {
		http.Error(w, "label is required", http.StatusBadRequest)
		return
	}

	if err := h.Queries.UpdateVocabularyEntry(r.Context(), db.UpdateVocabularyEntryParams{
		Label:     req.Label,
		SortOrder: req.SortOrder,
		Domain:    domain,
		Value:     value,
	}); err != nil {
		log.Printf("failed to update vocabulary entry %s/%s: %v", domain, value, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DeleteVocabularyEntry godoc
// @Summary     Delete a vocabulary entry
// @Tags        vocabulary
// @Param       domain path string true "Domain name"
// @Param       value  path string true "Value"
// @Success     204
// @Failure     500 {string} string "Internal server error"
// @Router      /vocabulary/{domain}/{value} [delete]
func (h VocabularyHandler) DeleteVocabularyEntry(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	value := chi.URLParam(r, "value")

	if err := h.Queries.DeleteVocabularyEntry(r.Context(), db.DeleteVocabularyEntryParams{
		Domain: domain,
		Value:  value,
	}); err != nil {
		log.Printf("failed to delete vocabulary entry %s/%s: %v", domain, value, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// isUniqueConstraintError detects SQLite UNIQUE violation by error message.
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return len(msg) >= 6 && msg[:6] == "UNIQUE"
}
