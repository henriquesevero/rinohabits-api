package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	neturl "net/url"
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
	storage         port.FileStorage
	googleBooksKey  string
}

func NewBookHandler(
	create usecasebook.CreateBookUseCase,
	list usecasebook.ListBooksUseCase,
	update usecasebook.UpdateBookUseCase,
	registerReading usecasebook.RegisterReadingUseCase,
	delete usecasebook.DeleteBookUseCase,
	readingStats stats.GetReadingStatsUseCase,
	books port.BookRepository,
	storage port.FileStorage,
	googleBooksKey string,
) BookHandler {
	return BookHandler{
		create: create, list: list, update: update,
		registerReading: registerReading, delete: delete,
		readingStats: readingStats,
		books: books, storage: storage,
		googleBooksKey: googleBooksKey,
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
		Collection: req.Collection,
		CoverURL:   req.CoverURL,
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
		UserID:      userID,
		BookID:      bookID,
		Title:       req.Title,
		Author:      req.Author,
		TotalPages:  req.TotalPages,
		Status:      domainbook.Status(req.Status),
		CurrentPage: req.CurrentPage,
		Collection:  req.Collection,
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

func (h BookHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.ReorderBooksRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, raw := range req.IDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid book id: "+raw)
			return
		}
		ids = append(ids, id)
	}

	if err := h.books.ReorderBooks(r.Context(), userID, ids); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder books")
		return
	}

	w.WriteHeader(http.StatusNoContent)
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
	contentType, ok := allowedImageTypes[ext]
	if !ok {
		writeError(w, http.StatusBadRequest, "only jpg, png and webp are allowed")
		return
	}

	filename := "books/" + bookID.String() + ext
	coverURL, err := h.storage.Upload(r.Context(), filename, file, contentType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upload cover")
		return
	}

	if err := h.books.UpdateCover(r.Context(), bookID, coverURL); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update cover url")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"cover_url": coverURL})
}

func (h BookHandler) SearchBooks(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		writeJSON(w, http.StatusOK, []dto.BookSearchResult{})
		return
	}

	searchType := r.URL.Query().Get("type")   // "title" | "author" | "" (general)
	source := r.URL.Query().Get("source")      // "google" | "" → openlib

	var (
		results []dto.BookSearchResult
		err     error
	)

	if source == "google" {
		results, err = h.searchGoogleBooks(r, q, searchType)
	} else {
		results, err = searchOpenLibrary(r, q, searchType)
	}

	if err != nil {
		log.Printf("books: search error (source=%s): %v", source, err)
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func searchOpenLibrary(r *http.Request, q, searchType string) ([]dto.BookSearchResult, error) {
	params := neturl.Values{}
	switch searchType {
	case "author":
		params.Set("author", q)
	case "title":
		params.Set("title", q)
	default:
		params.Set("q", q)
	}
	params.Set("limit", "15")
	params.Set("fields", "key,title,author_name,number_of_pages_median,cover_i")

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet,
		"https://openlibrary.org/search.json?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build Open Library request")
	}
	req.Header.Set("User-Agent", "rinohabits/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Open Library")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		io.ReadAll(resp.Body) //nolint:errcheck
		return nil, fmt.Errorf("Open Library returned HTTP %d", resp.StatusCode)
	}

	var olResp struct {
		Docs []struct {
			Key                 string   `json:"key"`
			Title               string   `json:"title"`
			AuthorName          []string `json:"author_name"`
			NumberOfPagesMedian *int     `json:"number_of_pages_median"`
			CoverI              *int64   `json:"cover_i"`
		} `json:"docs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&olResp); err != nil {
		return nil, fmt.Errorf("failed to parse Open Library response")
	}

	results := make([]dto.BookSearchResult, 0, len(olResp.Docs))
	for _, doc := range olResp.Docs {
		if doc.Title == "" {
			continue
		}
		pageCount := 0
		if doc.NumberOfPagesMedian != nil {
			pageCount = *doc.NumberOfPagesMedian
		}
		coverURL := ""
		if doc.CoverI != nil {
			coverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-M.jpg", *doc.CoverI)
		}
		results = append(results, dto.BookSearchResult{
			ID:        strings.TrimPrefix(doc.Key, "/works/"),
			Title:     doc.Title,
			Author:    strings.Join(doc.AuthorName, ", "),
			PageCount: pageCount,
			CoverURL:  coverURL,
		})
	}
	return results, nil
}

func (h BookHandler) searchGoogleBooks(r *http.Request, q, searchType string) ([]dto.BookSearchResult, error) {
	var gq string
	switch searchType {
	case "title":
		gq = "intitle:" + q
	case "author":
		gq = "inauthor:" + q
	default:
		gq = q
	}

	params := neturl.Values{}
	params.Set("q", gq)
	params.Set("maxResults", "15")
	params.Set("printType", "books")
	if h.googleBooksKey != "" {
		params.Set("key", h.googleBooksKey)
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet,
		"https://www.googleapis.com/books/v1/volumes?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build Google Books request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Google Books")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		io.ReadAll(resp.Body) //nolint:errcheck
		return nil, fmt.Errorf("Google Books returned HTTP %d", resp.StatusCode)
	}

	var gbResp struct {
		Items []struct {
			ID         string `json:"id"`
			VolumeInfo struct {
				Title      string   `json:"title"`
				Authors    []string `json:"authors"`
				PageCount  int      `json:"pageCount"`
				ImageLinks struct {
					Thumbnail string `json:"thumbnail"`
				} `json:"imageLinks"`
			} `json:"volumeInfo"`
		} `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gbResp); err != nil {
		return nil, fmt.Errorf("failed to parse Google Books response")
	}

	results := make([]dto.BookSearchResult, 0, len(gbResp.Items))
	for _, item := range gbResp.Items {
		if item.VolumeInfo.Title == "" {
			continue
		}
		coverURL := strings.ReplaceAll(item.VolumeInfo.ImageLinks.Thumbnail, "http://", "https://")
		results = append(results, dto.BookSearchResult{
			ID:        "gb_" + item.ID,
			Title:     item.VolumeInfo.Title,
			Author:    strings.Join(item.VolumeInfo.Authors, ", "),
			PageCount: item.VolumeInfo.PageCount,
			CoverURL:  coverURL,
		})
	}
	return results, nil
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
		Collection:  b.Collection,
		CoverURL:    b.CoverURL,
		StartedAt:   startedAt,
		FinishedAt:  finishedAt,
	}
}
