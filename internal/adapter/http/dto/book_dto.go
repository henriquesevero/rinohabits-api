package dto

type CreateBookRequest struct {
	Title      string  `json:"title"`
	Author     string  `json:"author"`
	TotalPages *int    `json:"total_pages"`
	Status     string  `json:"status"`
	Collection *string `json:"collection"`
	CoverURL   *string `json:"cover_url"`
}

type BookSearchResult struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	PageCount int    `json:"page_count"`
	CoverURL  string `json:"cover_url"`
}

type UpdateBookRequest struct {
	Title       *string `json:"title"`
	Author      *string `json:"author"`
	TotalPages  *int    `json:"total_pages"`
	Status      string  `json:"status"`
	CurrentPage *int    `json:"current_page"`
	Collection  *string `json:"collection"`
}

type RegisterReadingRequest struct {
	PagesReadNow int `json:"pages_read_now"`
}

type ReorderBooksRequest struct {
	IDs []string `json:"ids"`
}

type BookResponse struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Author      string  `json:"author"`
	Status      string  `json:"status"`
	TotalPages  *int    `json:"total_pages"`
	CurrentPage int     `json:"current_page"`
	Percentage  float64 `json:"percentage"`
	Collection  *string `json:"collection"`
	CoverURL    *string `json:"cover_url"`
	StartedAt   *string `json:"started_at"`
	FinishedAt  *string `json:"finished_at"`
}

type ReadingStatsResponse struct {
	PeriodType    string `json:"period_type"`
	Offset        int    `json:"offset"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	PagesRead     int    `json:"pages_read"`
	BooksFinished int    `json:"books_finished"`
}
