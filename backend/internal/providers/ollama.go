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

// OllamaProvider implements the Provider interface for local Ollama
type OllamaProvider struct {
	baseURL    string
	httpClient *http.Client
	models     []ModelInfo
}

// NewOllamaProvider creates a new Ollama provider
func NewOllamaProvider(baseURL string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	
	return &OllamaProvider{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Minute, // Local models can be slow
		},
		models: []ModelInfo{}, // Will be populated dynamically
	}
}

// Name returns the provider name
func (p *OllamaProvider) Name() string {
	return "ollama"
}

// ollamaRequest represents the Ollama API request format
type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  *ollamaOptions  `json:"options,omitempty"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaOptions struct {
	Temperature float64 `json:"temperature,omitempty"`
	TopP        float64 `json:"top_p,omitempty"`
	NumPredict  int     `json:"num_predict,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// ollamaResponse represents the Ollama API response format
type ollamaResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Message   ollamaMessage `json:"message"`
	Done      bool          `json:"done"`
	TotalDuration    int64  `json:"total_duration"`
	LoadDuration     int64  `json:"load_duration"`
	PromptEvalCount  int    `json:"prompt_eval_count"`
	EvalCount        int    `json:"eval_count"`
}

// ollamaModelsResponse represents the response from listing models
type ollamaModelsResponse struct {
	Models []struct {
		Name       string `json:"name"`
		Size       int64  `json:"size"`
		ModifiedAt string `json:"modified_at"`
	} `json:"models"`
}

// Complete sends a completion request
func (p *OllamaProvider) Complete(ctx context.Context, req *CompletionRequest) (*CompletionResponse, error) {
	messages := make([]ollamaMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	ollamaReq := ollamaRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   false,
		Options: &ollamaOptions{
			Temperature: req.Temperature,
			TopP:        req.TopP,
			NumPredict:  req.MaxTokens,
			Stop:        req.Stop,
		},
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", p.baseURL)
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
		return nil, fmt.Errorf("ollama API error: %d - %s", resp.StatusCode, string(bodyBytes))
	}

	var ollamaResp ollamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &CompletionResponse{
		ID:    fmt.Sprintf("ollama-%d", time.Now().UnixNano()),
		Model: ollamaResp.Model,
		Message: Message{
			Role:    ollamaResp.Message.Role,
			Content: ollamaResp.Message.Content,
		},
		FinishReason: "stop",
		Usage: TokenUsage{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
		CreatedAt: time.Now(),
	}, nil
}

// Stream sends a streaming completion request
func (p *OllamaProvider) Stream(ctx context.Context, req *CompletionRequest) (<-chan StreamChunk, error) {
	messages := make([]ollamaMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = ollamaMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	ollamaReq := ollamaRequest{
		Model:    req.Model,
		Messages: messages,
		Stream:   true,
		Options: &ollamaOptions{
			Temperature: req.Temperature,
			TopP:        req.TopP,
			NumPredict:  req.MaxTokens,
			Stop:        req.Stop,
		},
	}

	body, err := json.Marshal(ollamaReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/api/chat", p.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("ollama API error: %d", resp.StatusCode)
	}

	chunks := make(chan StreamChunk)
	go func() {
		defer close(chunks)
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		for {
			var streamResp ollamaResponse
			if err := decoder.Decode(&streamResp); err != nil {
				if err != io.EOF {
					chunks <- StreamChunk{Error: err}
				}
				return
			}

			chunks <- StreamChunk{
				ID:    fmt.Sprintf("ollama-%d", time.Now().UnixNano()),
				Delta: streamResp.Message.Content,
			}

			if streamResp.Done {
				chunks <- StreamChunk{
					FinishReason: "stop",
					Usage: &TokenUsage{
						PromptTokens:     streamResp.PromptEvalCount,
						CompletionTokens: streamResp.EvalCount,
						TotalTokens:      streamResp.PromptEvalCount + streamResp.EvalCount,
					},
				}
				return
			}
		}
	}()

	return chunks, nil
}

// CountTokens estimates token count
func (p *OllamaProvider) CountTokens(text string) (int, error) {
	// Approximate: ~4 chars per token
	return len(text) / 4, nil
}

// GetModels returns available models from Ollama
func (p *OllamaProvider) GetModels() []ModelInfo {
	// If we have cached models, return them
	if len(p.models) > 0 {
		return p.models
	}

	// Try to fetch from Ollama
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/api/tags", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var modelsResp ollamaModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil
	}

	models := make([]ModelInfo, len(modelsResp.Models))
	for i, m := range modelsResp.Models {
		models[i] = ModelInfo{
			ID:           m.Name,
			Name:         m.Name,
			Description:  "Local Ollama model",
			ContextWindow: 4096, // Default, varies by model
			MaxOutput:    2048,
			InputPrice:   0, // Local = free
			OutputPrice:  0,
			Capabilities: []string{"text"},
		}
	}

	p.models = models
	return models
}

// ValidateAPIKey validates the API key (for Ollama, just checks connectivity)
func (p *OllamaProvider) ValidateAPIKey(ctx context.Context, key string) error {
	// Ollama doesn't use API keys, just check if server is reachable
	url := fmt.Sprintf("%s/api/tags", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ollama server not reachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama server error: %d", resp.StatusCode)
	}

	return nil
}

