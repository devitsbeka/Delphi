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

const googleAPIURL = "https://generativelanguage.googleapis.com/v1beta/models"

// GoogleProvider implements the Provider interface for Google Gemini
type GoogleProvider struct {
	apiKey     string
	httpClient *http.Client
	models     []ModelInfo
}

// NewGoogleProvider creates a new Google AI provider
func NewGoogleProvider(apiKey string) *GoogleProvider {
	return &GoogleProvider{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
		models: []ModelInfo{
			{
				ID: "gemini-1.5-pro", Name: "Gemini 1.5 Pro", ContextWindow: 2000000, MaxOutput: 8192,
				InputPrice: 0.00125, OutputPrice: 0.005,
				Capabilities: []string{"text", "vision", "function_calling"},
			},
			{
				ID: "gemini-1.5-flash", Name: "Gemini 1.5 Flash", ContextWindow: 1000000, MaxOutput: 8192,
				InputPrice: 0.000075, OutputPrice: 0.0003,
				Capabilities: []string{"text", "vision", "function_calling"},
			},
			{
				ID: "gemini-1.5-flash-8b", Name: "Gemini 1.5 Flash 8B", ContextWindow: 1000000, MaxOutput: 8192,
				InputPrice: 0.0000375, OutputPrice: 0.00015,
				Capabilities: []string{"text", "vision"},
			},
		},
	}
}

// Name returns the provider name
func (p *GoogleProvider) Name() string {
	return "google"
}

// googleRequest represents the Google AI API request format
type googleRequest struct {
	Contents         []googleContent       `json:"contents"`
	SystemInstruction *googleContent       `json:"systemInstruction,omitempty"`
	GenerationConfig *googleGenerationConfig `json:"generationConfig,omitempty"`
	Tools            []googleTool          `json:"tools,omitempty"`
}

type googleContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []googlePart `json:"parts"`
}

type googlePart struct {
	Text string `json:"text,omitempty"`
}

type googleGenerationConfig struct {
	Temperature     float64  `json:"temperature,omitempty"`
	TopP            float64  `json:"topP,omitempty"`
	MaxOutputTokens int      `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type googleTool struct {
	FunctionDeclarations []googleFunctionDecl `json:"functionDeclarations,omitempty"`
}

type googleFunctionDecl struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// googleResponse represents the Google AI API response format
type googleResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
			Role string `json:"role"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// Complete sends a completion request
func (p *GoogleProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	// Extract system message and convert to Google format
	var systemContent *googleContent
	var contents []googleContent

	for _, msg := range req.Messages {
		if msg.Role == "system" {
			systemContent = &googleContent{
				Parts: []googlePart{{Text: msg.Content}},
			}
		} else {
			role := msg.Role
			if role == "assistant" {
				role = "model"
			}
			contents = append(contents, googleContent{
				Role:  role,
				Parts: []googlePart{{Text: msg.Content}},
			})
		}
	}

	googleReq := googleRequest{
		Contents:          contents,
		SystemInstruction: systemContent,
		GenerationConfig: &googleGenerationConfig{
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			MaxOutputTokens: req.MaxTokens,
			StopSequences:   req.Stop,
		},
	}

	// Add tools if provided
	if len(req.Tools) > 0 {
		funcDecls := make([]googleFunctionDecl, len(req.Tools))
		for i, tool := range req.Tools {
			funcDecls[i] = googleFunctionDecl{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			}
		}
		googleReq.Tools = []googleTool{{FunctionDeclarations: funcDecls}}
	}

	body, err := json.Marshal(googleReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/%s:generateContent?key=%s", googleAPIURL, req.Model, p.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var googleResp googleResponse
	if err := json.NewDecoder(resp.Body).Decode(&googleResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(googleResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	candidate := googleResp.Candidates[0]
	var content string
	for _, part := range candidate.Content.Parts {
		content += part.Text
	}

	return &CompletionResponse{
		ID:    fmt.Sprintf("google-%d", time.Now().UnixNano()),
		Model: req.Model,
		Message: Message{
			Role:    "assistant",
			Content: content,
		},
		FinishReason: candidate.FinishReason,
		Usage: TokenUsage{
			PromptTokens:     googleResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: googleResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      googleResp.UsageMetadata.TotalTokenCount,
		},
		CreatedAt: time.Now(),
	}, nil
}

// Stream sends a streaming completion request
func (p *GoogleProvider) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	// For now, fall back to non-streaming
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
func (p *GoogleProvider) CountTokens(text string) (int, error) {
	// Approximate: ~4 chars per token
	return len(text) / 4, nil
}

// GetModels returns available models
func (p *GoogleProvider) GetModels() []ModelInfo {
	return p.models
}

// ValidateAPIKey validates the API key
func (p *GoogleProvider) ValidateAPIKey(ctx context.Context, key string) error {
	// List models to verify the key
	url := fmt.Sprintf("%s?key=%s", googleAPIURL, key)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid API key: status %d", resp.StatusCode)
	}

	return nil
}

