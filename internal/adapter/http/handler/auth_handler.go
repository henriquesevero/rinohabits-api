package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/dto"
	"github.com/henriquesevero/rinohabits-api/internal/adapter/http/middleware"
	"github.com/henriquesevero/rinohabits-api/internal/domain/user"
	"github.com/henriquesevero/rinohabits-api/internal/port"
	"github.com/henriquesevero/rinohabits-api/internal/usecase/auth"
	usecasebook "github.com/henriquesevero/rinohabits-api/internal/usecase/book"
	usecasecourse "github.com/henriquesevero/rinohabits-api/internal/usecase/course"
	usecasehabit "github.com/henriquesevero/rinohabits-api/internal/usecase/habit"
)

type AuthHandler struct {
	register       auth.RegisterUseCase
	login          auth.LoginUseCase
	me             auth.GetCurrentUserUseCase
	changeEmail    auth.ChangeEmailUseCase
	changePassword auth.ChangePasswordUseCase
	deleteAccount  auth.DeleteAccountUseCase
	resetHabits    usecasehabit.ResetHabitsUseCase
	resetBooks     usecasebook.ResetBooksUseCase
	resetCourses   usecasecourse.ResetCoursesUseCase
	users          port.UserRepository
	storage        port.FileStorage
}

func NewAuthHandler(
	register auth.RegisterUseCase,
	login auth.LoginUseCase,
	me auth.GetCurrentUserUseCase,
	changeEmail auth.ChangeEmailUseCase,
	changePassword auth.ChangePasswordUseCase,
	deleteAccount auth.DeleteAccountUseCase,
	resetHabits usecasehabit.ResetHabitsUseCase,
	resetBooks usecasebook.ResetBooksUseCase,
	resetCourses usecasecourse.ResetCoursesUseCase,
	users port.UserRepository,
	storage port.FileStorage,
) AuthHandler {
	return AuthHandler{
		register: register, login: login, me: me,
		changeEmail: changeEmail, changePassword: changePassword, deleteAccount: deleteAccount,
		resetHabits: resetHabits, resetBooks: resetBooks, resetCourses: resetCourses,
		users: users, storage: storage,
	}
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
		Name:       req.Name,
		Email:      req.Email,
		Password:   req.Password,
		Timezone:   req.Timezone,
		InviteCode: req.InviteCode,
	}); err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidInviteCode):
			writeError(w, http.StatusForbidden, err.Error())
		case errors.Is(err, user.ErrEmailAlreadyRegistered):
			writeError(w, http.StatusConflict, err.Error())
		default:
			writeError(w, http.StatusInternalServerError, "failed to register user")
		}
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

	writeJSON(w, http.StatusOK, dto.UserResponse{
		ID:                    u.ID.String(),
		Name:                  u.Name,
		Email:                 u.Email,
		AvatarURL:             u.AvatarURL,
		BookCollectionOrder:   u.BookCollectionOrder,
		CourseCollectionOrder: u.CourseCollectionOrder,
	})
}

func (h AuthHandler) UploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large (max 5MB)")
		return
	}

	file, header, err := r.FormFile("avatar")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing avatar file")
		return
	}
	defer func() { _ = file.Close() }()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	contentType, ok := allowedImageTypes[ext]
	if !ok {
		writeError(w, http.StatusBadRequest, "only jpg, png and webp are allowed")
		return
	}
	if err := validateImageContent(file, contentType); err != nil {
		writeError(w, http.StatusBadRequest, "file content does not match its extension")
		return
	}

	filename := "avatars/" + userID.String() + ext
	avatarURL, err := h.storage.Upload(r.Context(), filename, file, contentType)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to upload avatar")
		return
	}

	if err := h.users.UpdateAvatarURL(r.Context(), userID, avatarURL); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save avatar")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"avatar_url": avatarURL})
}

func (h AuthHandler) UpdateBookCollectionOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.UpdateCollectionOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.users.UpdateBookCollectionOrder(r.Context(), userID, req.Order); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update book collection order")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h AuthHandler) UpdateCourseCollectionOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.UpdateCollectionOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.users.UpdateCourseCollectionOrder(r.Context(), userID, req.Order); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update course collection order")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h AuthHandler) ChangeEmail(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.ChangeEmailRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.NewEmail == "" || req.CurrentPassword == "" {
		writeError(w, http.StatusBadRequest, "new_email and current_password are required")
		return
	}

	err := h.changeEmail.Execute(r.Context(), auth.ChangeEmailInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewEmail:        req.NewEmail,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrWrongPassword):
			writeError(w, http.StatusUnauthorized, "Senha atual incorreta.")
		case errors.Is(err, user.ErrEmailAlreadyRegistered):
			writeError(w, http.StatusConflict, "Este e-mail já está em uso.")
		default:
			writeError(w, http.StatusInternalServerError, "failed to update email")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.ChangePasswordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "current_password and new_password are required")
		return
	}
	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "Nova senha deve ter pelo menos 8 caracteres.")
		return
	}

	if err := h.changePassword.Execute(r.Context(), auth.ChangePasswordInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	}); err != nil {
		if errors.Is(err, user.ErrWrongPassword) {
			writeError(w, http.StatusUnauthorized, "Senha atual incorreta.")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to update password")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h AuthHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}

	var req dto.DeleteAccountRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.CurrentPassword == "" {
		writeError(w, http.StatusBadRequest, "current_password is required")
		return
	}

	if err := h.deleteAccount.Execute(r.Context(), auth.DeleteAccountInput{
		UserID:          userID,
		CurrentPassword: req.CurrentPassword,
	}); err != nil {
		if errors.Is(err, user.ErrWrongPassword) {
			writeError(w, http.StatusUnauthorized, "Senha atual incorreta.")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to delete account")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h AuthHandler) ResetHabits(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	if err := h.resetHabits.Execute(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reset habits")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h AuthHandler) ResetBooks(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	if err := h.resetBooks.Execute(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reset books")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h AuthHandler) ResetCourses(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.UserIDFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "missing authenticated user")
		return
	}
	if err := h.resetCourses.Execute(r.Context(), userID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to reset courses")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
