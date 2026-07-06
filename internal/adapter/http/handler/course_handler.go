package handler

import (
	"errors"
	"net/http"
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
	reorder       usecasecourse.ReorderCoursesUseCase
	courses       port.CourseRepository
	storage       port.FileStorage
}

func NewCourseHandler(
	create usecasecourse.CreateCourseUseCase,
	list usecasecourse.ListCoursesUseCase,
	update usecasecourse.UpdateCourseUseCase,
	registerStudy usecasecourse.RegisterStudyUseCase,
	delete usecasecourse.DeleteCourseUseCase,
	reorder usecasecourse.ReorderCoursesUseCase,
	courses port.CourseRepository,
	storage port.FileStorage,
) CourseHandler {
	return CourseHandler{
		create: create, list: list, update: update,
		registerStudy: registerStudy, delete: delete, reorder: reorder,
		courses: courses, storage: storage,
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
	contentType, ok := allowedImageTypes[ext]
	if !ok {
		writeError(w, http.StatusBadRequest, "only jpg, png and webp are allowed")
		return
	}

	filename := "courses/" + courseID.String() + ext
	coverURL, err := h.storage.Upload(r.Context(), filename, file, contentType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upload cover")
		return
	}

	if err := h.courses.UpdateCover(r.Context(), courseID, coverURL); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update cover url")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"cover_url": coverURL})
}

func (h CourseHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.ReorderCoursesRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, s := range req.IDs {
		id, err := uuid.Parse(s)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid course id: "+s)
			return
		}
		ids = append(ids, id)
	}

	if err := h.reorder.Execute(r.Context(), userID, ids); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder courses")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
