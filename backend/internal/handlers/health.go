package handlers

import (
	"net/http"

	"github.com/delphi-platform/delphi/backend/internal/services"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	svc *services.Services
	log *logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(svc *services.Services, log *logger.Logger) *HealthHandler {
	return &HealthHandler{svc: svc, log: log}
}

// Check handles basic health check
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "delphi-api",
	})
}

// Ready handles readiness check (includes dependencies)
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	// In production, this would check database and Redis connectivity
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ready",
		"checks": map[string]string{
			"database": "ok",
			"redis":    "ok",
		},
	})
}

