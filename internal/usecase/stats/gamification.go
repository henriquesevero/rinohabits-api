package stats

import (
	"math"
	"time"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/domain/dailylog"
	"github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
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

func readingBonusForPages(pages int) int {
	switch {
	case pages >= 300:
		return 300
	case pages >= 150:
		return 150
	case pages >= 50:
		return 75
	case pages >= 1:
		return 25
	default:
		return 0
	}
}

func computeGamification(habits []*habit.Habit, allLogs []*dailylog.DailyLog, monthlyPages []port.MonthlyPages, currentStreak int, timezone string) GamificationResult {
	completedByDate := groupCompletedHabitsByDate(allLogs)

	xp, perfectByMonth, activeByMonth := scoreDailyCompletions(habits, completedByDate, timezone)
	xp += totalReadingBonus(monthlyPages)
	xp += streakBonus(currentStreak)
	xp += monthlyHabitBonus(perfectByMonth, activeByMonth)

	level := levelFromXP(xp)
	xpCurrent := xpForLevel(level)
	xpNext := xpForLevel(level + 1)

	curMonth := currentMonthKey(timezone)
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
		TotalPagesRead:       sumPages(monthlyPages),
	}
}

func sumPages(monthlyPages []port.MonthlyPages) int {
	total := 0
	for _, mp := range monthlyPages {
		total += mp.Pages
	}
	return total
}

func groupCompletedHabitsByDate(logs []*dailylog.DailyLog) map[string]map[uuid.UUID]bool {
	completedByDate := make(map[string]map[uuid.UUID]bool)
	for _, l := range logs {
		dateStr := l.LogDate.Format("2006-01-02")
		if completedByDate[dateStr] == nil {
			completedByDate[dateStr] = make(map[uuid.UUID]bool)
		}
		completedByDate[dateStr][l.HabitID] = true
	}
	return completedByDate
}

func scoreDailyCompletions(habits []*habit.Habit, completedByDate map[string]map[uuid.UUID]bool, timezone string) (xp int, perfectByMonth, activeByMonth map[monthKey]int) {
	perfectByMonth = make(map[monthKey]int)
	activeByMonth = make(map[monthKey]int)

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

		if allCompleted(effective, completedHabits) {
			xp += 50
			perfectByMonth[mk]++
		}
	}

	return xp, perfectByMonth, activeByMonth
}

func allCompleted(habits []*habit.Habit, completedHabits map[uuid.UUID]bool) bool {
	for _, h := range habits {
		if !completedHabits[h.ID] {
			return false
		}
	}
	return true
}

// Reading bonus is a fixed XP tier per month rather than per page, so it resets each month.
func totalReadingBonus(monthlyPages []port.MonthlyPages) int {
	bonus := 0
	for _, mp := range monthlyPages {
		bonus += readingBonusForPages(mp.Pages)
	}
	return bonus
}

func streakBonus(currentStreak int) int {
	switch {
	case currentStreak >= 30:
		return 600
	case currentStreak >= 14:
		return 250
	case currentStreak >= 7:
		return 100
	default:
		return 0
	}
}

func monthlyHabitBonus(perfectByMonth, activeByMonth map[monthKey]int) int {
	bonus := 0
	for mk, perfectDays := range perfectByMonth {
		activeDays := activeByMonth[mk]
		if activeDays == 0 {
			continue
		}
		pct := float64(perfectDays) / float64(activeDays) * 100
		switch {
		case pct >= 100:
			bonus += 700
		case pct >= 75:
			bonus += 300
		case pct >= 50:
			bonus += 150
		}
	}
	return bonus
}

func currentMonthKey(timezone string) monthKey {
	now := time.Now()
	nowLocal, err := usecasehabit.LocalToday(now, timezone)
	if err != nil {
		nowLocal = now.UTC().Truncate(24 * time.Hour)
	}
	return monthKey{nowLocal.Year(), nowLocal.Month()}
}
