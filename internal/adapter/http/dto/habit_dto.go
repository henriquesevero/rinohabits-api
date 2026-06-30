package dto

type CreateHabitRequest struct {
	Name           string `json:"name"`
	Icon           string `json:"icon"`
	Color          string `json:"color"`
	ActiveWeekdays []int  `json:"active_weekdays"`
	MonthlyTarget  *int   `json:"monthly_target"`
}

type HabitResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Icon           string `json:"icon"`
	Color          string `json:"color"`
	ActiveWeekdays []int  `json:"active_weekdays"`
	MonthlyTarget  *int   `json:"monthly_target"`
}

type TodayHabitResponse struct {
	Habit       HabitResponse `json:"habit"`
	IsCompleted bool          `json:"is_completed"`
}

type TodayDashboardResponse struct {
	Date   string               `json:"date"`
	Habits []TodayHabitResponse `json:"habits"`
	Streak int                  `json:"streak"`
}

type ToggleHabitLogResponse struct {
	IsCompleted bool `json:"is_completed"`
}
