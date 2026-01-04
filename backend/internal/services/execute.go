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

// ExecuteService handles agent execution
type ExecuteService struct {
	cfg   *config.Config
	repos *repository.Repositories
	redis *repository.RedisClient
	log   *logger.Logger
}

// NewExecuteService creates a new execute service
func NewExecuteService(cfg *config.Config, repos *repository.Repositories, redis *repository.RedisClient, log *logger.Logger) *ExecuteService {
	return &ExecuteService{
		cfg:   cfg,
		repos: repos,
		redis: redis,
		log:   log,
	}
}

// ExecuteRequest represents an execution request
type ExecuteRequest struct {
	AgentID uuid.UUID `json:"agent_id"`
	Prompt  string    `json:"prompt"`
	Context map[string]interface{} `json:"context,omitempty"`
}

// ExecuteResponse represents execution result
type ExecuteResponse struct {
	RunID     uuid.UUID       `json:"run_id"`
	Status    models.RunStatus `json:"status"`
	Result    json.RawMessage `json:"result,omitempty"`
	Error     string          `json:"error,omitempty"`
}

// Create creates a new execution
func (s *ExecuteService) Create(ctx context.Context, tenantID uuid.UUID, req *ExecuteRequest) (*models.AgentRun, error) {
	// Get agent
	agent, err := s.repos.Agents.GetByID(ctx, req.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	if agent == nil || agent.TenantID != tenantID {
		return nil, fmt.Errorf("agent not found")
	}

	// Check agent is ready
	if agent.Status != models.AgentStatusReady {
		return nil, fmt.Errorf("agent is not ready, current status: %s", agent.Status)
	}

	// Check budget limits
	if agent.Config.BudgetLimit > 0 {
		spent, err := s.repos.Costs.GetTotalByAgent(ctx, agent.ID, time.Now().AddDate(0, -1, 0))
		if err != nil {
			s.log.Warnw("failed to check budget", "agent_id", agent.ID, "error", err)
		} else if spent >= agent.Config.BudgetLimit {
			return nil, fmt.Errorf("agent has exceeded its monthly budget limit")
		}
	}

	// Create run record
	run := &models.AgentRun{
		ID:        uuid.New(),
		AgentID:   agent.ID,
		TenantID:  tenantID,
		Prompt:    req.Prompt,
		Status:    models.RunStatusPending,
		StartedAt: time.Now(),
	}

	if err := s.repos.AgentRuns.Create(ctx, run); err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	// Update agent status to executing
	if err := s.repos.Agents.UpdateStatus(ctx, agent.ID, models.AgentStatusExecuting); err != nil {
		s.log.Warnw("failed to update agent status", "agent_id", agent.ID, "error", err)
	}

	// Start execution asynchronously
	go s.executeRun(context.Background(), agent, run)

	s.log.Infow("execution started", "run_id", run.ID, "agent_id", agent.ID, "tenant_id", tenantID)

	return run, nil
}

// executeRun performs the actual agent execution
func (s *ExecuteService) executeRun(ctx context.Context, agent *models.Agent, run *models.AgentRun) {
	s.log.Infow("executing agent run", "run_id", run.ID, "agent_id", agent.ID)

	// Update status to running
	s.repos.AgentRuns.UpdateStatus(ctx, run.ID, models.RunStatusRunning)

	// In production, this would:
	// 1. Create a Fly.io Machine with the agent container
	// 2. Pass the prompt and context to the container
	// 3. Stream execution logs
	// 4. Collect results and costs
	// 5. Tear down the machine

	// For now, simulate execution
	time.Sleep(time.Duration(agent.Config.TimeoutSeconds/10) * time.Second)

	// Simulate successful completion
	result := json.RawMessage(`{"message": "Task completed successfully", "details": "This is a simulated execution result"}`)
	tokensUsed := 1500
	cost := float64(tokensUsed) * 0.00001 // Simplified cost calculation

	// Record cost
	costRecord := &models.CostRecord{
		ID:           uuid.New(),
		TenantID:     run.TenantID,
		AgentID:      &agent.ID,
		RunID:        &run.ID,
		Provider:     agent.Provider,
		Model:        agent.Model,
		InputTokens:  1000,
		OutputTokens: 500,
		Cost:         cost,
		CreatedAt:    time.Now(),
	}
	if err := s.repos.Costs.RecordCost(ctx, costRecord); err != nil {
		s.log.Warnw("failed to record cost", "run_id", run.ID, "error", err)
	}

	// Complete the run
	if err := s.repos.AgentRuns.Complete(ctx, run.ID, result, tokensUsed, cost); err != nil {
		s.log.Errorw("failed to complete run", "run_id", run.ID, "error", err)
		return
	}

	// Return agent to ready status
	if err := s.repos.Agents.UpdateStatus(ctx, agent.ID, models.AgentStatusReady); err != nil {
		s.log.Warnw("failed to update agent status", "agent_id", agent.ID, "error", err)
	}

	s.log.Infow("execution completed", "run_id", run.ID, "agent_id", agent.ID, "tokens", tokensUsed, "cost", cost)
}

// Get retrieves an execution by ID
func (s *ExecuteService) Get(ctx context.Context, tenantID, runID uuid.UUID) (*models.AgentRun, error) {
	run, err := s.repos.AgentRuns.GetByID(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}
	if run == nil || run.TenantID != tenantID {
		return nil, fmt.Errorf("run not found")
	}
	return run, nil
}

// Cancel cancels a running execution
func (s *ExecuteService) Cancel(ctx context.Context, tenantID, runID uuid.UUID) error {
	run, err := s.Get(ctx, tenantID, runID)
	if err != nil {
		return err
	}

	if run.Status != models.RunStatusPending && run.Status != models.RunStatusRunning && run.Status != models.RunStatusBriefing {
		return fmt.Errorf("run cannot be cancelled in status: %s", run.Status)
	}

	// In production, this would also terminate the Fly.io Machine

	if err := s.repos.AgentRuns.UpdateStatus(ctx, runID, models.RunStatusCancelled); err != nil {
		return fmt.Errorf("failed to cancel run: %w", err)
	}

	// Return agent to ready status
	if err := s.repos.Agents.UpdateStatus(ctx, run.AgentID, models.AgentStatusReady); err != nil {
		s.log.Warnw("failed to update agent status", "agent_id", run.AgentID, "error", err)
	}

	s.log.Infow("execution cancelled", "run_id", runID, "tenant_id", tenantID)

	return nil
}

