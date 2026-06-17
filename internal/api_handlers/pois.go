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

type POIsHandler struct {
	Queries *db.Queries
}

type poiResponse struct {
	ID             int64    `json:"id"`
	CourseHoleID   int64    `json:"course_hole_id"`
	SpecificTee    *string  `json:"specific_tee"`
	PoiType        string   `json:"poi_type"`
	Side           *string  `json:"side"`
	ReferencePoint *string  `json:"reference_point"`
	DistanceStart  *float64 `json:"distance_start"`
	DistanceEnd    *float64 `json:"distance_end"`
	Label          string   `json:"label"`
}

type createPOIRequest struct {
	CourseHoleID   int64    `json:"course_hole_id"  example:"1"`
	SpecificTee    *string  `json:"specific_tee"    example:"white"`
	PoiType        string   `json:"poi_type"        example:"bunker"`
	Side           *string  `json:"side"            example:"left"`
	ReferencePoint *string  `json:"reference_point" example:"tee"`
	DistanceStart  *float64 `json:"distance_start"  example:"180.0"`
	DistanceEnd    *float64 `json:"distance_end"    example:"200.0"`
	Label          string   `json:"label"           example:"Fairway bunker left"`
}

type updatePOIRequest struct {
	SpecificTee    *string  `json:"specific_tee"`
	PoiType        string   `json:"poi_type"        example:"bunker"`
	Side           *string  `json:"side"            example:"left"`
	ReferencePoint *string  `json:"reference_point" example:"tee"`
	DistanceStart  *float64 `json:"distance_start"  example:"180.0"`
	DistanceEnd    *float64 `json:"distance_end"    example:"200.0"`
	Label          string   `json:"label"           example:"Fairway bunker left"`
}

func toPOIResponse(p db.HolePointsOfInterest) poiResponse {
	return poiResponse{
		ID:             p.ID,
		CourseHoleID:   p.CourseHoleID,
		SpecificTee:    helpers.NullableString(p.SpecificTee),
		PoiType:        p.PoiType,
		Side:           helpers.NullableString(p.Side),
		ReferencePoint: helpers.NullableString(p.ReferencePoint),
		DistanceStart:  helpers.NullableFloat64(p.DistanceStart),
		DistanceEnd:    helpers.NullableFloat64(p.DistanceEnd),
		Label:          p.Label,
	}
}

// ListPOIsByHole godoc
// @Summary     List POIs for a course hole
// @Tags        pois
// @Produce     json
// @Param       courseHoleId path     int true "Course Hole ID"
// @Success     200          {array}  poiResponse
// @Failure     400          {string} string "Invalid course hole ID"
// @Failure     500          {string} string "Internal server error"
// @Router      /pois/hole/{courseHoleId} [get]
func (h POIsHandler) ListPOIsByHole(w http.ResponseWriter, r *http.Request) {
	courseHoleID, err := strconv.ParseInt(chi.URLParam(r, "courseHoleId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course hole ID", http.StatusBadRequest)
		return
	}

	pois, err := h.Queries.ListPOIsByHole(r.Context(), courseHoleID)
	if err != nil {
		log.Printf("failed to list POIs: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]poiResponse, len(pois))
	for i, p := range pois {
		resp[i] = toPOIResponse(p)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// ListPOIsByHoleAndTee godoc
// @Summary     List POIs for a course hole filtered by tee
// @Tags        pois
// @Produce     json
// @Param       courseHoleId path     int    true "Course Hole ID"
// @Param       tee          path     string true "Tee name"
// @Success     200          {array}  poiResponse
// @Failure     400          {string} string "Invalid course hole ID"
// @Failure     500          {string} string "Internal server error"
// @Router      /pois/hole/{courseHoleId}/tee/{tee} [get]
func (h POIsHandler) ListPOIsByHoleAndTee(w http.ResponseWriter, r *http.Request) {
	courseHoleID, err := strconv.ParseInt(chi.URLParam(r, "courseHoleId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course hole ID", http.StatusBadRequest)
		return
	}

	pois, err := h.Queries.ListPOIsByHoleAndTee(r.Context(), db.ListPOIsByHoleAndTeeParams{
		CourseHoleID: courseHoleID,
		SpecificTee:  sql.NullString{String: chi.URLParam(r, "tee"), Valid: true},
	})
	if err != nil {
		log.Printf("failed to list POIs by tee: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]poiResponse, len(pois))
	for i, p := range pois {
		resp[i] = toPOIResponse(p)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetPOIByID godoc
// @Summary     Get a POI by ID
// @Tags        pois
// @Produce     json
// @Param       id  path     int true "POI ID"
// @Success     200 {object} poiResponse
// @Failure     400 {string} string "Invalid ID"
// @Failure     404 {string} string "POI not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /pois/{id} [get]
func (h POIsHandler) GetPOIByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	poi, err := h.Queries.GetPOIByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "POI not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get POI: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toPOIResponse(poi)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreatePOI godoc
// @Summary     Create a point of interest
// @Tags        pois
// @Accept      json
// @Produce     json
// @Param       poi body     createPOIRequest true "POI to create"
// @Success     201 {object} poiResponse
// @Failure     400 {string} string "Invalid request body"
// @Failure     500 {string} string "Internal server error"
// @Router      /pois [post]
func (h POIsHandler) CreatePOI(w http.ResponseWriter, r *http.Request) {
	var req createPOIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.CourseHoleID == 0 || req.PoiType == "" || req.Label == "" {
		http.Error(w, "course_hole_id, poi_type and label are required", http.StatusBadRequest)
		return
	}

	if !helpers.CheckVocab(w, r.Context(), h.Queries, "poi_type", req.PoiType) {
		return
	}
	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "poi_side", req.Side) {
		return
	}
	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "reference_point", req.ReferencePoint) {
		return
	}

	params := db.CreatePOIParams{
		CourseHoleID: req.CourseHoleID,
		PoiType:      req.PoiType,
		Label:        req.Label,
	}
	if req.SpecificTee != nil {
		params.SpecificTee = sql.NullString{String: *req.SpecificTee, Valid: true}
	}
	if req.Side != nil {
		params.Side = sql.NullString{String: *req.Side, Valid: true}
	}
	if req.ReferencePoint != nil {
		params.ReferencePoint = sql.NullString{String: *req.ReferencePoint, Valid: true}
	}
	if req.DistanceStart != nil {
		params.DistanceStart = sql.NullFloat64{Float64: *req.DistanceStart, Valid: true}
	}
	if req.DistanceEnd != nil {
		params.DistanceEnd = sql.NullFloat64{Float64: *req.DistanceEnd, Valid: true}
	}

	result, err := h.Queries.CreatePOI(r.Context(), params)
	if err != nil {
		log.Printf("failed to create POI: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	poi, err := h.Queries.GetPOIByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to fetch created POI: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toPOIResponse(poi))
}

// UpdatePOI godoc
// @Summary     Update a point of interest
// @Tags        pois
// @Accept      json
// @Param       id  path int           true "POI ID"
// @Param       poi body updatePOIRequest true "Updated fields"
// @Success     204
// @Failure     400 {string} string "Invalid request"
// @Failure     500 {string} string "Internal server error"
// @Router      /pois/{id} [put]
func (h POIsHandler) UpdatePOI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req updatePOIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if !helpers.CheckVocab(w, r.Context(), h.Queries, "poi_type", req.PoiType) {
		return
	}
	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "poi_side", req.Side) {
		return
	}
	if !helpers.CheckOptVocab(w, r.Context(), h.Queries, "reference_point", req.ReferencePoint) {
		return
	}

	params := db.UpdatePOIParams{ID: id, PoiType: req.PoiType, Label: req.Label}
	if req.SpecificTee != nil {
		params.SpecificTee = sql.NullString{String: *req.SpecificTee, Valid: true}
	}
	if req.Side != nil {
		params.Side = sql.NullString{String: *req.Side, Valid: true}
	}
	if req.ReferencePoint != nil {
		params.ReferencePoint = sql.NullString{String: *req.ReferencePoint, Valid: true}
	}
	if req.DistanceStart != nil {
		params.DistanceStart = sql.NullFloat64{Float64: *req.DistanceStart, Valid: true}
	}
	if req.DistanceEnd != nil {
		params.DistanceEnd = sql.NullFloat64{Float64: *req.DistanceEnd, Valid: true}
	}

	if err = h.Queries.UpdatePOI(r.Context(), params); err != nil {
		log.Printf("failed to update POI: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePOI godoc
// @Summary     Delete a POI
// @Tags        pois
// @Param       id path int true "POI ID"
// @Success     204
// @Failure     400 {string} string "Invalid ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /pois/{id} [delete]
func (h POIsHandler) DeletePOI(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeletePOI(r.Context(), id); err != nil {
		log.Printf("failed to delete POI: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePOIsByHole godoc
// @Summary     Delete all POIs for a course hole
// @Tags        pois
// @Param       courseHoleId path int true "Course Hole ID"
// @Success     204
// @Failure     400 {string} string "Invalid course hole ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /pois/hole/{courseHoleId} [delete]
func (h POIsHandler) DeletePOIsByHole(w http.ResponseWriter, r *http.Request) {
	courseHoleID, err := strconv.ParseInt(chi.URLParam(r, "courseHoleId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course hole ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeletePOIsByHole(r.Context(), courseHoleID); err != nil {
		log.Printf("failed to delete POIs by hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
