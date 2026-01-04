package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/config"
	"github.com/delphi-platform/delphi/backend/internal/models"
	"github.com/delphi-platform/delphi/backend/internal/repository"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
)

// AgentService handles agent operations
type AgentService struct {
	cfg   *config.Config
	repos *repository.Repositories
	redis *repository.RedisClient
	log   *logger.Logger
}

// NewAgentService creates a new agent service
func NewAgentService(cfg *config.Config, repos *repository.Repositories, redis *repository.RedisClient, log *logger.Logger) *AgentService {
	return &AgentService{
		cfg:   cfg,
		repos: repos,
		redis: redis,
		log:   log,
	}
}

// CreateAgentRequest represents agent creation input
type CreateAgentRequest struct {
	Name           string              `json:"name"`
	Description    string              `json:"description"`
	Type           models.AgentType    `json:"type"`
	Provider       models.AIProvider   `json:"provider"`
	Model          string              `json:"model"`
	SystemPrompt   string              `json:"system_prompt"`
	Tools          json.RawMessage     `json:"tools"`
	KnowledgeBases []uuid.UUID         `json:"knowledge_bases"`
	Config         models.AgentConfig  `json:"config"`
}

// Create creates a new agent
func (s *AgentService) Create(ctx context.Context, tenantID uuid.UUID, req *CreateAgentRequest) (*models.Agent, error) {
	now := time.Now()

	// Set defaults for config if not provided
	if req.Config.Temperature == 0 {
		req.Config.Temperature = 0.7
	}
	if req.Config.MaxTokens == 0 {
		req.Config.MaxTokens = 4096
	}
	if req.Config.TimeoutSeconds == 0 {
		req.Config.TimeoutSeconds = 300
	}
	if req.Config.BriefingDepth == "" {
		req.Config.BriefingDepth = "standard"
	}
	req.Config.BriefingRequired = true // Always require briefing

	agent := &models.Agent{
		ID:             uuid.New(),
		TenantID:       tenantID,
		Name:           req.Name,
		Description:    req.Description,
		Type:           req.Type,
		Provider:       req.Provider,
		Model:          req.Model,
		SystemPrompt:   req.SystemPrompt,
		Tools:          req.Tools,
		KnowledgeBases: req.KnowledgeBases,
		Config:         req.Config,
		Status:         models.AgentStatusConfigured,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repos.Agents.Create(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	s.log.Infow("agent created", "agent_id", agent.ID, "tenant_id", tenantID, "type", agent.Type)

	return agent, nil
}

// Get retrieves an agent by ID
func (s *AgentService) Get(ctx context.Context, tenantID, agentID uuid.UUID) (*models.Agent, error) {
	agent, err := s.repos.Agents.GetByID(ctx, agentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil || agent.TenantID != tenantID {
		return nil, fmt.Errorf("agent not found")
	}
	return agent, nil
}

// List returns all agents for a tenant
func (s *AgentService) List(ctx context.Context, tenantID uuid.UUID) ([]*models.Agent, error) {
	return s.repos.Agents.ListByTenant(ctx, tenantID)
}

// Update updates an agent
func (s *AgentService) Update(ctx context.Context, tenantID, agentID uuid.UUID, updates map[string]interface{}) (*models.Agent, error) {
	agent, err := s.Get(ctx, tenantID, agentID)
	if err != nil {
		return nil, err
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		agent.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		agent.Description = desc
	}
	if prompt, ok := updates["system_prompt"].(string); ok {
		agent.SystemPrompt = prompt
	}
	if model, ok := updates["model"].(string); ok {
		agent.Model = model
	}
	if provider, ok := updates["provider"].(string); ok {
		agent.Provider = models.AIProvider(provider)
	}
	if tools, ok := updates["tools"].(json.RawMessage); ok {
		agent.Tools = tools
	}
	if configData, ok := updates["config"].(map[string]interface{}); ok {
		configJSON, _ := json.Marshal(configData)
		json.Unmarshal(configJSON, &agent.Config)
	}

	agent.UpdatedAt = time.Now()

	if err := s.repos.Agents.Update(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to update agent: %w", err)
	}

	return agent, nil
}

// Delete deletes an agent
func (s *AgentService) Delete(ctx context.Context, tenantID, agentID uuid.UUID) error {
	agent, err := s.Get(ctx, tenantID, agentID)
	if err != nil {
		return err
	}

	// Can't delete running agents
	if agent.Status == models.AgentStatusExecuting {
		return fmt.Errorf("cannot delete running agent")
	}

	return s.repos.Agents.Delete(ctx, agentID)
}

// Launch starts an agent (moves to briefing phase)
func (s *AgentService) Launch(ctx context.Context, tenantID, agentID uuid.UUID) (*models.Agent, error) {
	agent, err := s.Get(ctx, tenantID, agentID)
	if err != nil {
		return nil, err
	}

	if agent.Status != models.AgentStatusConfigured && agent.Status != models.AgentStatusPaused && agent.Status != models.AgentStatusTerminated {
		return nil, fmt.Errorf("agent cannot be launched from status: %s", agent.Status)
	}

	// Move to briefing phase
	if err := s.repos.Agents.UpdateStatus(ctx, agentID, models.AgentStatusBriefing); err != nil {
		return nil, fmt.Errorf("failed to update agent status: %w", err)
	}

	agent.Status = models.AgentStatusBriefing

	// Trigger briefing process asynchronously
	go s.runBriefing(context.Background(), agent)

	s.log.Infow("agent launched", "agent_id", agentID, "tenant_id", tenantID)

	return agent, nil
}

// runBriefing performs the agent briefing/grooming phase
func (s *AgentService) runBriefing(ctx context.Context, agent *models.Agent) {
	s.log.Infow("starting agent briefing", "agent_id", agent.ID)

	// 1. Load tenant-wide context
	// 2. Load project-specific context
	// 3. Load recent activity from knowledge base
	// 4. Generate contextual system prompt
	// 5. Verify agent readiness

	// Simulate briefing time based on depth
	var duration time.Duration
	switch agent.Config.BriefingDepth {
	case "quick":
		duration = 2 * time.Second
	case "full":
		duration = 10 * time.Second
	default: // standard
		duration = 5 * time.Second
	}

	time.Sleep(duration)

	// Update status to ready
	if err := s.repos.Agents.UpdateStatus(ctx, agent.ID, models.AgentStatusReady); err != nil {
		s.log.Errorw("failed to update agent status after briefing", "agent_id", agent.ID, "error", err)
		s.repos.Agents.UpdateStatus(ctx, agent.ID, models.AgentStatusError)
		return
	}

	s.log.Infow("agent briefing complete", "agent_id", agent.ID)
}

// Pause pauses an agent
func (s *AgentService) Pause(ctx context.Context, tenantID, agentID uuid.UUID) (*models.Agent, error) {
	agent, err := s.Get(ctx, tenantID, agentID)
	if err != nil {
		return nil, err
	}

	if agent.Status != models.AgentStatusReady && agent.Status != models.AgentStatusExecuting {
		return nil, fmt.Errorf("agent cannot be paused from status: %s", agent.Status)
	}

	if err := s.repos.Agents.UpdateStatus(ctx, agentID, models.AgentStatusPaused); err != nil {
		return nil, fmt.Errorf("failed to update agent status: %w", err)
	}

	agent.Status = models.AgentStatusPaused
	s.log.Infow("agent paused", "agent_id", agentID, "tenant_id", tenantID)

	return agent, nil
}

// Terminate terminates an agent
func (s *AgentService) Terminate(ctx context.Context, tenantID, agentID uuid.UUID) (*models.Agent, error) {
	agent, err := s.Get(ctx, tenantID, agentID)
	if err != nil {
		return nil, err
	}

	if err := s.repos.Agents.UpdateStatus(ctx, agentID, models.AgentStatusTerminated); err != nil {
		return nil, fmt.Errorf("failed to update agent status: %w", err)
	}

	agent.Status = models.AgentStatusTerminated
	s.log.Infow("agent terminated", "agent_id", agentID, "tenant_id", tenantID)

	return agent, nil
}

// ListRuns returns runs for an agent
func (s *AgentService) ListRuns(ctx context.Context, tenantID, agentID uuid.UUID, limit int) ([]*models.AgentRun, error) {
	// Verify agent belongs to tenant
	_, err := s.Get(ctx, tenantID, agentID)
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 50
	}

	return s.repos.AgentRuns.ListByAgent(ctx, agentID, limit)
}

// GetRun returns a specific run
func (s *AgentService) GetRun(ctx context.Context, tenantID, agentID, runID uuid.UUID) (*models.AgentRun, error) {
	// Verify agent belongs to tenant
	_, err := s.Get(ctx, tenantID, agentID)
	if err != nil {
		return nil, err
	}

	run, err := s.repos.AgentRuns.GetByID(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}
	if run == nil || run.AgentID != agentID {
		return nil, fmt.Errorf("run not found")
	}

	return run, nil
}

// GetTemplates returns available agent templates
func (s *AgentService) GetTemplates(ctx context.Context) ([]*models.AgentTemplate, error) {
	// Return predefined templates
	templates := []*models.AgentTemplate{
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000001"),
			Name:        "Full-Stack Developer",
			Description: "A versatile coding agent for full-stack development tasks",
			Type:        models.AgentTypeCoding,
			Category:    "development",
			SystemPrompt: `You are an expert full-stack developer. You write clean, maintainable code following best practices. 
You are proficient in React, TypeScript, Node.js, Go, Python, and SQL. 
Always commit to the dev branch and create descriptive commit messages.`,
			DefaultConfig: models.AgentConfig{
				Temperature:      0.3,
				MaxTokens:        8192,
				TimeoutSeconds:   600,
				BriefingRequired: true,
				BriefingDepth:    "standard",
			},
			IsPublic: true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000002"),
			Name:        "Business Strategist",
			Description: "Analyzes market trends and provides strategic business insights",
			Type:        models.AgentTypeBusiness,
			Category:    "business",
			SystemPrompt: `You are a senior business strategist with expertise in market analysis, competitive intelligence, and growth strategy.
You provide data-driven insights and actionable recommendations.`,
			DefaultConfig: models.AgentConfig{
				Temperature:      0.7,
				MaxTokens:        4096,
				TimeoutSeconds:   300,
				BriefingRequired: true,
				BriefingDepth:    "full",
			},
			IsPublic: true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000003"),
			Name:        "Financial Analyst",
			Description: "Processes financial data and generates accounting reports",
			Type:        models.AgentTypeAccounting,
			Category:    "finance",
			SystemPrompt: `You are a financial analyst and accountant. You analyze financial data, categorize transactions, 
generate reports, and ensure compliance with accounting standards.`,
			DefaultConfig: models.AgentConfig{
				Temperature:      0.2,
				MaxTokens:        4096,
				TimeoutSeconds:   300,
				BriefingRequired: true,
				BriefingDepth:    "full",
			},
			IsPublic: true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000004"),
			Name:        "Marketing Content Creator",
			Description: "Creates engaging marketing content for various platforms",
			Type:        models.AgentTypeMarketing,
			Category:    "marketing",
			SystemPrompt: `You are a creative marketing specialist. You create compelling content for social media, 
blogs, email campaigns, and advertisements. You understand SEO and audience engagement.`,
			DefaultConfig: models.AgentConfig{
				Temperature:      0.8,
				MaxTokens:        2048,
				TimeoutSeconds:   180,
				BriefingRequired: true,
				BriefingDepth:    "quick",
			},
			IsPublic: true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000005"),
			Name:        "DevOps Engineer",
			Description: "Manages infrastructure, CI/CD, and deployment automation",
			Type:        models.AgentTypeCoding,
			Category:    "development",
			SystemPrompt: `You are a DevOps engineer specializing in infrastructure as code, CI/CD pipelines, 
containerization, and cloud services (AWS, GCP, Fly.io). You prioritize security and reliability.`,
			DefaultConfig: models.AgentConfig{
				Temperature:      0.3,
				MaxTokens:        4096,
				TimeoutSeconds:   300,
				BriefingRequired: true,
				BriefingDepth:    "standard",
			},
			IsPublic: true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000006"),
			Name:        "Product Manager",
			Description: "Helps with product planning, feature prioritization, and user research",
			Type:        models.AgentTypeProduct,
			Category:    "product",
			SystemPrompt: `You are an experienced product manager. You help define product vision, prioritize features, 
analyze user feedback, and create product specifications. You think strategically about product-market fit.`,
			DefaultConfig: models.AgentConfig{
				Temperature:      0.6,
				MaxTokens:        4096,
				TimeoutSeconds:   300,
				BriefingRequired: true,
				BriefingDepth:    "full",
			},
			IsPublic: true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000007"),
			Name:        "Executive Assistant",
			Description: "Manages emails, scheduling, and day-to-day communications",
			Type:        models.AgentTypeAssistant,
			Category:    "assistant",
			SystemPrompt: `You are a professional executive assistant. You help manage communications, 
draft emails, organize schedules, and handle administrative tasks efficiently and professionally.`,
			DefaultConfig: models.AgentConfig{
				Temperature:      0.5,
				MaxTokens:        2048,
				TimeoutSeconds:   180,
				BriefingRequired: true,
				BriefingDepth:    "quick",
			},
			IsPublic: true,
		},
		{
			ID:          uuid.MustParse("00000000-0000-0000-0000-000000000008"),
			Name:        "Security Reviewer",
			Description: "Reviews code for security vulnerabilities and best practices",
			Type:        models.AgentTypeCoding,
			Category:    "development",
			SystemPrompt: `You are a security expert specializing in code review and vulnerability assessment. 
You identify security issues, suggest fixes, and ensure code follows security best practices.`,
			DefaultConfig: models.AgentConfig{
				Temperature:      0.2,
				MaxTokens:        8192,
				TimeoutSeconds:   600,
				BriefingRequired: true,
				BriefingDepth:    "full",
			},
			IsPublic: true,
		},
	}

	return templates, nil
}

