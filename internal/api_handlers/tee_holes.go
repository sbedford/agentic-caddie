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

type TeeHolesHandler struct {
	Queries *db.Queries
}

type teeHoleResponse struct {
	ID           int64    `json:"id"`
	CourseHoleID int64    `json:"course_hole_id"`
	TeeID        int64    `json:"tee_id"`
	Par          int64    `json:"par"`
	StrokeIndex  *int64   `json:"stroke_index"`
	Distance     int64    `json:"distance"`
	TeeCentreLat *float64 `json:"tee_centre_lat"`
	TeeCentreLng *float64 `json:"tee_centre_lng"`
}

type createTeeHoleRequest struct {
	CourseHoleID int64    `json:"course_hole_id" example:"1"`
	TeeID        int64    `json:"tee_id"         example:"1"`
	Par          int64    `json:"par"            example:"4"`
	StrokeIndex  *int64   `json:"stroke_index"   example:"7"`
	Distance     int64    `json:"distance"       example:"385"`
	TeeCentreLat *float64 `json:"tee_centre_lat" example:"-33.8688"`
	TeeCentreLng *float64 `json:"tee_centre_lng" example:"151.2093"`
}

type updateTeeHoleRequest struct {
	Par          int64    `json:"par"            example:"4"`
	StrokeIndex  *int64   `json:"stroke_index"   example:"7"`
	Distance     int64    `json:"distance"       example:"385"`
	TeeCentreLat *float64 `json:"tee_centre_lat" example:"-33.8688"`
	TeeCentreLng *float64 `json:"tee_centre_lng" example:"151.2093"`
}

func toTeeHoleResponse(th db.TeeHole) teeHoleResponse {
	return teeHoleResponse{
		ID:           th.ID,
		CourseHoleID: th.CourseHoleID,
		TeeID:        th.TeeID,
		Par:          th.Par,
		StrokeIndex:  nullableInt64(th.StrokeIndex),
		Distance:     th.Distance,
		TeeCentreLat: nullableFloat64(th.TeeCentreLat),
		TeeCentreLng: nullableFloat64(th.TeeCentreLng),
	}
}

// ListTeeHoles godoc
// @Summary     List tee holes for a tee
// @Tags        tee-holes
// @Produce     json
// @Param       teeId path     int true "Tee ID"
// @Success     200   {array}  teeHoleResponse
// @Failure     400   {string} string "Invalid tee ID"
// @Failure     500   {string} string "Internal server error"
// @Router      /tee-holes/tee/{teeId} [get]
func (h TeeHolesHandler) ListTeeHoles(w http.ResponseWriter, r *http.Request) {
	teeID, err := strconv.ParseInt(chi.URLParam(r, "teeId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid tee ID", http.StatusBadRequest)
		return
	}

	holes, err := h.Queries.ListTeeHoles(r.Context(), teeID)
	if err != nil {
		log.Printf("failed to list tee holes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]teeHoleResponse, len(holes))
	for i, th := range holes {
		resp[i] = toTeeHoleResponse(th)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetTeeHoleByID godoc
// @Summary     Get a tee hole by ID
// @Tags        tee-holes
// @Produce     json
// @Param       id  path     int true "Tee Hole ID"
// @Success     200 {object} teeHoleResponse
// @Failure     400 {string} string "Invalid ID"
// @Failure     404 {string} string "Tee hole not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /tee-holes/{id} [get]
func (h TeeHolesHandler) GetTeeHoleByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	hole, err := h.Queries.GetTeeHoleByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Tee hole not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get tee hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toTeeHoleResponse(hole)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetTeeHoleByHoleAndTee godoc
// @Summary     Get a tee hole by course hole and tee
// @Tags        tee-holes
// @Produce     json
// @Param       courseHoleId path     int true "Course Hole ID"
// @Param       teeId        path     int true "Tee ID"
// @Success     200          {object} teeHoleResponse
// @Failure     400          {string} string "Invalid parameters"
// @Failure     404          {string} string "Tee hole not found"
// @Failure     500          {string} string "Internal server error"
// @Router      /tee-holes/hole/{courseHoleId}/tee/{teeId} [get]
func (h TeeHolesHandler) GetTeeHoleByHoleAndTee(w http.ResponseWriter, r *http.Request) {
	courseHoleID, err := strconv.ParseInt(chi.URLParam(r, "courseHoleId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course hole ID", http.StatusBadRequest)
		return
	}
	teeID, err := strconv.ParseInt(chi.URLParam(r, "teeId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid tee ID", http.StatusBadRequest)
		return
	}

	hole, err := h.Queries.GetTeeHoleByHoleAndTee(r.Context(), db.GetTeeHoleByHoleAndTeeParams{
		CourseHoleID: courseHoleID,
		TeeID:        teeID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Tee hole not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get tee hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toTeeHoleResponse(hole)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreateTeeHole godoc
// @Summary     Create a tee hole
// @Tags        tee-holes
// @Accept      json
// @Produce     json
// @Param       hole body     createTeeHoleRequest true "Tee hole to create"
// @Success     201  {object} teeHoleResponse
// @Failure     400  {string} string "Invalid request body"
// @Failure     500  {string} string "Internal server error"
// @Router      /tee-holes [post]
func (h TeeHolesHandler) CreateTeeHole(w http.ResponseWriter, r *http.Request) {
	var req createTeeHoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.CourseHoleID == 0 || req.TeeID == 0 || req.Par == 0 {
		http.Error(w, "course_hole_id, tee_id and par are required", http.StatusBadRequest)
		return
	}

	params := db.CreateTeeHoleParams{
		CourseHoleID: req.CourseHoleID,
		TeeID:        req.TeeID,
		Par:          req.Par,
		Distance:     req.Distance,
	}
	if req.StrokeIndex != nil {
		params.StrokeIndex = sql.NullInt64{Int64: *req.StrokeIndex, Valid: true}
	}
	if req.TeeCentreLat != nil {
		params.TeeCentreLat = sql.NullFloat64{Float64: *req.TeeCentreLat, Valid: true}
	}
	if req.TeeCentreLng != nil {
		params.TeeCentreLng = sql.NullFloat64{Float64: *req.TeeCentreLng, Valid: true}
	}

	result, err := h.Queries.CreateTeeHole(r.Context(), params)
	if err != nil {
		log.Printf("failed to create tee hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	hole, err := h.Queries.GetTeeHoleByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to fetch created tee hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toTeeHoleResponse(hole))
}

// UpdateTeeHole godoc
// @Summary     Update a tee hole
// @Tags        tee-holes
// @Accept      json
// @Param       id   path int               true "Tee Hole ID"
// @Param       hole body updateTeeHoleRequest true "Updated fields"
// @Success     204
// @Failure     400 {string} string "Invalid request"
// @Failure     500 {string} string "Internal server error"
// @Router      /tee-holes/{id} [put]
func (h TeeHolesHandler) UpdateTeeHole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req updateTeeHoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := db.UpdateTeeHoleParams{ID: id, Par: req.Par, Distance: req.Distance}
	if req.StrokeIndex != nil {
		params.StrokeIndex = sql.NullInt64{Int64: *req.StrokeIndex, Valid: true}
	}
	if req.TeeCentreLat != nil {
		params.TeeCentreLat = sql.NullFloat64{Float64: *req.TeeCentreLat, Valid: true}
	}
	if req.TeeCentreLng != nil {
		params.TeeCentreLng = sql.NullFloat64{Float64: *req.TeeCentreLng, Valid: true}
	}

	if err = h.Queries.UpdateTeeHole(r.Context(), params); err != nil {
		log.Printf("failed to update tee hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteTeeHole godoc
// @Summary     Delete a tee hole
// @Tags        tee-holes
// @Param       id path int true "Tee Hole ID"
// @Success     204
// @Failure     400 {string} string "Invalid ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /tee-holes/{id} [delete]
func (h TeeHolesHandler) DeleteTeeHole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeleteTeeHole(r.Context(), id); err != nil {
		log.Printf("failed to delete tee hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
