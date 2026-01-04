package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

// AnthropicProvider implements the Provider interface for Anthropic Claude
type AnthropicProvider struct {
	apiKey     string
	httpClient *http.Client
	models     []ModelInfo
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(apiKey string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		models: []ModelInfo{
			{
				ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet", ContextWindow: 200000, MaxOutput: 8192,
				InputPrice: 0.003, OutputPrice: 0.015,
				Capabilities: []string{"text", "vision", "function_calling"},
			},
			{
				ID: "claude-3-opus-20240229", Name: "Claude 3 Opus", ContextWindow: 200000, MaxOutput: 4096,
				InputPrice: 0.015, OutputPrice: 0.075,
				Capabilities: []string{"text", "vision", "function_calling"},
			},
			{
				ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku", ContextWindow: 200000, MaxOutput: 4096,
				InputPrice: 0.00025, OutputPrice: 0.00125,
				Capabilities: []string{"text", "vision", "function_calling"},
			},
			{
				ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku", ContextWindow: 200000, MaxOutput: 8192,
				InputPrice: 0.001, OutputPrice: 0.005,
				Capabilities: []string{"text", "vision", "function_calling"},
			},
		},
	}
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// anthropicRequest represents the Anthropic API request format
type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	Temperature float64            `json:"temperature,omitempty"`
	TopP        float64            `json:"top_p,omitempty"`
	Tools       []anthropicTool    `json:"tools,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// anthropicResponse represents the Anthropic API response format
type anthropicResponse struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Role       string `json:"role"`
	Content    []struct {
		Type string `json:"type"`
		Text string `json:"text,omitempty"`
	} `json:"content"`
	Model      string `json:"model"`
	StopReason string `json:"stop_reason"`
	Usage      struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
}

// Complete sends a completion request
func (p *AnthropicProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Extract system message
	var systemPrompt string
	var messages []anthropicMessage
	
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemPrompt = msg.Content
		} else {
			messages = append(messages, anthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 4096
	}

	anthropicReq := anthropicRequest{
		Model:       req.Model,
		MaxTokens:   maxTokens,
		Messages:    messages,
		System:      systemPrompt,
		Temperature: req.Temperature,
		TopP:        req.TopP,
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		anthropicReq.Tools = make([]anthropicTool, len(req.Tools))
		for i, tool := range req.Tools {
			anthropicReq.Tools[i] = anthropicTool{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				InputSchema: tool.Function.Parameters,
			}
		}
	}

	body, err := json.Marshal(anthropicReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", anthropicAPIURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text content
	var content string
	for _, c := range anthropicResp.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &CompletionResponse{
		ID:    anthropicResp.ID,
		Model: anthropicResp.Model,
		Message: Message{
			Role:    anthropicResp.Role,
			Content: content,
		},
		FinishReason: anthropicResp.StopReason,
		Usage: TokenUsage{
			PromptTokens:     anthropicResp.Usage.InputTokens,
			CompletionTokens: anthropicResp.Usage.OutputTokens,
			TotalTokens:      anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens,
		},
		CreatedAt: time.Now(),
	}, nil
}

// Stream sends a streaming completion request
func (p *AnthropicProvider) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	// For now, fall back to non-streaming
	// TODO: Implement SSE streaming
	resp, err := p.Complete(ctx, req)
	if err != nil {
		return nil, err
	}

	chunks := make(chan StreamChunk, 1)
	go func() {
		defer close(chunks)
		chunks <- StreamChunk{
			ID:           resp.ID,
			Delta:        resp.Message.Content,
			FinishReason: resp.FinishReason,
			Usage:        &resp.Usage,
		}
	}()

	return chunks, nil
}

// CountTokens estimates token count
func (p *AnthropicProvider) CountTokens(text string) (int, error) {
	// Approximate: ~4 chars per token for English
	return len(text) / 4, nil
}

// GetModels returns available models
func (p *AnthropicProvider) GetModels() []ModelInfo {
	return p.models
}

// ValidateAPIKey validates the API key
func (p *AnthropicProvider) ValidateAPIKey(ctx context.Context, key string) error {
	// Send a minimal request to verify the key
	tempProvider := NewAnthropicProvider(key)
	_, err := tempProvider.Complete(ctx, &CompletionRequest{
		Model: "claude-3-haiku-20240307",
		Messages: []Message{
			{Role: "user", Content: "Hi"},
		},
		MaxTokens: 10,
	})
	if err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}
	return nil
}

