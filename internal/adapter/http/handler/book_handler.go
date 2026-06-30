package handler

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/dto"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	domainbook "github.com/henriquesevero/rinohabits-api/internal/domain/book"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasebook "github.com/henriquesevero/rinohabits-api/internal/usecase/book"
	"github.com/henriquesevero/rinohabits-api/internal/usecase/stats"
)

type BookHandler struct {
	create          usecasebook.CreateBookUseCase
	list            usecasebook.ListBooksUseCase
	update          usecasebook.UpdateBookUseCase
	registerReading usecasebook.RegisterReadingUseCase
	delete          usecasebook.DeleteBookUseCase
	readingStats    stats.GetReadingStatsUseCase
	books           port.BookRepository
	uploadsDir      string
	apiBaseURL      string
}

func NewBookHandler(
	create usecasebook.CreateBookUseCase,
	list usecasebook.ListBooksUseCase,
	update usecasebook.UpdateBookUseCase,
	registerReading usecasebook.RegisterReadingUseCase,
	delete usecasebook.DeleteBookUseCase,
	readingStats stats.GetReadingStatsUseCase,
	books port.BookRepository,
	uploadsDir string,
	apiBaseURL string,
) BookHandler {
	return BookHandler{
		create: create, list: list, update: update,
		registerReading: registerReading, delete: delete,
		readingStats: readingStats,
		books: books, uploadsDir: uploadsDir, apiBaseURL: apiBaseURL,
	}
}

func (h BookHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.CreateBookRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, err := h.create.Execute(r.Context(), usecasebook.CreateBookInput{
		UserID:     userID,
		Title:      req.Title,
		Author:     req.Author,
		TotalPages: req.TotalPages,
		Status:     domainbook.Status(req.Status),
	})
	if err != nil {
		if errors.Is(err, domainbook.ErrInvalidTitle) || errors.Is(err, domainbook.ErrInvalidStatus) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create book")
		return
	}

	writeJSON(w, http.StatusCreated, toBookResponse(b))
}

func (h BookHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var statusFilter *domainbook.Status
	if s := r.URL.Query().Get("status"); s != "" {
		st := domainbook.Status(s)
		statusFilter = &st
	}

	books, err := h.list.Execute(r.Context(), userID, statusFilter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list books")
		return
	}

	resp := make([]dto.BookResponse, 0, len(books))
	for _, b := range books {
		resp = append(resp, toBookResponse(b))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h BookHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	bookID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var req dto.UpdateBookRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, err := h.update.Execute(r.Context(), usecasebook.UpdateBookInput{
		UserID:     userID,
		BookID:     bookID,
		Title:      req.Title,
		Author:     req.Author,
		TotalPages: req.TotalPages,
		Status:     domainbook.Status(req.Status),
	})
	if err != nil {
		switch {
		case errors.Is(err, domainbook.ErrNotFound):
			writeError(w, http.StatusNotFound, "book not found")
		case errors.Is(err, domainbook.ErrInvalidStatus):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to update book")
		}
		return
	}

	writeJSON(w, http.StatusOK, toBookResponse(b))
}

func (h BookHandler) RegisterReading(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	bookID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	var req dto.RegisterReadingRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, err := h.registerReading.Execute(r.Context(), usecasebook.RegisterReadingInput{
		UserID:       userID,
		BookID:       bookID,
		PagesReadNow: req.PagesReadNow,
	})
	if err != nil {
		switch {
		case errors.Is(err, domainbook.ErrNotFound):
			writeError(w, http.StatusNotFound, "book not found")
		case errors.Is(err, domainbook.ErrNoProgress):
			writeError(w, http.StatusBadRequest, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to register reading")
		}
		return
	}

	writeJSON(w, http.StatusOK, toBookResponse(b))
}

func (h BookHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	bookID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	if err := h.delete.Execute(r.Context(), userID, bookID); err != nil {
		if errors.Is(err, domainbook.ErrNotFound) {
			writeError(w, http.StatusNotFound, "book not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete book")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h BookHandler) ReadingStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	periodType := stats.PeriodType(r.URL.Query().Get("period"))
	if periodType == "" {
		periodType = stats.PeriodMonth
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	rs, err := h.readingStats.Execute(r.Context(), userID, periodType, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load reading stats")
		return
	}

	writeJSON(w, http.StatusOK, dto.ReadingStatsResponse{
		PeriodType:    string(rs.PeriodType),
		Offset:        rs.Offset,
		StartDate:     rs.Start,
		EndDate:       rs.End,
		PagesRead:     rs.PagesRead,
		BooksFinished: rs.BooksFinished,
	})
}

func (h BookHandler) UploadCover(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	bookID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid book id")
		return
	}

	b, err := h.books.FindByID(r.Context(), bookID)
	if err != nil || b.UserID != userID {
		writeError(w, http.StatusNotFound, "book not found")
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

	dir := filepath.Join(h.uploadsDir, "books")
	if err := os.MkdirAll(dir, 0755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create upload dir")
		return
	}

	filename := fmt.Sprintf("%s%s", bookID.String(), ext)
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

	coverURL := fmt.Sprintf("%s/uploads/books/%s", h.apiBaseURL, filename)
	if err := h.books.UpdateCover(r.Context(), bookID, coverURL); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update cover url")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"cover_url": coverURL})
}

func toBookResponse(b *domainbook.Book) dto.BookResponse {
	var percentage float64
	if b.TotalPages != nil && *b.TotalPages > 0 {
		percentage = float64(b.CurrentPage) / float64(*b.TotalPages) * 100
	}

	var startedAt, finishedAt *string
	if b.StartedAt != nil {
		s := b.StartedAt.Format(time.RFC3339)
		startedAt = &s
	}
	if b.FinishedAt != nil {
		s := b.FinishedAt.Format(time.RFC3339)
		finishedAt = &s
	}

	return dto.BookResponse{
		ID:          b.ID.String(),
		Title:       b.Title,
		Author:      b.Author,
		Status:      string(b.Status),
		TotalPages:  b.TotalPages,
		CurrentPage: b.CurrentPage,
		Percentage:  percentage,
		CoverURL:    b.CoverURL,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
	}
}
