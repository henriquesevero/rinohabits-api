package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/dto"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	"github.com/henriquesevero/rinohabits-api/internal/usecase/stats"
)

type StatsHandler struct {
	overview       stats.GetPeriodOverviewUseCase
	trend          stats.GetTrendUseCase
	calendar       stats.GetCalendarUseCase
	dailyBreakdown stats.GetDailyBreakdownUseCase
}

func NewStatsHandler(
	overview stats.GetPeriodOverviewUseCase,
	trend stats.GetTrendUseCase,
	calendar stats.GetCalendarUseCase,
	dailyBreakdown stats.GetDailyBreakdownUseCase,
) StatsHandler {
	return StatsHandler{overview: overview, trend: trend, calendar: calendar, dailyBreakdown: dailyBreakdown}
}

func (h StatsHandler) Overview(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	periodType := stats.PeriodType(r.URL.Query().Get("period"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	overview, err := h.overview.Execute(r.Context(), userID, periodType, offset)
	if err != nil {
		if errors.Is(err, stats.ErrInvalidPeriodType) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load period overview")
		return
	}

	habitsResp := make([]dto.HabitProgressResponse, 0, len(overview.Habits))
	for _, hp := range overview.Habits {
		habitsResp = append(habitsResp, dto.HabitProgressResponse{
			HabitID:        hp.Habit.ID.String(),
			Name:           hp.Habit.Name,
			Icon:           hp.Habit.Icon,
			Color:          hp.Habit.Color,
			RequiredCount:  hp.RequiredCount,
			CompletedCount: hp.CompletedCount,
			Percentage:     hp.Percentage,
		})
	}

	writeJSON(w, http.StatusOK, dto.PeriodOverviewResponse{
		PeriodType:        string(overview.PeriodType),
		Offset:            overview.Offset,
		StartDate:         overview.Start.Format("2006-01-02"),
		EndDate:           overview.End.Format("2006-01-02"),
		OverallPercentage: overview.OverallPercentage,
		RequiredTotal:     overview.RequiredTotal,
		CompletedTotal:    overview.CompletedTotal,
		Habits:            habitsResp,
	})
}

func (h StatsHandler) Trend(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	periodType := stats.PeriodType(r.URL.Query().Get("period"))

	count, err := strconv.Atoi(r.URL.Query().Get("count"))
	if err != nil || count <= 0 {
		count = 8
	}
	if count > 52 {
		count = 52
	}

	points, err := h.trend.Execute(r.Context(), userID, periodType, count)
	if err != nil {
		if errors.Is(err, stats.ErrInvalidPeriodType) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load trend")
		return
	}

	resp := make([]dto.TrendPointResponse, 0, len(points))
	for _, p := range points {
		resp = append(resp, dto.TrendPointResponse{Label: p.Label, Percentage: p.Percentage})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h StatsHandler) DailyBreakdown(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	periodType := stats.PeriodType(r.URL.Query().Get("period"))
	if periodType == "" {
		periodType = stats.PeriodWeek
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	days, err := h.dailyBreakdown.Execute(r.Context(), userID, periodType, offset)
	if err != nil {
		if errors.Is(err, stats.ErrInvalidPeriodType) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to load daily breakdown")
		return
	}

	resp := make([]dto.DailyStatusResponse, 0, len(days))
	for _, d := range days {
		resp = append(resp, dto.DailyStatusResponse{
			Date:           d.Date.Format("2006-01-02"),
			RequiredCount:  d.RequiredCount,
			CompletedCount: d.CompletedCount,
			Percentage:     d.Percentage,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h StatsHandler) Calendar(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	year, err := strconv.Atoi(r.URL.Query().Get("year"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid year")
		return
	}

	monthInt, err := strconv.Atoi(r.URL.Query().Get("month"))
	if err != nil || monthInt < 1 || monthInt > 12 {
		writeError(w, http.StatusBadRequest, "invalid month")
		return
	}

	summary, err := h.calendar.Execute(r.Context(), userID, year, time.Month(monthInt))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load calendar")
		return
	}

	days := make([]dto.CalendarDayResponse, 0, len(summary.Days))
	for _, d := range summary.Days {
		ids := make([]string, 0, len(d.CompletedHabitIDs))
		for _, id := range d.CompletedHabitIDs {
			ids = append(ids, id.String())
		}

		days = append(days, dto.CalendarDayResponse{
			Date:              d.Date.Format("2006-01-02"),
			Status:            string(d.Status),
			RequiredCount:     d.RequiredCount,
			CompletedCount:    d.CompletedCount,
			CompletedHabitIDs: ids,
		})
	}

	habits := make([]dto.CalendarHabitResponse, 0, len(summary.Habits))
	for _, hb := range summary.Habits {
		habits = append(habits, dto.CalendarHabitResponse{
			ID:    hb.ID.String(),
			Name:  hb.Name,
			Icon:  hb.Icon,
			Color: hb.Color,
		})
	}

	writeJSON(w, http.StatusOK, dto.CalendarResponse{
		Days:        days,
		ActiveDays:  summary.ActiveDays,
		PerfectDays: summary.PerfectDays,
		TotalChecks: summary.TotalChecks,
		TotalHabits: summary.TotalHabits,
		Habits:      habits,
	})
}
