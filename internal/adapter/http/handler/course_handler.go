package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/dto"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	domaincourse "github.com/henriquesevero/rinohabits-api/internal/domain/course"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasecourse "github.com/henriquesevero/rinohabits-api/internal/usecase/course"
)

type CourseHandler struct {
	create        usecasecourse.CreateCourseUseCase
	list          usecasecourse.ListCoursesUseCase
	update        usecasecourse.UpdateCourseUseCase
	registerStudy usecasecourse.RegisterStudyUseCase
	delete        usecasecourse.DeleteCourseUseCase
	courses       port.CourseRepository
	uploadsDir    string
	apiBaseURL    string
}

func NewCourseHandler(
	create usecasecourse.CreateCourseUseCase,
	list usecasecourse.ListCoursesUseCase,
	update usecasecourse.UpdateCourseUseCase,
	registerStudy usecasecourse.RegisterStudyUseCase,
	delete usecasecourse.DeleteCourseUseCase,
	courses port.CourseRepository,
	uploadsDir string,
	apiBaseURL string,
) CourseHandler {
	return CourseHandler{
		create: create, list: list, update: update,
		registerStudy: registerStudy, delete: delete,
		courses: courses, uploadsDir: uploadsDir, apiBaseURL: apiBaseURL,
	}
}

func (h CourseHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.CreateCourseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	c, err := h.create.Execute(r.Context(), usecasecourse.CreateCourseInput{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		Link:        req.Link,
		TotalHours:  req.TotalHours,
		Status:      domaincourse.Status(req.Status),
	})
	if err != nil {
		if errors.Is(err, domaincourse.ErrInvalidTitle) || errors.Is(err, domaincourse.ErrInvalidStatus) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create course")
		return
	}

	writeJSON(w, http.StatusCreated, toCourseResponse(c))
}

func (h CourseHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var statusFilter *domaincourse.Status
	if s := r.URL.Query().Get("status"); s != "" {
		st := domaincourse.Status(s)
		statusFilter = &st
	}

	courses, err := h.list.Execute(r.Context(), userID, statusFilter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list courses")
		return
	}

	resp := make([]dto.CourseResponse, 0, len(courses))
	for _, c := range courses {
		resp = append(resp, toCourseResponse(c))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h CourseHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	courseID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	var req dto.UpdateCourseRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	c, err := h.update.Execute(r.Context(), usecasecourse.UpdateCourseInput{
		UserID:      userID,
		CourseID:    courseID,
		Title:       req.Title,
		Description: req.Description,
		Link:        req.Link,
		TotalHours:  req.TotalHours,
		Status:      domaincourse.Status(req.Status),
	})
	if err != nil {
		switch {
		case errors.Is(err, domaincourse.ErrNotFound):
			writeError(w, http.StatusNotFound, "course not found")
		case errors.Is(err, domaincourse.ErrInvalidStatus):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to update course")
		}
		return
	}

	writeJSON(w, http.StatusOK, toCourseResponse(c))
}

func (h CourseHandler) RegisterStudy(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	courseID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	var req dto.RegisterStudyRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	c, err := h.registerStudy.Execute(r.Context(), usecasecourse.RegisterStudyInput{
		UserID:         userID,
		CourseID:       courseID,
		HoursLoggedNow: req.HoursLoggedNow,
	})
	if err != nil {
		switch {
		case errors.Is(err, domaincourse.ErrNotFound):
			writeError(w, http.StatusNotFound, "course not found")
		case errors.Is(err, domaincourse.ErrNoProgress):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to register study session")
		}
		return
	}

	writeJSON(w, http.StatusOK, toCourseResponse(c))
}

func (h CourseHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	courseID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	if err := h.delete.Execute(r.Context(), userID, courseID); err != nil {
		if errors.Is(err, domaincourse.ErrNotFound) {
			writeError(w, http.StatusNotFound, "course not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete course")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h CourseHandler) UploadCover(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	courseID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid course id")
		return
	}

	c, err := h.courses.FindByID(r.Context(), courseID)
	if err != nil || c.UserID != userID {
		writeError(w, http.StatusNotFound, "course not found")
		return
	}

	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large (max 8MB)")
		return
	}

	file, header, err := r.FormFile("cover")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing cover file")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" {
		writeError(w, http.StatusBadRequest, "only jpg, png and webp are allowed")
		return
	}

	dir := filepath.Join(h.uploadsDir, "courses")
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create upload dir")
		return
	}

	filename := fmt.Sprintf("%s%s", courseID.String(), ext)
	dest := filepath.Join(dir, filename)

	out, err := os.Create(dest)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to write file")
		return
	}

	coverURL := fmt.Sprintf("%s/uploads/courses/%s", h.apiBaseURL, filename)
	if err := h.courses.UpdateCover(r.Context(), courseID, coverURL); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update cover url")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"cover_url": coverURL})
}

func toCourseResponse(c *domaincourse.Course) dto.CourseResponse {
	var percentage float64
	if c.TotalHours != nil && *c.TotalHours > 0 {
		percentage = c.CurrentHours / *c.TotalHours * 100
	}

	var startedAt, finishedAt *string
	if c.StartedAt != nil {
		s := c.StartedAt.Format(time.RFC3339)
		startedAt = &s
	}
	if c.FinishedAt != nil {
		s := c.FinishedAt.Format(time.RFC3339)
		finishedAt = &s
	}

	return dto.CourseResponse{
		ID:           c.ID.String(),
		Title:        c.Title,
		Description:  c.Description,
		Link:         c.Link,
		Status:       string(c.Status),
		TotalHours:   c.TotalHours,
		CurrentHours: c.CurrentHours,
		Percentage:   percentage,
		CoverURL:     c.CoverURL,
		StartedAt:    startedAt,
		FinishedAt:   finishedAt,
	}
}
