package dto

type CreateCourseRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Link        string   `json:"link"`
	TotalHours  *float64 `json:"total_hours"`
	Status      string   `json:"status"`
	Collection  *string  `json:"collection"`
}

type UpdateCourseRequest struct {
	Title       *string  `json:"title"`
	Description *string  `json:"description"`
	Link        *string  `json:"link"`
	TotalHours  *float64 `json:"total_hours"`
	Status      string   `json:"status"`
	Collection  *string  `json:"collection"`
}

type RegisterStudyRequest struct {
	HoursLoggedNow float64 `json:"hours_logged_now"`
}

type ReorderCoursesRequest struct {
	IDs []string `json:"ids"`
}

type StudyStatsResponse struct {
	PeriodType      string  `json:"period_type"`
	Offset          int     `json:"offset"`
	StartDate       string  `json:"start_date"`
	EndDate         string  `json:"end_date"`
	HoursStudied    float64 `json:"hours_studied"`
	CoursesFinished int     `json:"courses_finished"`
}

type CourseResponse struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Link         string   `json:"link"`
	Status       string   `json:"status"`
	TotalHours   *float64 `json:"total_hours"`
	CurrentHours float64  `json:"current_hours"`
	Percentage   float64  `json:"percentage"`
	Collection   *string  `json:"collection"`
	CoverURL     *string  `json:"cover_url"`
	StartedAt    *string  `json:"started_at"`
	FinishedAt   *string  `json:"finished_at"`
}
