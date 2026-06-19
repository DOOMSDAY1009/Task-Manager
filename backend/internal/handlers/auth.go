package handlers

import (
	"errors"
	"net/http"

	"github.com/hireft/task-manager/internal/auth"
	"github.com/hireft/task-manager/internal/models"
	"github.com/hireft/task-manager/internal/store"
	"github.com/hireft/task-manager/internal/validation"
)

// Signup handles POST /auth/signup.
func (h *Handler) Signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if errs := validation.ValidateSignup(req); len(errs) > 0 {
		writeValidationError(w, errs)
		return
	}

	hash, err := auth.HashPassword(req.Password, h.BcryptCost)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not process password")
		return
	}

	user, err := h.Store.CreateUser(r.Context(), req.Email, hash, models.RoleUser)
	if err != nil {
		if errors.Is(err, store.ErrEmailTaken) {
			writeError(w, http.StatusConflict, "email already registered")
			return
		}
		writeError(w, http.StatusInternalServerError, "could not create user")
		return
	}

	h.issueToken(w, http.StatusCreated, user)
}

// Login handles POST /auth/login.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.Store.GetUserByEmail(r.Context(), req.Email)
	if err != nil || !auth.CheckPassword(user.PasswordHash, req.Password) {
		// Same response whether the email is unknown or the password is wrong,
		// to avoid leaking which emails are registered.
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	h.issueToken(w, http.StatusOK, user)
}

// Me handles GET /auth/me and returns the current authenticated user.
func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user, err := h.currentUser(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *Handler) issueToken(w http.ResponseWriter, status int, user models.User) {
	token, err := h.Auth.Generate(user, h.Now())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "could not issue token")
		return
	}
	user.PasswordHash = ""
	writeJSON(w, status, models.AuthResponse{Token: token, User: user})
}
