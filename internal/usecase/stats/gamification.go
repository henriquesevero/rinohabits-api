package stats

import (
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/dailylog"
	"github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type GamificationResult struct {
	TotalXP              int
	Level                int
	XPInCurrentLevel     int
	XPForNextLevel       int
	CurrentStreak        int
	PerfectDaysThisMonth int
	ActiveDaysThisMonth  int
	MonthlyPct           float64
	MonthlyMedal         string
	TotalPagesRead       int
}

type monthKey struct {
	year  int
	month time.Month
}

func xpForLevel(n int) int {
	if n <= 1 {
		return 0
	}
	return 100 * (n - 1) * n / 2
}

func levelFromXP(xp int) int {
	if xp <= 0 {
		return 1
	}
	n := int((1.0 + math.Sqrt(1.0+float64(8*xp)/100)) / 2.0)
	if n < 1 {
		n = 1
	}
	for xpForLevel(n+1) <= xp {
		n++
	}
	return n
}

func medalFromPct(pct float64) string {
	switch {
	case pct >= 100:
		return "diamond"
	case pct >= 90:
		return "platinum"
	case pct >= 75:
		return "gold"
	case pct >= 50:
		return "silver"
	case pct >= 25:
		return "bronze"
	default:
		return "none"
	}
}

func computeGamification(habits []*habit.Habit, allLogs []*dailylog.DailyLog, totalPages int, currentStreak int, timezone string) GamificationResult {
	if len(habits) == 0 && len(allLogs) == 0 {
		pages := (totalPages / 10) * 5
		level := levelFromXP(pages)
		return GamificationResult{
			TotalXP:        pages,
			Level:          level,
			XPInCurrentLevel: pages - xpForLevel(level),
			XPForNextLevel: xpForLevel(level+1) - xpForLevel(level),
			TotalPagesRead: totalPages,
			MonthlyMedal:   "none",
		}
	}

	// Group completed habits by date
	completedByDate := make(map[string]map[uuid.UUID]bool)
	for _, l := range allLogs {
		dateStr := l.LogDate.Format("2006-01-02")
		if completedByDate[dateStr] == nil {
			completedByDate[dateStr] = make(map[uuid.UUID]bool)
		}
		completedByDate[dateStr][l.HabitID] = true
	}

	xp := 0
	perfectByMonth := make(map[monthKey]int)
	activeByMonth := make(map[monthKey]int)

	now := time.Now()
	nowLocal, err := usecasehabit.LocalToday(now, timezone)
	if err != nil {
		nowLocal = now.UTC().Truncate(24 * time.Hour)
	}
	curMonth := monthKey{nowLocal.Year(), nowLocal.Month()}

	for dateStr, completedHabits := range completedByDate {
		date, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			continue
		}

		required := usecasehabit.RequiredHabitsOn(habits, date, timezone)
		if len(required) == 0 {
			continue
		}

		effective := usecasehabit.EffectiveRequiredHabits(required, date, timezone, completedHabits)
		if len(effective) == 0 {
			continue
		}

		mk := monthKey{date.Year(), date.Month()}
		activeByMonth[mk]++

		allCompleted := true
		for _, h := range effective {
			if !completedHabits[h.ID] {
				allCompleted = false
				break
			}
		}

		if allCompleted {
			xp += 50
			perfectByMonth[mk]++
		}
	}

	// Pages XP: 5 per 10 pages
	xp += (totalPages / 10) * 5

	// Streak milestone (current streak only, not cumulative)
	switch {
	case currentStreak >= 30:
		xp += 600
	case currentStreak >= 14:
		xp += 250
	case currentStreak >= 7:
		xp += 100
	}

	// Monthly bonus XP (all months including current)
	for mk, perfectDays := range perfectByMonth {
		activeDays := activeByMonth[mk]
		if activeDays == 0 {
			continue
		}
		pct := float64(perfectDays) / float64(activeDays) * 100
		switch {
		case pct >= 100:
			xp += 700
		case pct >= 75:
			xp += 300
		case pct >= 50:
			xp += 150
		}
	}

	level := levelFromXP(xp)
	xpCurrent := xpForLevel(level)
	xpNext := xpForLevel(level + 1)

	perfectThisMonth := perfectByMonth[curMonth]
	activeThisMonth := activeByMonth[curMonth]
	monthlyPct := 0.0
	if activeThisMonth > 0 {
		monthlyPct = float64(perfectThisMonth) / float64(activeThisMonth) * 100
	}

	return GamificationResult{
		TotalXP:              xp,
		Level:                level,
		XPInCurrentLevel:     xp - xpCurrent,
		XPForNextLevel:       xpNext - xpCurrent,
		CurrentStreak:        currentStreak,
		PerfectDaysThisMonth: perfectThisMonth,
		ActiveDaysThisMonth:  activeThisMonth,
		MonthlyPct:           monthlyPct,
		MonthlyMedal:         medalFromPct(monthlyPct),
		TotalPagesRead:       totalPages,
	}
}
