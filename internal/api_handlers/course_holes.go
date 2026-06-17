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

type CourseHolesHandler struct {
	Queries *db.Queries
}

type courseHoleResponse struct {
	ID             int64    `json:"id"`
	CourseID       int64    `json:"course_id"`
	HoleNumber     int64    `json:"hole_number"`
	GreenCentreLat *float64 `json:"green_centre_lat"`
	GreenCentreLng *float64 `json:"green_centre_lng"`
}

type createCourseHoleRequest struct {
	CourseID       int64    `json:"course_id"        example:"1"`
	HoleNumber     int64    `json:"hole_number"      example:"1"`
	GreenCentreLat *float64 `json:"green_centre_lat" example:"-33.8688"`
	GreenCentreLng *float64 `json:"green_centre_lng" example:"151.2093"`
}

type updateCourseHoleCoordinatesRequest struct {
	GreenCentreLat *float64 `json:"green_centre_lat" example:"-33.8688"`
	GreenCentreLng *float64 `json:"green_centre_lng" example:"151.2093"`
}

func toCourseHoleResponse(ch db.CourseHole) courseHoleResponse {
	return courseHoleResponse{
		ID:             ch.ID,
		CourseID:       ch.CourseID,
		HoleNumber:     ch.HoleNumber,
		GreenCentreLat: helpers.NullableFloat64(ch.GreenCentreLat),
		GreenCentreLng: helpers.NullableFloat64(ch.GreenCentreLng),
	}
}

// ListCourseHoles godoc
// @Summary     List holes for a course
// @Tags        course-holes
// @Produce     json
// @Param       courseId path     int true "Course ID"
// @Success     200      {array}  courseHoleResponse
// @Failure     400      {string} string "Invalid course ID"
// @Failure     500      {string} string "Internal server error"
// @Router      /course-holes/course/{courseId} [get]
func (h CourseHolesHandler) ListCourseHoles(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	holes, err := h.Queries.ListCourseHoles(r.Context(), courseID)
	if err != nil {
		log.Printf("failed to list course holes: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]courseHoleResponse, len(holes))
	for i, ch := range holes {
		resp[i] = toCourseHoleResponse(ch)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetCourseHoleByID godoc
// @Summary     Get a course hole by ID
// @Tags        course-holes
// @Produce     json
// @Param       id  path     int true "Course Hole ID"
// @Success     200 {object} courseHoleResponse
// @Failure     400 {string} string "Invalid ID"
// @Failure     404 {string} string "Course hole not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /course-holes/{id} [get]
func (h CourseHolesHandler) GetCourseHoleByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	hole, err := h.Queries.GetCourseHoleByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Course hole not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get course hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toCourseHoleResponse(hole)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetCourseHoleByCourseAndNumber godoc
// @Summary     Get a course hole by course and hole number
// @Tags        course-holes
// @Produce     json
// @Param       courseId    path     int true "Course ID"
// @Param       holeNumber  path     int true "Hole number"
// @Success     200         {object} courseHoleResponse
// @Failure     400         {string} string "Invalid parameters"
// @Failure     404         {string} string "Course hole not found"
// @Failure     500         {string} string "Internal server error"
// @Router      /course-holes/course/{courseId}/number/{holeNumber} [get]
func (h CourseHolesHandler) GetCourseHoleByCourseAndNumber(w http.ResponseWriter, r *http.Request) {
	courseID, err := strconv.ParseInt(chi.URLParam(r, "courseId"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}
	holeNumber, err := strconv.ParseInt(chi.URLParam(r, "holeNumber"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid hole number", http.StatusBadRequest)
		return
	}

	hole, err := h.Queries.GetCourseHoleByCourseAndNumber(r.Context(), db.GetCourseHoleByCourseAndNumberParams{
		CourseID:   courseID,
		HoleNumber: holeNumber,
	})
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Course hole not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get course hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toCourseHoleResponse(hole)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreateCourseHole godoc
// @Summary     Create a course hole
// @Tags        course-holes
// @Accept      json
// @Produce     json
// @Param       hole body     createCourseHoleRequest true "Hole to create"
// @Success     201  {object} courseHoleResponse
// @Failure     400  {string} string "Invalid request body"
// @Failure     500  {string} string "Internal server error"
// @Router      /course-holes [post]
func (h CourseHolesHandler) CreateCourseHole(w http.ResponseWriter, r *http.Request) {
	var req createCourseHoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.CourseID == 0 || req.HoleNumber == 0 {
		http.Error(w, "course_id and hole_number are required", http.StatusBadRequest)
		return
	}

	params := db.CreateCourseHoleParams{CourseID: req.CourseID, HoleNumber: req.HoleNumber}
	if req.GreenCentreLat != nil {
		params.GreenCentreLat = sql.NullFloat64{Float64: *req.GreenCentreLat, Valid: true}
	}
	if req.GreenCentreLng != nil {
		params.GreenCentreLng = sql.NullFloat64{Float64: *req.GreenCentreLng, Valid: true}
	}

	result, err := h.Queries.CreateCourseHole(r.Context(), params)
	if err != nil {
		log.Printf("failed to create course hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	hole, err := h.Queries.GetCourseHoleByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to fetch created course hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toCourseHoleResponse(hole))
}

// UpdateCourseHoleCoordinates godoc
// @Summary     Update green coordinates for a course hole
// @Tags        course-holes
// @Accept      json
// @Param       id   path int                                true "Course Hole ID"
// @Param       body body updateCourseHoleCoordinatesRequest true "Coordinates"
// @Success     204
// @Failure     400 {string} string "Invalid request"
// @Failure     500 {string} string "Internal server error"
// @Router      /course-holes/{id}/coordinates [patch]
func (h CourseHolesHandler) UpdateCourseHoleCoordinates(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var req updateCourseHoleCoordinatesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := db.UpdateCourseHoleCoordinatesParams{ID: id}
	if req.GreenCentreLat != nil {
		params.GreenCentreLat = sql.NullFloat64{Float64: *req.GreenCentreLat, Valid: true}
	}
	if req.GreenCentreLng != nil {
		params.GreenCentreLng = sql.NullFloat64{Float64: *req.GreenCentreLng, Valid: true}
	}

	if err = h.Queries.UpdateCourseHoleCoordinates(r.Context(), params); err != nil {
		log.Printf("failed to update course hole coordinates: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteCourseHole godoc
// @Summary     Delete a course hole
// @Tags        course-holes
// @Param       id path int true "Course Hole ID"
// @Success     204
// @Failure     400 {string} string "Invalid ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /course-holes/{id} [delete]
func (h CourseHolesHandler) DeleteCourseHole(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeleteCourseHole(r.Context(), id); err != nil {
		log.Printf("failed to delete course hole: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
