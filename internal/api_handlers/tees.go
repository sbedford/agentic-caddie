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

type TeesHandler struct {
	Queries *db.Queries
}

type teeResponse struct {
	ID           int64    `json:"id"`
	CourseID     int64    `json:"course_id"`
	Name         string   `json:"name"`
	SlopeRating  *int64   `json:"slope_rating"`
	CourseRating *float64 `json:"course_rating"`
}

type createTeeRequest struct {
	CourseID     int64    `json:"course_id"    example:"1"`
	Name         string   `json:"name"         example:"white"`
	SlopeRating  *int64   `json:"slope_rating" example:"113"`
	CourseRating *float64 `json:"course_rating" example:"69.5"`
}

type updateTeeRequest struct {
	SlopeRating  *int64   `json:"slope_rating"  example:"113"`
	CourseRating *float64 `json:"course_rating" example:"69.5"`
}

func toTeeResponse(t db.Tee) teeResponse {
	return teeResponse{
		ID:           t.ID,
		CourseID:     t.CourseID,
		Name:         t.Name,
		SlopeRating:  helpers.NullableInt64(t.SlopeRating),
		CourseRating: helpers.NullableFloat64(t.CourseRating),
	}
}

// ListTees godoc
// @Summary     List all tees
// @Tags        tees
// @Produce     json
// @Success     200 {array}  teeResponse
// @Failure     500 {string} string "Internal server error"
// @Router      /tees [get]
func (h TeesHandler) ListTees(w http.ResponseWriter, r *http.Request) {

	tees, err := h.Queries.ListTees(r.Context())
	if err != nil {
		log.Printf("failed to list tees: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]teeResponse, len(tees))
	for i, t := range tees {
		resp[i] = toTeeResponse(t)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetTeeByID godoc
// @Summary     Get a tee by ID
// @Tags        tees
// @Produce     json
// @Param       id  path     int true "Tee ID"
// @Success     200 {object} teeResponse
// @Failure     400 {string} string "Invalid tee ID"
// @Failure     404 {string} string "Tee not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /tees/{id} [get]
func (h TeesHandler) GetTeeByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid tee ID", http.StatusBadRequest)
		return
	}

	tee, err := h.Queries.GetTeeByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Tee not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get tee: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toTeeResponse(tee)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetTeesByCourse godoc
// @Summary     List tees for a course
// @Tags        tees
// @Produce     json
// @Param       courseId path     int true "Course ID"
// @Success     200      {array}  teeResponse
// @Failure     400      {string} string "Invalid course ID"
// @Failure     500      {string} string "Internal server error"
// @Router      /tees/course/{courseId} [get]
func (h TeesHandler) GetTeesByCourse(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	tees, err := h.Queries.GetTeesByCourse(r.Context(), courseID)
	if err != nil {
		log.Printf("failed to get tees by course: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]teeResponse, len(tees))
	for i, t := range tees {
		resp[i] = toTeeResponse(t)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetTeeByCourseAndName godoc
// @Summary     Get a tee by course and name
// @Tags        tees
// @Produce     json
// @Param       courseId path     int    true "Course ID"
// @Param       name     path     string true "Tee name"
// @Success     200      {object} teeResponse
// @Failure     400      {string} string "Invalid course ID"
// @Failure     404      {string} string "Tee not found"
// @Failure     500      {string} string "Internal server error"
// @Router      /tees/course/{courseId}/{name} [get]
func (h TeesHandler) GetTeeByCourseAndName(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	tee, err := h.Queries.GetTeeByCourseAndName(r.Context(), db.GetTeeByCourseAndNameParams{
		CourseID: courseID,
		Name:     chi.URLParam(r, "name"),
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Tee not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get tee by course and name: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toTeeResponse(tee)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreateTee godoc
// @Summary     Create a tee
// @Tags        tees
// @Accept      json
// @Produce     json
// @Param       tee body     createTeeRequest true "Tee to create"
// @Success     201 {object} teeResponse
// @Failure     400 {string} string "Invalid request body"
// @Failure     500 {string} string "Internal server error"
// @Router      /tees [post]
func (h TeesHandler) CreateTee(w http.ResponseWriter, r *http.Request) {
	var req createTeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.CourseID == 0 {
		http.Error(w, "course_id and name are required", http.StatusBadRequest)
		return
	}

	params := db.CreateTeeParams{CourseID: req.CourseID, Name: req.Name}
	if req.SlopeRating != nil {
		params.SlopeRating = sql.NullInt64{Int64: *req.SlopeRating, Valid: true}
	}
	if req.CourseRating != nil {
		params.CourseRating = sql.NullFloat64{Float64: *req.CourseRating, Valid: true}
	}

	result, err := h.Queries.CreateTee(r.Context(), params)
	if err != nil {
		log.Printf("failed to create tee: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	tee, err := h.Queries.GetTeeByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to fetch created tee: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toTeeResponse(tee))
}

// UpdateTee godoc
// @Summary     Update a tee's ratings
// @Tags        tees
// @Accept      json
// @Param       id  path int           true "Tee ID"
// @Param       tee body updateTeeRequest true "Updated ratings"
// @Success     204
// @Failure     400 {string} string "Invalid request"
// @Failure     500 {string} string "Internal server error"
// @Router      /tees/{id} [put]
func (h TeesHandler) UpdateTee(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid tee ID", http.StatusBadRequest)
		return
	}

	var req updateTeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := db.UpdateTeeParams{ID: id}
	if req.SlopeRating != nil {
		params.SlopeRating = sql.NullInt64{Int64: *req.SlopeRating, Valid: true}
	}
	if req.CourseRating != nil {
		params.CourseRating = sql.NullFloat64{Float64: *req.CourseRating, Valid: true}
	}

	if err = h.Queries.UpdateTee(r.Context(), params); err != nil {
		log.Printf("failed to update tee: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteTee godoc
// @Summary     Delete a tee
// @Tags        tees
// @Param       id path int true "Tee ID"
// @Success     204
// @Failure     400 {string} string "Invalid tee ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /tees/{id} [delete]
func (h TeesHandler) DeleteTee(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid tee ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeleteTee(r.Context(), id); err != nil {
		log.Printf("failed to delete tee: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
