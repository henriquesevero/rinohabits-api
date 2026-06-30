package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/dto"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	domainhabit "github.com/henriquesevero/rinohabits-api/internal/domain/habit"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type HabitHandler struct {
	create     usecasehabit.CreateHabitUseCase
	listToday  usecasehabit.ListTodayHabitsUseCase
	toggleLog  usecasehabit.ToggleHabitLogUseCase
	calcStreak usecasehabit.CalculateStreakUseCase
	delete     usecasehabit.DeleteHabitUseCase
}

func NewHabitHandler(
	create usecasehabit.CreateHabitUseCase,
	listToday usecasehabit.ListTodayHabitsUseCase,
	toggleLog usecasehabit.ToggleHabitLogUseCase,
	calcStreak usecasehabit.CalculateStreakUseCase,
	delete usecasehabit.DeleteHabitUseCase,
) HabitHandler {
	return HabitHandler{create: create, listToday: listToday, toggleLog: toggleLog, calcStreak: calcStreak, delete: delete}
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
		UserID:         userID,
		Name:           req.Name,
		Icon:           req.Icon,
		Color:          req.Color,
		ActiveWeekdays: req.ActiveWeekdays,
		MonthlyTarget:  req.MonthlyTarget,
	})
	if err != nil {
		if errors.Is(err, domainhabit.ErrNoActiveWeekday) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to create habit")
		return
	}

	writeJSON(w, http.StatusCreated, toHabitResponse(created))
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
			Habit:       toHabitResponse(th.Habit),
			IsCompleted: th.IsCompleted,
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

func toHabitResponse(h *domainhabit.Habit) dto.HabitResponse {
	return dto.HabitResponse{
		ID:             h.ID.String(),
		Name:           h.Name,
		Icon:           h.Icon,
		Color:          h.Color,
		ActiveWeekdays: h.ActiveWeekdays,
		MonthlyTarget:  h.MonthlyTarget,
	}
}
