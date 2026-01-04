package services

import (
	"github.com/delphi-platform/delphi/backend/internal/config"
	"github.com/delphi-platform/delphi/backend/internal/repository"
	"github.com/delphi-platform/delphi/backend/pkg/auth"
	"github.com/delphi-platform/delphi/backend/pkg/crypto"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
)

// Services contains all service instances
type Services struct {
	Auth       *AuthService
	Tenant     *TenantService
	User       *UserService
	APIKey     *APIKeyService
	Agent      *AgentService
	Execute    *ExecuteService
	Knowledge  *KnowledgeService
	Repository *RepositoryService
	Business   *BusinessService
	Project    *ProjectService
	Financial  *FinancialService
	Social     *SocialService
	IoT        *IoTService
	Cost       *CostService
	Dashboard  *DashboardService
	Audit      *AuditService
	Settings   *SettingsService
	Webhook    *WebhookService
	WebSocket  *WebSocketService
}

// NewServices creates all service instances
func NewServices(cfg *config.Config, repos *repository.Repositories, redis *repository.RedisClient, log *logger.Logger) *Services {
	// Initialize encryptor for secrets
	var encryptor *crypto.Encryptor
	if cfg.EncryptionKey != "" {
		var err error
		encryptor, err = crypto.NewEncryptor(cfg.EncryptionKey)
		if err != nil {
			log.Warnw("failed to create encryptor, using nil", "error", err)
		}
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.SupabaseServiceRoleKey, 60, 7) // 60 min access, 7 day refresh

	return &Services{
		Auth:       NewAuthService(cfg, repos, jwtManager, log),
		Tenant:     NewTenantService(repos, log),
		User:       NewUserService(repos, log),
		APIKey:     NewAPIKeyService(repos, encryptor, log),
		Agent:      NewAgentService(cfg, repos, redis, log),
		Execute:    NewExecuteService(cfg, repos, redis, log),
		Knowledge:  NewKnowledgeService(repos, log),
		Repository: NewRepositoryService(cfg, repos, log),
		Business:   NewBusinessService(repos, log),
		Project:    NewProjectService(repos, log),
		Financial:  NewFinancialService(repos, log),
		Social:     NewSocialService(cfg, repos, log),
		IoT:        NewIoTService(repos, encryptor, log),
		Cost:       NewCostService(repos, redis, log),
		Dashboard:  NewDashboardService(repos, redis, log),
		Audit:      NewAuditService(repos, log),
		Settings:   NewSettingsService(repos, log),
		Webhook:    NewWebhookService(cfg, repos, log),
		WebSocket:  NewWebSocketService(redis, log),
	}
}

