package dto

type GamificationResponse struct {
	TotalXP              int     `json:"total_xp"`
	Level                int     `json:"level"`
	XPInCurrentLevel     int     `json:"xp_in_current_level"`
	XPForNextLevel       int     `json:"xp_for_next_level"`
	CurrentStreak        int     `json:"current_streak"`
	PerfectDaysThisMonth int     `json:"perfect_days_this_month"`
	ActiveDaysThisMonth  int     `json:"active_days_this_month"`
	MonthlyPct           float64 `json:"monthly_pct"`
	MonthlyMedal         string  `json:"monthly_medal"`
	TotalPagesRead       int     `json:"total_pages_read"`
}

type RankEntryResponse struct {
	UserID        string  `json:"user_id"`
	Name          string  `json:"name"`
	AvatarURL     *string `json:"avatar_url"`
	TotalXP       int     `json:"total_xp"`
	Level         int     `json:"level"`
	CurrentStreak int     `json:"current_streak"`
	MonthlyMedal  string  `json:"monthly_medal"`
	Rank          int     `json:"rank"`
	IsCurrentUser bool    `json:"is_current_user"`
}
