package providers

import (
	"context"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
)

// OpenAIProvider implements the Provider interface for OpenAI
type OpenAIProvider struct {
	client *openai.Client
	models []ModelInfo
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider(apiKey string) *OpenAIProvider {
	client := openai.NewClient(apiKey)
	
	return &OpenAIProvider{
		client: client,
		models: []ModelInfo{
			{
				ID: "gpt-4o", Name: "GPT-4o", ContextWindow: 128000, MaxOutput: 16384,
				InputPrice: 0.005, OutputPrice: 0.015,
				Capabilities: []string{"text", "vision", "function_calling"},
			},
			{
				ID: "gpt-4o-mini", Name: "GPT-4o Mini", ContextWindow: 128000, MaxOutput: 16384,
				InputPrice: 0.00015, OutputPrice: 0.0006,
				Capabilities: []string{"text", "vision", "function_calling"},
			},
			{
				ID: "gpt-4-turbo", Name: "GPT-4 Turbo", ContextWindow: 128000, MaxOutput: 4096,
				InputPrice: 0.01, OutputPrice: 0.03,
				Capabilities: []string{"text", "vision", "function_calling"},
			},
			{
				ID: "o1", Name: "o1", ContextWindow: 200000, MaxOutput: 100000,
				InputPrice: 0.015, OutputPrice: 0.06,
				Capabilities: []string{"text", "reasoning"},
			},
			{
				ID: "o1-mini", Name: "o1 Mini", ContextWindow: 128000, MaxOutput: 65536,
				InputPrice: 0.003, OutputPrice: 0.012,
				Capabilities: []string{"text", "reasoning"},
			},
		},
	}
}

// Name returns the provider name
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// Complete sends a completion request
func (p *OpenAIProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	chatReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: float32(req.Temperature),
		TopP:        float32(req.TopP),
		Stop:        req.Stop,
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		chatReq.Tools = make([]openai.Tool, len(req.Tools))
		for i, tool := range req.Tools {
			chatReq.Tools[i] = openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        tool.Function.Name,
					Description: tool.Function.Description,
					Parameters:  tool.Function.Parameters,
				},
			}
		}
	}

	resp, err := p.client.CreateChatCompletion(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("openai completion failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	choice := resp.Choices[0]
	
	// Convert tool calls
	var toolCalls []ToolCall
	if len(choice.Message.ToolCalls) > 0 {
		toolCalls = make([]ToolCall, len(choice.Message.ToolCalls))
		for i, tc := range choice.Message.ToolCalls {
			toolCalls[i] = ToolCall{
				ID:   tc.ID,
				Type: string(tc.Type),
				Function: FunctionCall{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			}
		}
	}

	return &CompletionResponse{
		ID:    resp.ID,
		Model: resp.Model,
		Message: Message{
			Role:      choice.Message.Role,
			Content:   choice.Message.Content,
			ToolCalls: toolCalls,
		},
		FinishReason: string(choice.FinishReason),
		Usage: TokenUsage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

// Stream sends a streaming completion request
func (p *OpenAIProvider) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	messages := make([]openai.ChatCompletionMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	chatReq := openai.ChatCompletionRequest{
		Model:       req.Model,
		Messages:    messages,
		MaxTokens:   req.MaxTokens,
		Temperature: float32(req.Temperature),
		TopP:        float32(req.TopP),
		Stop:        req.Stop,
		Stream:      true,
	}

	stream, err := p.client.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return nil, fmt.Errorf("openai stream failed: %w", err)
	}

	chunks := make(chan StreamChunk)
	go func() {
		defer close(chunks)
		defer stream.Close()

		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				chunks <- StreamChunk{Error: err}
				return
			}

			if len(resp.Choices) > 0 {
				chunks <- StreamChunk{
					ID:           resp.ID,
					Delta:        resp.Choices[0].Delta.Content,
					FinishReason: string(resp.Choices[0].FinishReason),
				}
			}
		}
	}()

	return chunks, nil
}

// CountTokens estimates token count
func (p *OpenAIProvider) CountTokens(text string) (int, error) {
	// Approximate: ~4 chars per token for English
	return len(text) / 4, nil
}

// GetModels returns available models
func (p *OpenAIProvider) GetModels() []ModelInfo {
	return p.models
}

// ValidateAPIKey validates the API key
func (p *OpenAIProvider) ValidateAPIKey(ctx context.Context, key string) error {
	client := openai.NewClient(key)
	_, err := client.ListModels(ctx)
	if err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}
	return nil
}

