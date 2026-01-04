package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/models"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
)

const (
	flyAPIBaseURL = "https://api.machines.dev/v1"
)

// FlyMachineManager manages Fly.io Machines for agent execution
type FlyMachineManager struct {
	apiToken   string
	org        string
	appName    string
	region     string
	httpClient *http.Client
	log        *logger.Logger
}

// MachineConfig represents the configuration for a Fly Machine
type MachineConfig struct {
	Image    string            `json:"image"`
	Env      map[string]string `json:"env"`
	Guest    GuestConfig       `json:"guest"`
	Services []ServiceConfig   `json:"services,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// GuestConfig represents the VM resources
type GuestConfig struct {
	CPUKind  string `json:"cpu_kind"`  // shared, performance
	CPUs     int    `json:"cpus"`
	MemoryMB int    `json:"memory_mb"`
}

// ServiceConfig represents a service configuration
type ServiceConfig struct {
	Ports    []PortConfig `json:"ports"`
	Protocol string       `json:"protocol"`
}

// PortConfig represents a port configuration
type PortConfig struct {
	Port     int      `json:"port"`
	Handlers []string `json:"handlers"`
}

// Machine represents a Fly.io Machine
type Machine struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	State      string        `json:"state"`
	Region     string        `json:"region"`
	InstanceID string        `json:"instance_id"`
	Config     MachineConfig `json:"config"`
	CreatedAt  string        `json:"created_at"`
	UpdatedAt  string        `json:"updated_at"`
}

// CreateMachineRequest represents a request to create a machine
type CreateMachineRequest struct {
	Name   string        `json:"name"`
	Region string        `json:"region"`
	Config MachineConfig `json:"config"`
}

// NewFlyMachineManager creates a new Fly Machine manager
func NewFlyMachineManager(apiToken, org, appName, region string, log *logger.Logger) *FlyMachineManager {
	return &FlyMachineManager{
		apiToken: apiToken,
		org:      org,
		appName:  appName,
		region:   region,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		log: log,
	}
}

// CreateMachine creates a new Fly Machine for agent execution
func (m *FlyMachineManager) CreateMachine(ctx context.Context, agent *models.Agent, run *models.AgentRun, secrets map[string]string) (*Machine, error) {
	machineName := fmt.Sprintf("delphi-agent-%s-%s", agent.ID.String()[:8], run.ID.String()[:8])

	// Prepare environment variables
	env := map[string]string{
		"AGENT_ID":    agent.ID.String(),
		"RUN_ID":      run.ID.String(),
		"TENANT_ID":   agent.TenantID.String(),
		"AGENT_TYPE":  string(agent.Type),
		"AGENT_MODEL": agent.Model,
	}

	// Merge secrets
	for k, v := range secrets {
		env[k] = v
	}

	// Determine resources based on agent type
	guest := m.getGuestConfig(agent)

	req := CreateMachineRequest{
		Name:   machineName,
		Region: m.region,
		Config: MachineConfig{
			Image: m.getAgentImage(agent.Type),
			Env:   env,
			Guest: guest,
			Metadata: map[string]string{
				"agent_id":  agent.ID.String(),
				"run_id":    run.ID.String(),
				"tenant_id": agent.TenantID.String(),
			},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/apps/%s/machines", flyAPIBaseURL, m.appName)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+m.apiToken)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := m.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fly API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var machine Machine
	if err := json.NewDecoder(resp.Body).Decode(&machine); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	m.log.Infow("machine created", "machine_id", machine.ID, "name", machine.Name, "region", machine.Region)

	return &machine, nil
}

// WaitForMachine waits for a machine to reach a desired state
func (m *FlyMachineManager) WaitForMachine(ctx context.Context, machineID string, desiredState string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		machine, err := m.GetMachine(ctx, machineID)
		if err != nil {
			return err
		}

		if machine.State == desiredState {
			return nil
		}

		if machine.State == "destroyed" || machine.State == "failed" {
			return fmt.Errorf("machine entered unexpected state: %s", machine.State)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(2 * time.Second):
			continue
		}
	}

	return fmt.Errorf("timeout waiting for machine state: %s", desiredState)
}

// GetMachine retrieves a machine by ID
func (m *FlyMachineManager) GetMachine(ctx context.Context, machineID string) (*Machine, error) {
	url := fmt.Sprintf("%s/apps/%s/machines/%s", flyAPIBaseURL, m.appName, machineID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+m.apiToken)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get machine: %d", resp.StatusCode)
	}

	var machine Machine
	if err := json.NewDecoder(resp.Body).Decode(&machine); err != nil {
		return nil, err
	}

	return &machine, nil
}

// StopMachine stops a running machine
func (m *FlyMachineManager) StopMachine(ctx context.Context, machineID string) error {
	url := fmt.Sprintf("%s/apps/%s/machines/%s/stop", flyAPIBaseURL, m.appName, machineID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+m.apiToken)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to stop machine: %d", resp.StatusCode)
	}

	m.log.Infow("machine stopped", "machine_id", machineID)
	return nil
}

// DestroyMachine destroys a machine
func (m *FlyMachineManager) DestroyMachine(ctx context.Context, machineID string) error {
	url := fmt.Sprintf("%s/apps/%s/machines/%s?force=true", flyAPIBaseURL, m.appName, machineID)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+m.apiToken)

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to destroy machine: %d", resp.StatusCode)
	}

	m.log.Infow("machine destroyed", "machine_id", machineID)
	return nil
}

// getAgentImage returns the container image for an agent type
func (m *FlyMachineManager) getAgentImage(agentType models.AgentType) string {
	// In production, use actual registry images
	baseImage := "registry.fly.io/delphi-agent"

	switch agentType {
	case models.AgentTypeCoding:
		return baseImage + ":coding-latest"
	case models.AgentTypeBusiness, models.AgentTypeMarketing, models.AgentTypeProduct:
		return baseImage + ":general-latest"
	case models.AgentTypeAccounting:
		return baseImage + ":accounting-latest"
	default:
		return baseImage + ":general-latest"
	}
}

// getGuestConfig returns the VM resources for an agent type
func (m *FlyMachineManager) getGuestConfig(agent *models.Agent) GuestConfig {
	// Coding agents need more resources
	if agent.Type == models.AgentTypeCoding {
		return GuestConfig{
			CPUKind:  "shared",
			CPUs:     2,
			MemoryMB: 2048,
		}
	}

	// Default for other agents
	return GuestConfig{
		CPUKind:  "shared",
		CPUs:     1,
		MemoryMB: 1024,
	}
}

// =============================================================================
// Execution Runner
// =============================================================================

// ExecutionRunner orchestrates agent execution
type ExecutionRunner struct {
	machineManager *FlyMachineManager
	briefingEngine *BriefingEngine
	log            *logger.Logger
}

// NewExecutionRunner creates a new execution runner
func NewExecutionRunner(machineManager *FlyMachineManager, briefingEngine *BriefingEngine, log *logger.Logger) *ExecutionRunner {
	return &ExecutionRunner{
		machineManager: machineManager,
		briefingEngine: briefingEngine,
		log:            log,
	}
}

// ExecutionRequest represents an execution request
type ExecutionRequest struct {
	Agent           *models.Agent
	Run             *models.AgentRun
	Prompt          string
	Context         map[string]interface{}
	Secrets         map[string]string
	BriefingContext *BriefingContext
}

// ExecutionResult represents the result of an execution
type ExecutionResult struct {
	RunID        uuid.UUID
	Success      bool
	Response     string
	TokensUsed   int
	Cost         float64
	Duration     time.Duration
	MachineID    string
	Error        string
}

// Execute runs an agent in a sandboxed container
func (r *ExecutionRunner) Execute(ctx context.Context, req *ExecutionRequest) (*ExecutionResult, error) {
	start := time.Now()
	result := &ExecutionResult{
		RunID:   req.Run.ID,
		Success: false,
	}

	// Step 1: Perform briefing
	briefingResult, err := r.briefingEngine.Brief(ctx, req.Agent, req.BriefingContext)
	if err != nil {
		result.Error = fmt.Sprintf("briefing failed: %v", err)
		return result, err
	}

	r.log.Infow("briefing complete", 
		"run_id", req.Run.ID, 
		"estimated_tokens", briefingResult.EstimatedTokens,
	)

	// Step 2: Create machine (in production)
	if r.machineManager != nil && r.machineManager.apiToken != "" {
		machine, err := r.machineManager.CreateMachine(ctx, req.Agent, req.Run, req.Secrets)
		if err != nil {
			result.Error = fmt.Sprintf("failed to create machine: %v", err)
			return result, err
		}
		result.MachineID = machine.ID

		// Wait for machine to be ready
		if err := r.machineManager.WaitForMachine(ctx, machine.ID, "started", 2*time.Minute); err != nil {
			r.machineManager.DestroyMachine(ctx, machine.ID)
			result.Error = fmt.Sprintf("machine failed to start: %v", err)
			return result, err
		}

		// In production, we would:
		// 1. Send the request to the agent container
		// 2. Stream logs back to the client
		// 3. Collect the response
		// 4. Destroy the machine

		defer r.machineManager.DestroyMachine(context.Background(), machine.ID)
	}

	// For now, simulate successful execution
	result.Success = true
	result.Response = "Execution completed successfully"
	result.TokensUsed = briefingResult.EstimatedTokens + 500
	result.Cost = float64(result.TokensUsed) * 0.00001
	result.Duration = time.Since(start)

	return result, nil
}

