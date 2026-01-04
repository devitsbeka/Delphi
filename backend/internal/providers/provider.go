package providers

import (
	"context"
	"fmt"
	"time"
)

// =============================================================================
// Core Provider Interface
// =============================================================================

// Provider is the unified interface for all AI providers
type Provider interface {
	// Name returns the provider identifier
	Name() string

	// Complete sends a completion request and returns the response
	Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error)

	// Stream sends a completion request and streams the response
	Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error)

	// CountTokens estimates token count for the given text
	CountTokens(text string) (int, error)

	// GetModels returns available models for this provider
	GetModels() []ModelInfo

	// ValidateAPIKey checks if the API key is valid
	ValidateAPIKey(ctx context.Context, key string) error
}

// =============================================================================
// Request/Response Types
// =============================================================================

// CompletionRequest represents a completion request
type CompletionRequest struct {
	Model       string            `json:"model"`
	Messages    []Message         `json:"messages"`
	Temperature float64           `json:"temperature,omitempty"`
	MaxTokens   int               `json:"max_tokens,omitempty"`
	TopP        float64           `json:"top_p,omitempty"`
	Stop        []string          `json:"stop,omitempty"`
	Tools       []Tool            `json:"tools,omitempty"`
	ToolChoice  string            `json:"tool_choice,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role       string      `json:"role"` // system, user, assistant, tool
	Content    string      `json:"content"`
	Name       string      `json:"name,omitempty"`
	ToolCallID string      `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall  `json:"tool_calls,omitempty"`
}

// Tool represents a function/tool available to the model
type Tool struct {
	Type     string       `json:"type"` // function
	Function ToolFunction `json:"function"`
}

// ToolFunction represents a function definition
type ToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents a tool call made by the model
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// CompletionResponse represents a completion response
type CompletionResponse struct {
	ID           string     `json:"id"`
	Model        string     `json:"model"`
	Message      Message    `json:"message"`
	FinishReason string     `json:"finish_reason"`
	Usage        TokenUsage `json:"usage"`
	CreatedAt    time.Time  `json:"created_at"`
}

// TokenUsage represents token consumption
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a streaming response chunk
type StreamChunk struct {
	ID           string     `json:"id"`
	Delta        string     `json:"delta"`
	FinishReason string     `json:"finish_reason,omitempty"`
	Usage        *TokenUsage `json:"usage,omitempty"`
	Error        error      `json:"-"`
}

// =============================================================================
// Model Information
// =============================================================================

// ModelInfo contains information about a model
type ModelInfo struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	ContextWindow int    `json:"context_window"`
	MaxOutput    int     `json:"max_output"`
	InputPrice   float64 `json:"input_price"`   // per 1K tokens
	OutputPrice  float64 `json:"output_price"`  // per 1K tokens
	Capabilities []string `json:"capabilities"` // text, vision, function_calling, etc.
}

// =============================================================================
// Provider Registry
// =============================================================================

// Registry manages available providers
type Registry struct {
	providers map[string]Provider
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(provider Provider) {
	r.providers[provider.Name()] = provider
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, error) {
	provider, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

// List returns all registered provider names
func (r *Registry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// =============================================================================
// Cost Calculator
// =============================================================================

// CostCalculator calculates costs for API usage
type CostCalculator struct {
	pricing map[string]ModelInfo
}

// NewCostCalculator creates a new cost calculator
func NewCostCalculator() *CostCalculator {
	return &CostCalculator{
		pricing: make(map[string]ModelInfo),
	}
}

// SetPricing sets the pricing for a model
func (c *CostCalculator) SetPricing(model string, info ModelInfo) {
	c.pricing[model] = info
}

// Calculate computes the cost for a completion
func (c *CostCalculator) Calculate(model string, usage TokenUsage) float64 {
	info, ok := c.pricing[model]
	if !ok {
		// Default pricing if model not found
		return float64(usage.TotalTokens) * 0.00001
	}

	inputCost := float64(usage.PromptTokens) / 1000.0 * info.InputPrice
	outputCost := float64(usage.CompletionTokens) / 1000.0 * info.OutputPrice
	return inputCost + outputCost
}

// =============================================================================
// Default Pricing
// =============================================================================

// DefaultPricing returns default model pricing
func DefaultPricing() map[string]ModelInfo {
	return map[string]ModelInfo{
		// OpenAI
		"gpt-4o": {
			ID: "gpt-4o", Name: "GPT-4o", ContextWindow: 128000, MaxOutput: 16384,
			InputPrice: 0.005, OutputPrice: 0.015,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		"gpt-4o-mini": {
			ID: "gpt-4o-mini", Name: "GPT-4o Mini", ContextWindow: 128000, MaxOutput: 16384,
			InputPrice: 0.00015, OutputPrice: 0.0006,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		"gpt-4-turbo": {
			ID: "gpt-4-turbo", Name: "GPT-4 Turbo", ContextWindow: 128000, MaxOutput: 4096,
			InputPrice: 0.01, OutputPrice: 0.03,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		"o1": {
			ID: "o1", Name: "o1", ContextWindow: 200000, MaxOutput: 100000,
			InputPrice: 0.015, OutputPrice: 0.06,
			Capabilities: []string{"text", "reasoning"},
		},
		"o1-mini": {
			ID: "o1-mini", Name: "o1 Mini", ContextWindow: 128000, MaxOutput: 65536,
			InputPrice: 0.003, OutputPrice: 0.012,
			Capabilities: []string{"text", "reasoning"},
		},

		// Anthropic
		"claude-3-5-sonnet-20241022": {
			ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", ContextWindow: 200000, MaxOutput: 8192,
			InputPrice: 0.003, OutputPrice: 0.015,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		"claude-3-opus-20240229": {
			ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", ContextWindow: 200000, MaxOutput: 4096,
			InputPrice: 0.015, OutputPrice: 0.075,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		"claude-3-haiku-20240307": {
			ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku", ContextWindow: 200000, MaxOutput: 4096,
			InputPrice: 0.00025, OutputPrice: 0.00125,
			Capabilities: []string{"text", "vision", "function_calling"},
		},

		// Google
		"gemini-1.5-pro": {
			ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", ContextWindow: 2000000, MaxOutput: 8192,
			InputPrice: 0.00125, OutputPrice: 0.005,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
		"gemini-1.5-flash": {
			ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", ContextWindow: 1000000, MaxOutput: 8192,
			InputPrice: 0.000075, OutputPrice: 0.0003,
			Capabilities: []string{"text", "vision", "function_calling"},
		},
	}
}

