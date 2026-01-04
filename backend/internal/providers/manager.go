package providers

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/models"
	"github.com/google/uuid"
)

// Manager handles provider selection and usage
type Manager struct {
	registry       *Registry
	costCalculator *CostCalculator
	mu             sync.RWMutex
}

// NewManager creates a new provider manager
func NewManager() *Manager {
	m := &Manager{
		registry:       NewRegistry(),
		costCalculator: NewCostCalculator(),
	}

	// Load default pricing
	for model, info := range DefaultPricing() {
		m.costCalculator.SetPricing(model, info)
	}

	return m
}

// RegisterProvider registers a provider with the manager
func (m *Manager) RegisterProvider(provider Provider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registry.Register(provider)

	// Add pricing for this provider's models
	for _, model := range provider.GetModels() {
		m.costCalculator.SetPricing(model.ID, model)
	}
}

// GetProvider returns a provider by name
func (m *Manager) GetProvider(name string) (Provider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.registry.Get(name)
}

// CreateProviderWithKey creates a provider instance with the given API key
func (m *Manager) CreateProviderWithKey(providerName models.AIProvider, apiKey string, baseURL string) (Provider, error) {
	switch providerName {
	case models.ProviderOpenAI:
		return NewOpenAIProvider(apiKey), nil
	case models.ProviderAnthropic:
		return NewAnthropicProvider(apiKey), nil
	case models.ProviderGoogle:
		return NewGoogleProvider(apiKey), nil
	case models.ProviderOllama:
		return NewOllamaProvider(baseURL), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

// Complete sends a completion request using the specified provider
func (m *Manager) Complete(ctx context.Context, provider Provider, req *CompletionRequest) (*CompletionResponse, error) {
	start := time.Now()
	
	resp, err := provider.Complete(ctx, req)
	if err != nil {
		return nil, err
	}

	// Calculate cost
	cost := m.costCalculator.Calculate(req.Model, resp.Usage)
	
	// Add timing and cost metadata
	resp.CreatedAt = start

	return resp, nil
}

// Stream sends a streaming request using the specified provider
func (m *Manager) Stream(ctx context.Context, provider Provider, req *CompletionRequest) (<-chan StreamChunk, error) {
	return provider.Stream(ctx, req)
}

// CalculateCost calculates the cost for a given usage
func (m *Manager) CalculateCost(model string, usage TokenUsage) float64 {
	return m.costCalculator.Calculate(model, usage)
}

// GetModelInfo returns information about a model
func (m *Manager) GetModelInfo(model string) (ModelInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Search across all providers
	for _, name := range m.registry.List() {
		provider, _ := m.registry.Get(name)
		for _, info := range provider.GetModels() {
			if info.ID == model {
				return info, true
			}
		}
	}

	// Check default pricing
	pricing := DefaultPricing()
	if info, ok := pricing[model]; ok {
		return info, true
	}

	return ModelInfo{}, false
}

// ListAllModels returns all available models across providers
func (m *Manager) ListAllModels() []ModelInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var models []ModelInfo
	seen := make(map[string]bool)

	for _, name := range m.registry.List() {
		provider, _ := m.registry.Get(name)
		for _, info := range provider.GetModels() {
			if !seen[info.ID] {
				models = append(models, info)
				seen[info.ID] = true
			}
		}
	}

	return models
}

// =============================================================================
// Completion Request Builder
// =============================================================================

// RequestBuilder helps build completion requests
type RequestBuilder struct {
	req *CompletionRequest
}

// NewRequestBuilder creates a new request builder
func NewRequestBuilder(model string) *RequestBuilder {
	return &RequestBuilder{
		req: &CompletionRequest{
			Model:       model,
			Messages:    []Message{},
			Temperature: 0.7,
			MaxTokens:   4096,
		},
	}
}

// WithSystemPrompt adds a system prompt
func (b *RequestBuilder) WithSystemPrompt(prompt string) *RequestBuilder {
	b.req.Messages = append([]Message{{Role: "system", Content: prompt}}, b.req.Messages...)
	return b
}

// WithUserMessage adds a user message
func (b *RequestBuilder) WithUserMessage(content string) *RequestBuilder {
	b.req.Messages = append(b.req.Messages, Message{Role: "user", Content: content})
	return b
}

// WithAssistantMessage adds an assistant message
func (b *RequestBuilder) WithAssistantMessage(content string) *RequestBuilder {
	b.req.Messages = append(b.req.Messages, Message{Role: "assistant", Content: content})
	return b
}

// WithMessages sets all messages
func (b *RequestBuilder) WithMessages(messages []Message) *RequestBuilder {
	b.req.Messages = messages
	return b
}

// WithTemperature sets the temperature
func (b *RequestBuilder) WithTemperature(temp float64) *RequestBuilder {
	b.req.Temperature = temp
	return b
}

// WithMaxTokens sets the max tokens
func (b *RequestBuilder) WithMaxTokens(tokens int) *RequestBuilder {
	b.req.MaxTokens = tokens
	return b
}

// WithTools adds tools
func (b *RequestBuilder) WithTools(tools []Tool) *RequestBuilder {
	b.req.Tools = tools
	return b
}

// Build returns the completed request
func (b *RequestBuilder) Build() *CompletionRequest {
	return b.req
}

// =============================================================================
// Usage Tracking
// =============================================================================

// UsageRecord represents a usage record for tracking
type UsageRecord struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	AgentID      *uuid.UUID
	RunID        *uuid.UUID
	Provider     string
	Model        string
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Cost         float64
	Duration     time.Duration
	CreatedAt    time.Time
}

// NewUsageRecord creates a usage record from a completion response
func NewUsageRecord(tenantID uuid.UUID, agentID, runID *uuid.UUID, provider, model string, usage TokenUsage, cost float64, duration time.Duration) *UsageRecord {
	return &UsageRecord{
		ID:           uuid.New(),
		TenantID:     tenantID,
		AgentID:      agentID,
		RunID:        runID,
		Provider:     provider,
		Model:        model,
		InputTokens:  usage.PromptTokens,
		OutputTokens: usage.CompletionTokens,
		TotalTokens:  usage.TotalTokens,
		Cost:         cost,
		Duration:     duration,
		CreatedAt:    time.Now(),
	}
}

