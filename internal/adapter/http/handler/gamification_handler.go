package handler

import (
	"net/http"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/dto"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	"github.com/henriquesevero/rinohabits-api/internal/usecase/stats"
)

type GamificationHandler struct {
	getGamification stats.GetGamificationUseCase
	getRanking      stats.GetRankingUseCase
}

func NewGamificationHandler(
	getGamification stats.GetGamificationUseCase,
	getRanking stats.GetRankingUseCase,
) GamificationHandler {
	return GamificationHandler{getGamification: getGamification, getRanking: getRanking}
}

func (h GamificationHandler) MyStats(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	result, err := h.getGamification.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to compute gamification stats")
		return
	}

	writeJSON(w, http.StatusOK, dto.GamificationResponse{
		TotalXP:              result.TotalXP,
		Level:                result.Level,
		XPInCurrentLevel:     result.XPInCurrentLevel,
		XPForNextLevel:       result.XPForNextLevel,
		CurrentStreak:        result.CurrentStreak,
		PerfectDaysThisMonth: result.PerfectDaysThisMonth,
		ActiveDaysThisMonth:  result.ActiveDaysThisMonth,
		MonthlyPct:           result.MonthlyPct,
		MonthlyMedal:         result.MonthlyMedal,
		TotalPagesRead:       result.TotalPagesRead,
	})
}

func (h GamificationHandler) Ranking(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	entries, err := h.getRanking.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load ranking")
		return
	}

	resp := make([]dto.RankEntryResponse, 0, len(entries))
	for _, e := range entries {
		resp = append(resp, dto.RankEntryResponse{
			UserID:        e.User.ID.String(),
			Name:          e.User.Name,
			AvatarURL:     e.User.AvatarURL,
			TotalXP:       e.Result.TotalXP,
			Level:         e.Result.Level,
			CurrentStreak: e.Result.CurrentStreak,
			MonthlyMedal:  e.Result.MonthlyMedal,
			Rank:          e.Rank,
			IsCurrentUser: e.User.ID == userID,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}
