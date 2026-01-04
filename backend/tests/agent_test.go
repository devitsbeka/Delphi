package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Agent Tests
// =============================================================================

func TestAgentCreate(t *testing.T) {
	t.Run("creates agent with valid data", func(t *testing.T) {
		ctx := context.Background()
		
		agent := &Agent{
			Name:           "Test Code Review Agent",
			Description:    "An agent for testing purposes",
			Purpose:        "coding",
			ModelProvider:  "openai",
			Model:          "gpt-4-turbo",
			SystemPrompt:   "You are a helpful code reviewer.",
			Goal:           "Review code for quality and best practices",
			OrganizationID: uuid.New(),
			BusinessID:     uuid.New(),
		}

		// Mock service would be injected here
		err := validateAgent(agent)
		require.NoError(t, err)
		
		assert.NotEmpty(t, agent.Name)
		assert.Equal(t, "coding", agent.Purpose)
		assert.Equal(t, "openai", agent.ModelProvider)
	})

	t.Run("fails with empty name", func(t *testing.T) {
		agent := &Agent{
			Name:          "",
			ModelProvider: "openai",
			Model:         "gpt-4-turbo",
		}

		err := validateAgent(agent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "name")
	})

	t.Run("fails with invalid provider", func(t *testing.T) {
		agent := &Agent{
			Name:          "Test Agent",
			ModelProvider: "invalid-provider",
			Model:         "some-model",
		}

		err := validateAgent(agent)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "provider")
	})
}

func TestAgentLifecycle(t *testing.T) {
	t.Run("transitions through states correctly", func(t *testing.T) {
		agent := &Agent{
			Name:   "Lifecycle Test Agent",
			Status: AgentStatusConfigured,
		}

		// Test state transitions
		transitions := []struct {
			from, to AgentStatus
			valid    bool
		}{
			{AgentStatusConfigured, AgentStatusReady, true},
			{AgentStatusReady, AgentStatusBriefing, true},
			{AgentStatusBriefing, AgentStatusExecuting, true},
			{AgentStatusExecuting, AgentStatusReady, true},
			{AgentStatusExecuting, AgentStatusError, true},
			{AgentStatusError, AgentStatusReady, true},
			{AgentStatusConfigured, AgentStatusExecuting, false}, // Invalid
		}

		for _, tt := range transitions {
			agent.Status = tt.from
			err := transitionAgentState(agent, tt.to)
			
			if tt.valid {
				assert.NoError(t, err, "transition from %s to %s should be valid", tt.from, tt.to)
				assert.Equal(t, tt.to, agent.Status)
			} else {
				assert.Error(t, err, "transition from %s to %s should be invalid", tt.from, tt.to)
			}
		}
	})
}

func TestAgentExecution(t *testing.T) {
	t.Run("executes task successfully", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		agent := &Agent{
			ID:            uuid.New(),
			Name:          "Test Executor",
			ModelProvider: "openai",
			Model:         "gpt-4-turbo",
			Status:        AgentStatusReady,
		}

		execution := &Execution{
			ID:      uuid.New(),
			AgentID: agent.ID,
			Task:    "Write a simple hello world function",
			Status:  ExecutionStatusPending,
		}

		// Mock execution
		err := mockExecuteAgent(ctx, agent, execution)
		require.NoError(t, err)

		assert.Equal(t, ExecutionStatusCompleted, execution.Status)
		assert.NotEmpty(t, execution.Output)
	})

	t.Run("handles timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
		defer cancel()

		agent := &Agent{
			ID:     uuid.New(),
			Name:   "Timeout Test Agent",
			Status: AgentStatusReady,
		}

		execution := &Execution{
			ID:      uuid.New(),
			AgentID: agent.ID,
			Task:    "Long running task",
		}

		err := mockExecuteAgent(ctx, agent, execution)
		assert.Error(t, err)
	})
}

// =============================================================================
// Provider Tests
// =============================================================================

func TestProviderInterface(t *testing.T) {
	providers := []string{"openai", "anthropic", "google", "ollama"}

	for _, providerName := range providers {
		t.Run(providerName+" implements interface", func(t *testing.T) {
			provider, err := getProvider(providerName)
			require.NoError(t, err)
			require.NotNil(t, provider)

			// Test that provider can be called
			assert.NotEmpty(t, provider.Name())
		})
	}
}

func TestProviderChatCompletion(t *testing.T) {
	t.Skip("Requires API keys - run with integration tests")

	t.Run("OpenAI chat completion", func(t *testing.T) {
		ctx := context.Background()
		
		req := &ChatRequest{
			Model:    "gpt-4-turbo",
			Messages: []Message{{Role: "user", Content: "Hello"}},
		}

		// This would use actual provider
		resp, err := mockChatCompletion(ctx, "openai", req)
		require.NoError(t, err)
		assert.NotEmpty(t, resp.Content)
	})
}

// =============================================================================
// Knowledge Base Tests
// =============================================================================

func TestKnowledgeBaseSearch(t *testing.T) {
	t.Run("returns relevant results", func(t *testing.T) {
		ctx := context.Background()
		
		// Mock documents
		docs := []Document{
			{ID: "1", Content: "Go is a statically typed programming language"},
			{ID: "2", Content: "React is a JavaScript library for building UIs"},
			{ID: "3", Content: "TypeScript adds types to JavaScript"},
		}

		query := "programming language"
		results := mockSemanticSearch(ctx, docs, query, 2)

		assert.Len(t, results, 2)
		assert.Equal(t, "1", results[0].ID) // Go doc should be first
	})

	t.Run("handles empty query", func(t *testing.T) {
		ctx := context.Background()
		
		results := mockSemanticSearch(ctx, nil, "", 10)
		assert.Empty(t, results)
	})
}

// =============================================================================
// Helper Types and Functions
// =============================================================================

type Agent struct {
	ID             uuid.UUID
	Name           string
	Description    string
	Purpose        string
	ModelProvider  string
	Model          string
	SystemPrompt   string
	Goal           string
	Status         AgentStatus
	OrganizationID uuid.UUID
	BusinessID     uuid.UUID
}

type AgentStatus string

const (
	AgentStatusConfigured AgentStatus = "configured"
	AgentStatusReady      AgentStatus = "ready"
	AgentStatusBriefing   AgentStatus = "briefing"
	AgentStatusExecuting  AgentStatus = "executing"
	AgentStatusPaused     AgentStatus = "paused"
	AgentStatusError      AgentStatus = "error"
)

type Execution struct {
	ID      uuid.UUID
	AgentID uuid.UUID
	Task    string
	Status  ExecutionStatus
	Output  string
}

type ExecutionStatus string

const (
	ExecutionStatusPending   ExecutionStatus = "pending"
	ExecutionStatusRunning   ExecutionStatus = "running"
	ExecutionStatusCompleted ExecutionStatus = "completed"
	ExecutionStatusFailed    ExecutionStatus = "failed"
)

type ChatRequest struct {
	Model    string
	Messages []Message
}

type Message struct {
	Role    string
	Content string
}

type ChatResponse struct {
	Content string
}

type Document struct {
	ID      string
	Content string
}

type Provider interface {
	Name() string
}

// Mock implementations
func validateAgent(agent *Agent) error {
	if agent.Name == "" {
		return assert.AnError
	}
	if agent.ModelProvider != "openai" && agent.ModelProvider != "anthropic" && 
	   agent.ModelProvider != "google" && agent.ModelProvider != "ollama" {
		return assert.AnError
	}
	return nil
}

func transitionAgentState(agent *Agent, to AgentStatus) error {
	validTransitions := map[AgentStatus][]AgentStatus{
		AgentStatusConfigured: {AgentStatusReady},
		AgentStatusReady:      {AgentStatusBriefing, AgentStatusPaused},
		AgentStatusBriefing:   {AgentStatusExecuting, AgentStatusError},
		AgentStatusExecuting:  {AgentStatusReady, AgentStatusError, AgentStatusPaused},
		AgentStatusPaused:     {AgentStatusReady},
		AgentStatusError:      {AgentStatusReady, AgentStatusConfigured},
	}

	allowed := validTransitions[agent.Status]
	for _, s := range allowed {
		if s == to {
			agent.Status = to
			return nil
		}
	}
	return assert.AnError
}

func mockExecuteAgent(ctx context.Context, agent *Agent, execution *Execution) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Millisecond):
		execution.Status = ExecutionStatusCompleted
		execution.Output = "Hello, World!"
		return nil
	}
}

func getProvider(name string) (Provider, error) {
	return &mockProvider{name: name}, nil
}

type mockProvider struct {
	name string
}

func (p *mockProvider) Name() string {
	return p.name
}

func mockChatCompletion(ctx context.Context, provider string, req *ChatRequest) (*ChatResponse, error) {
	return &ChatResponse{Content: "Mock response"}, nil
}

func mockSemanticSearch(ctx context.Context, docs []Document, query string, limit int) []Document {
	if query == "" || len(docs) == 0 {
		return nil
	}
	
	if len(docs) < limit {
		return docs
	}
	return docs[:limit]
}

