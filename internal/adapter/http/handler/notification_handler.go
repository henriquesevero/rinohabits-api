package handler

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/postgres"
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
	Endpoint string `json:"endpoint"`
	P256DH   string `json:"p256dh"`
	Auth     string `json:"auth"`
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

	sub := &notification.PushSubscription{
		ID:       uuid.New(),
		UserID:   userID,
		Endpoint: req.Endpoint,
		P256DH:   req.P256DH,
		Auth:     req.Auth,
	}

	if err := h.repo.Save(r.Context(), sub); err != nil {
		log.Printf("notifications: save subscription error: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to save subscription")
		return
	}

	log.Printf("notifications: subscription saved for user %s", userID)
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

