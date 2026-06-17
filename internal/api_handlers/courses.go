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

type CoursesHandler struct {
	Queries *db.Queries
}

type courseResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	GolfAPIID *string   `json:"golf_api_id"`
	CreatedAt time.Time `json:"created_at"`
}

type createCourseRequest struct {
	Name      string  `json:"name"        example:"Augusta National"`
	GolfAPIID *string `json:"golf_api_id" example:"abc123"`
}

type updateCourseRequest struct {
	Name      string  `json:"name"        example:"Augusta National"`
	GolfAPIID *string `json:"golf_api_id" example:"abc123"`
}

func toCourseResponse(c db.Course) courseResponse {
	return courseResponse{
		ID:        c.ID,
		Name:      c.Name,
		GolfAPIID: nullableString(c.GolfApiID),
		CreatedAt: c.CreatedAt,
	}
}

// ListCourses godoc
// @Summary     List all courses
// @Tags        courses
// @Produce     json
// @Success     200 {array}  courseResponse
// @Failure     500 {string} string "Internal server error"
// @Router      /courses [get]
func (h CoursesHandler) ListCourses(w http.ResponseWriter, r *http.Request) {

	courses, err := h.Queries.ListCourses(r.Context())
	if err != nil {
		log.Printf("failed to list courses: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	resp := make([]courseResponse, len(courses))
	for i, c := range courses {
		resp[i] = toCourseResponse(c)
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetCourseByID godoc
// @Summary     Get a course by ID
// @Tags        courses
// @Produce     json
// @Param       id  path     int true "Course ID"
// @Success     200 {object} courseResponse
// @Failure     400 {string} string "Invalid course ID"
// @Failure     404 {string} string "Course not found"
// @Failure     500 {string} string "Internal server error"
// @Router      /courses/{id} [get]
func (h CoursesHandler) GetCourseByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	course, err := h.Queries.GetCourseByID(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get course: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toCourseResponse(course)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// GetCourseByName godoc
// @Summary     Get a course by name
// @Tags        courses
// @Produce     json
// @Param       name path     string true "Course name"
// @Success     200  {object} courseResponse
// @Failure     404  {string} string "Course not found"
// @Failure     500  {string} string "Internal server error"
// @Router      /courses/name/{name} [get]
func (h CoursesHandler) GetCourseByName(w http.ResponseWriter, r *http.Request) {

	course, err := h.Queries.GetCourseByName(r.Context(), chi.URLParam(r, "name"))
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(w, "Course not found", http.StatusNotFound)
		return
	}
	if err != nil {
		log.Printf("failed to get course by name: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(toCourseResponse(course)); err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
	}
}

// CreateCourse godoc
// @Summary     Create a course
// @Tags        courses
// @Accept      json
// @Produce     json
// @Param       course body     createCourseRequest true "Course to create"
// @Success     201    {object} courseResponse
// @Failure     400    {string} string "Invalid request body"
// @Failure     500    {string} string "Internal server error"
// @Router      /courses [post]
func (h CoursesHandler) CreateCourse(w http.ResponseWriter, r *http.Request) {
	var req createCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	params := db.CreateCourseParams{Name: req.Name}
	if req.GolfAPIID != nil {
		params.GolfApiID = sql.NullString{String: *req.GolfAPIID, Valid: true}
	}

	result, err := h.Queries.CreateCourse(r.Context(), params)
	if err != nil {
		log.Printf("failed to create course: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("failed to get last insert id: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	course, err := h.Queries.GetCourseByID(r.Context(), id)
	if err != nil {
		log.Printf("failed to fetch created course: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(toCourseResponse(course))
}

// UpdateCourse godoc
// @Summary     Update a course
// @Tags        courses
// @Accept      json
// @Param       id     path int               true "Course ID"
// @Param       course body updateCourseRequest true "Updated fields"
// @Success     204
// @Failure     400 {string} string "Invalid request"
// @Failure     500 {string} string "Internal server error"
// @Router      /courses/{id} [put]
func (h CoursesHandler) UpdateCourse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	var req updateCourseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	params := db.UpdateCourseParams{Name: req.Name, ID: id}
	if req.GolfAPIID != nil {
		params.GolfApiID = sql.NullString{String: *req.GolfAPIID, Valid: true}
	}

	if err = h.Queries.UpdateCourse(r.Context(), params); err != nil {
		log.Printf("failed to update course: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteCourse godoc
// @Summary     Delete a course
// @Tags        courses
// @Param       id path int true "Course ID"
// @Success     204
// @Failure     400 {string} string "Invalid course ID"
// @Failure     500 {string} string "Internal server error"
// @Router      /courses/{id} [delete]
func (h CoursesHandler) DeleteCourse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid course ID", http.StatusBadRequest)
		return
	}

	if err = h.Queries.DeleteCourse(r.Context(), id); err != nil {
		log.Printf("failed to delete course: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
