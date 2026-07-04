package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/postgres"
	"github.com/henriquesevero/rinohabits-api/internal/domain/notification"
)

type NotificationHandler struct {
	repo postgres.PushSubscriptionRepository
}

func NewNotificationHandler(repo postgres.PushSubscriptionRepository) NotificationHandler {
	return NotificationHandler{repo: repo}
}

type subscribeRequest struct {
	Endpoint       string `json:"endpoint"`
	P256DH         string `json:"p256dh"`
	Auth           string `json:"auth"`
	ReminderHour   int    `json:"reminder_hour"`
	ReminderMinute int    `json:"reminder_minute"`
}

func (h NotificationHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Endpoint == "" {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	hour := req.ReminderHour
	if hour < 0 || hour > 23 {
		hour = 20
	}
	minute := req.ReminderMinute
	if minute < 0 || minute > 59 {
		minute = 0
	}

	sub := &notification.PushSubscription{
		ID:             uuid.New(),
		UserID:         userID,
		Endpoint:       req.Endpoint,
		P256DH:         req.P256DH,
		Auth:           req.Auth,
		ReminderHour:   hour,
		ReminderMinute: minute,
	}

	if err := h.repo.Save(r.Context(), sub); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h NotificationHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req struct {
		Endpoint string `json:"endpoint"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Endpoint == "" {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.repo.DeleteByEndpoint(r.Context(), userID, req.Endpoint); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
