package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/postgres"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/push"
	"github.com/henriquesevero/rinohabits-api/internal/domain/notification"
)

type NotificationHandler struct {
	repo            postgres.PushSubscriptionRepository
	vapidPublicKey  string
	vapidPrivateKey string
	vapidEmail      string
}

func NewNotificationHandler(repo postgres.PushSubscriptionRepository, pubKey, privKey, email string) NotificationHandler {
	return NotificationHandler{repo: repo, vapidPublicKey: pubKey, vapidPrivateKey: privKey, vapidEmail: email}
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
		log.Printf("notifications: save subscription error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to save subscription")
		return
	}

	log.Printf("notifications: subscription saved for user %s at %02d:%02d", userID, hour, minute)
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

// TestNotify sends an immediate test notification to all subscriptions of the current user.
func (h NotificationHandler) TestNotify(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	targets, err := h.repo.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list subscriptions")
		return
	}
	if len(targets) == 0 {
		writeError(w, http.StatusNotFound, "no subscriptions found for this user")
		return
	}

	var lastErr error
	for _, t := range targets {
		if err := push.Send(t, "RinoHabits", "Notificação de teste — funcionou! 🎉", h.vapidPublicKey, h.vapidPrivateKey, h.vapidEmail); err != nil {
			log.Printf("notifications: test send error: %v", err)
			lastErr = err
		}
	}

	if lastErr != nil {
		writeError(w, http.StatusBadGateway, "push send failed: "+lastErr.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
