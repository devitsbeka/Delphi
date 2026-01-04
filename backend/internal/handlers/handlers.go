package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/delphi-platform/delphi/backend/internal/services"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
)

// Handlers contains all handler instances
type Handlers struct {
	Health     *HealthHandler
	Auth       *AuthHandler
	User       *UserHandler
	Tenant     *TenantHandler
	APIKey     *APIKeyHandler
	Agent      *AgentHandler
	Execute    *ExecuteHandler
	Knowledge  *KnowledgeHandler
	Repository *RepositoryHandler
	Business   *BusinessHandler
	Project    *ProjectHandler
	Financial  *FinancialHandler
	Social     *SocialHandler
	IoT        *IoTHandler
	Cost       *CostHandler
	Dashboard  *DashboardHandler
	Audit      *AuditHandler
	Settings   *SettingsHandler
	Webhook    *WebhookHandler
	WebSocket  *WebSocketHandler
}

// NewHandlers creates all handler instances
func NewHandlers(svc *services.Services, log *logger.Logger) *Handlers {
	return &Handlers{
		Health:     NewHealthHandler(svc, log),
		Auth:       NewAuthHandler(svc.Auth, log),
		User:       NewUserHandler(svc.User, log),
		Tenant:     NewTenantHandler(svc.Tenant, log),
		APIKey:     NewAPIKeyHandler(svc.APIKey, log),
		Agent:      NewAgentHandler(svc.Agent, log),
		Execute:    NewExecuteHandler(svc.Execute, log),
		Knowledge:  NewKnowledgeHandler(svc.Knowledge, log),
		Repository: NewRepositoryHandler(svc.Repository, log),
		Business:   NewBusinessHandler(svc.Business, log),
		Project:    NewProjectHandler(svc.Project, log),
		Financial:  NewFinancialHandler(svc.Financial, log),
		Social:     NewSocialHandler(svc.Social, log),
		IoT:        NewIoTHandler(svc.IoT, log),
		Cost:       NewCostHandler(svc.Cost, log),
		Dashboard:  NewDashboardHandler(svc.Dashboard, log),
		Audit:      NewAuditHandler(svc.Audit, log),
		Settings:   NewSettingsHandler(svc.Settings, log),
		Webhook:    NewWebhookHandler(svc.Webhook, log),
		WebSocket:  NewWebSocketHandler(svc.WebSocket, log),
	}
}

// Response helpers

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError sends an error response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// decodeJSON decodes JSON request body
func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

