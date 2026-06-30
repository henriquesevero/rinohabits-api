package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/dto"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/usecase/auth"
)

type AuthHandler struct {
	register auth.RegisterUseCase
	login    auth.LoginUseCase
	me       auth.GetCurrentUserUseCase
}

func NewAuthHandler(register auth.RegisterUseCase, login auth.LoginUseCase, me auth.GetCurrentUserUseCase) AuthHandler {
	return AuthHandler{register: register, login: login, me: me}
}

func (h AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "name, email and password are required")
		return
	}
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}
	if req.Timezone == "" {
		req.Timezone = "UTC"
	}

	if _, err := h.register.Execute(r.Context(), auth.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Timezone: req.Timezone,
	}); err != nil {
		if errors.Is(err, user.ErrEmailAlreadyRegistered) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to register user")
		return
	}

	token, err := h.login.Execute(r.Context(), auth.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to authenticate after registration")
		return
	}

	writeJSON(w, http.StatusCreated, dto.AuthResponse{Token: token})
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	token, err := h.login.Execute(r.Context(), auth.LoginInput{Email: req.Email, Password: req.Password, Timezone: req.Timezone})
	if err != nil {
		if errors.Is(err, user.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to login")
		return
	}

	writeJSON(w, http.StatusOK, dto.AuthResponse{Token: token})
}

func (h AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	u, err := h.me.Execute(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	writeJSON(w, http.StatusOK, dto.UserResponse{ID: u.ID.String(), Name: u.Name, Email: u.Email})
}
