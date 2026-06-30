package dto

type HabitProgressResponse struct {
	HabitID        string  `json:"habit_id"`
	Name           string  `json:"name"`
	Icon           string  `json:"icon"`
	Color          string  `json:"color"`
	RequiredCount  int     `json:"required_count"`
	CompletedCount int     `json:"completed_count"`
	Percentage     float64 `json:"percentage"`
}

type PeriodOverviewResponse struct {
	PeriodType        string                  `json:"period_type"`
	Offset            int                     `json:"offset"`
	StartDate         string                  `json:"start_date"`
	EndDate           string                  `json:"end_date"`
	OverallPercentage float64                 `json:"overall_percentage"`
	RequiredTotal     int                     `json:"required_total"`
	CompletedTotal    int                     `json:"completed_total"`
	Habits            []HabitProgressResponse `json:"habits"`
}

type TrendPointResponse struct {
	Label      string  `json:"label"`
	Percentage float64 `json:"percentage"`
}

type DailyStatusResponse struct {
	Date           string  `json:"date"`
	RequiredCount  int     `json:"required_count"`
	CompletedCount int     `json:"completed_count"`
	Percentage     float64 `json:"percentage"`
}

type CalendarHabitResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
}

type CalendarDayResponse struct {
	Date              string   `json:"date"`
	Status            string   `json:"status"`
	RequiredCount     int      `json:"required_count"`
	CompletedCount    int      `json:"completed_count"`
	CompletedHabitIDs []string `json:"completed_habit_ids"`
}

type CalendarResponse struct {
	Days        []CalendarDayResponse   `json:"days"`
	ActiveDays  int                     `json:"active_days"`
	PerfectDays int                     `json:"perfect_days"`
	TotalChecks int                     `json:"total_checks"`
	TotalHabits int                     `json:"total_habits"`
	Habits      []CalendarHabitResponse `json:"habits"`
}
