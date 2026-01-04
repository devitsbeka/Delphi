package handlers

import (
	"net/http"
	"strconv"

	"github.com/delphi-platform/delphi/backend/internal/middleware"
	"github.com/delphi-platform/delphi/backend/internal/services"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// AgentHandler handles agent endpoints
type AgentHandler struct {
	svc *services.AgentService
	log *logger.Logger
}

// NewAgentHandler creates a new agent handler
func NewAgentHandler(svc *services.AgentService, log *logger.Logger) *AgentHandler {
	return &AgentHandler{svc: svc, log: log}
}

// List returns all agents for the tenant
func (h *AgentHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	agents, err := h.svc.List(r.Context(), tenantID)
	if err != nil {
		h.log.Errorw("failed to list agents", "tenant_id", tenantID, "error", err)
		respondError(w, http.StatusInternalServerError, "failed to list agents")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"agents": agents,
		"count":  len(agents),
	})
}

// Create creates a new agent
func (h *AgentHandler) Create(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	var req services.CreateAgentRequest
	if err := decodeJSON(r, &req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}

	agent, err := h.svc.Create(r.Context(), tenantID, &req)
	if err != nil {
		h.log.Errorw("failed to create agent", "tenant_id", tenantID, "error", err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, agent)
}

// Get returns an agent by ID
func (h *AgentHandler) Get(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	agent, err := h.svc.Get(r.Context(), tenantID, agentID)
	if err != nil {
		respondError(w, http.StatusNotFound, "agent not found")
		return
	}

	respondJSON(w, http.StatusOK, agent)
}

// Update updates an agent
func (h *AgentHandler) Update(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	var updates map[string]interface{}
	if err := decodeJSON(r, &updates); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	agent, err := h.svc.Update(r.Context(), tenantID, agentID, updates)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, agent)
}

// Delete deletes an agent
func (h *AgentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	if err := h.svc.Delete(r.Context(), tenantID, agentID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "agent deleted"})
}

// Launch starts an agent
func (h *AgentHandler) Launch(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	agent, err := h.svc.Launch(r.Context(), tenantID, agentID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, agent)
}

// Pause pauses an agent
func (h *AgentHandler) Pause(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	agent, err := h.svc.Pause(r.Context(), tenantID, agentID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, agent)
}

// Terminate terminates an agent
func (h *AgentHandler) Terminate(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	agent, err := h.svc.Terminate(r.Context(), tenantID, agentID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, agent)
}

// ListRuns returns runs for an agent
func (h *AgentHandler) ListRuns(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	limit := 50
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	runs, err := h.svc.ListRuns(r.Context(), tenantID, agentID, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"runs":  runs,
		"count": len(runs),
	})
}

// GetRun returns a specific run
func (h *AgentHandler) GetRun(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := middleware.GetTenantID(r.Context())
	if !ok {
		respondError(w, http.StatusUnauthorized, "tenant context required")
		return
	}

	agentID, err := uuid.Parse(chi.URLParam(r, "agentID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid agent ID")
		return
	}

	runID, err := uuid.Parse(chi.URLParam(r, "runID"))
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid run ID")
		return
	}

	run, err := h.svc.GetRun(r.Context(), tenantID, agentID, runID)
	if err != nil {
		respondError(w, http.StatusNotFound, "run not found")
		return
	}

	respondJSON(w, http.StatusOK, run)
}

// GetRunLogs returns logs for a run
func (h *AgentHandler) GetRunLogs(w http.ResponseWriter, r *http.Request) {
	// Placeholder - would stream logs in production
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"logs": []interface{}{},
	})
}

// ListTemplates returns available agent templates
func (h *AgentHandler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.svc.GetTemplates(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"templates": templates,
		"count":     len(templates),
	})
}

