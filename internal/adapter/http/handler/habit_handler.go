package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/dto"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type HabitHandler struct {
	create     usecasehabit.CreateHabitUseCase
	listToday  usecasehabit.ListTodayHabitsUseCase
	toggleLog  usecasehabit.ToggleHabitLogUseCase
	calcStreak usecasehabit.CalculateStreakUseCase
	update     usecasehabit.UpdateHabitUseCase
	delete     usecasehabit.DeleteHabitUseCase
	habits     port.HabitRepository
}

func NewHabitHandler(
	create usecasehabit.CreateHabitUseCase,
	listToday usecasehabit.ListTodayHabitsUseCase,
	toggleLog usecasehabit.ToggleHabitLogUseCase,
	calcStreak usecasehabit.CalculateStreakUseCase,
	update usecasehabit.UpdateHabitUseCase,
	delete usecasehabit.DeleteHabitUseCase,
	habits port.HabitRepository,
) HabitHandler {
	return HabitHandler{create: create, listToday: listToday, toggleLog: toggleLog, calcStreak: calcStreak, update: update, delete: delete, habits: habits}
}

func (h HabitHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.CreateHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	created, err := h.create.Execute(r.Context(), usecasehabit.CreateHabitInput{
		UserID:          userID,
		Name:            req.Name,
		Icon:            req.Icon,
		Color:           req.Color,
		ActiveWeekdays:  req.ActiveWeekdays,
		WeeklyFrequency: req.WeeklyFrequency,
		MonthlyTarget:   req.MonthlyTarget,
	})
	if err != nil {
		if errors.Is(err, domainhabit.ErrNoSchedule) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create habit")
		return
	}

	writeJSON(w, http.StatusCreated, toHabitResponse(created))
}

func (h HabitHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	habits, err := h.habits.ListActiveByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list habits")
		return
	}

	resp := make([]dto.HabitResponse, 0, len(habits))
	for _, h := range habits {
		resp = append(resp, toHabitResponse(h))
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h HabitHandler) Today(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	todayHabits, date, err := h.listToday.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load today's habits")
		return
	}

	streak, err := h.calcStreak.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to calculate streak")
		return
	}

	responses := make([]dto.TodayHabitResponse, 0, len(todayHabits))
	for _, th := range todayHabits {
		responses = append(responses, dto.TodayHabitResponse{
			Habit:           toHabitResponse(th.Habit),
			IsCompleted:     th.IsCompleted,
			WeekCompletions: th.WeekCompletions,
		})
	}

	writeJSON(w, http.StatusOK, dto.TodayDashboardResponse{
		Date:   date.Format("2006-01-02"),
		Habits: responses,
		Streak: streak,
	})
}

func (h HabitHandler) ToggleLog(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	habitID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	isCompleted, err := h.toggleLog.Execute(r.Context(), usecasehabit.ToggleHabitLogInput{UserID: userID, HabitID: habitID})
	if err != nil {
		if errors.Is(err, domainhabit.ErrNotFound) {
			writeError(w, http.StatusNotFound, "habit not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to toggle habit log")
		return
	}

	writeJSON(w, http.StatusOK, dto.ToggleHabitLogResponse{IsCompleted: isCompleted})
}

func (h HabitHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	habitID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	var req dto.UpdateHabitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	updated, err := h.update.Execute(r.Context(), usecasehabit.UpdateHabitInput{
		UserID:          userID,
		HabitID:         habitID,
		Name:            req.Name,
		Icon:            req.Icon,
		Color:           req.Color,
		ActiveWeekdays:  req.ActiveWeekdays,
		WeeklyFrequency: req.WeeklyFrequency,
		MonthlyTarget:   req.MonthlyTarget,
	})
	if err != nil {
		if errors.Is(err, domainhabit.ErrNoSchedule) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, domainhabit.ErrNotFound) {
			writeError(w, http.StatusNotFound, "habit not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update habit")
		return
	}

	writeJSON(w, http.StatusOK, toHabitResponse(updated))
}

func (h HabitHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	habitID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid habit id")
		return
	}

	if err := h.delete.Execute(r.Context(), userID, habitID); err != nil {
		if errors.Is(err, domainhabit.ErrNotFound) {
			writeError(w, http.StatusNotFound, "habit not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete habit")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h HabitHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.ReorderHabitsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ids := make([]uuid.UUID, 0, len(req.IDs))
	for _, raw := range req.IDs {
		id, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid habit id: "+raw)
			return
		}
		ids = append(ids, id)
	}

	if err := h.habits.ReorderHabits(r.Context(), userID, ids); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reorder habits")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func toHabitResponse(h *domainhabit.Habit) dto.HabitResponse {
	return dto.HabitResponse{
		ID:              h.ID.String(),
		Name:            h.Name,
		Icon:            h.Icon,
		Color:           h.Color,
		ActiveWeekdays:  h.ActiveWeekdays,
		WeeklyFrequency: h.WeeklyFrequency,
		MonthlyTarget:   h.MonthlyTarget,
	}
}
