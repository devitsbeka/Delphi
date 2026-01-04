package handlers

import (
	"net/http"

	"github.com/delphi-platform/delphi/backend/internal/services"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	svc *services.AuthService
	log *logger.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(svc *services.AuthService, log *logger.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, log: log}
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req services.LoginRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	tokens, user, err := h.svc.Login(r.Context(), &req)
	if err != nil {
		h.log.Warnw("login failed", "email", req.Email, "error", err)
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"tokens": tokens,
		"user":   user,
	})
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req services.RegisterRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" || req.TenantName == "" || req.TenantSlug == "" {
		respondError(w, http.StatusBadRequest, "email, password, name, tenant_name, and tenant_slug are required")
		return
	}

	tokens, user, err := h.svc.Register(r.Context(), &req)
	if err != nil {
		h.log.Warnw("registration failed", "email", req.Email, "error", err)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"tokens": tokens,
		"user":   user,
	})
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		respondError(w, http.StatusBadRequest, "refresh_token is required")
		return
	}

	tokens, err := h.svc.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	respondJSON(w, http.StatusOK, tokens)
}

// ForgotPassword handles password reset request
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		respondError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Always return success to prevent email enumeration
	_ = h.svc.ForgotPassword(r.Context(), req.Email)

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "if the email exists, a password reset link has been sent",
	})
}

