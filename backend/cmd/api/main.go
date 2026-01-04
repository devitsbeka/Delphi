package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

// AI Provider interfaces and implementations
type AIProvider interface {
	Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	Name() string
}

// OpenAI Provider
type OpenAIProvider struct {
	apiKey string
	model  string
}

func NewOpenAIProvider(apiKey, model string) *OpenAIProvider {
	if model == "" {
		model = "gpt-4o"
	}
	return &OpenAIProvider{apiKey: apiKey, model: model}
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"max_tokens": 4096,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Choices) > 0 {
		return result.Choices[0].Message.Content, nil
	}
	return "", fmt.Errorf("no response from OpenAI")
}

// Anthropic Provider
type AnthropicProvider struct {
	apiKey string
	model  string
}

func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	if model == "" {
		model = "claude-sonnet-4-20250514"
	}
	return &AnthropicProvider{apiKey: apiKey, model: model}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

func (p *AnthropicProvider) Complete(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 4096,
		"system":     systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": userPrompt},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Anthropic API error: %s", string(body))
	}

	var result struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if len(result.Content) > 0 {
		return result.Content[0].Text, nil
	}
	return "", fmt.Errorf("no response from Anthropic")
}

// Agent store (in-memory for now, would be database in production)
type Agent struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Purpose       string    `json:"purpose"`
	Goal          string    `json:"goal"`
	ModelProvider string    `json:"model_provider"`
	Model         string    `json:"model"`
	Status        string    `json:"status"`
	SystemPrompt  string    `json:"system_prompt"`
	OrgID         string    `json:"organization_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Execution struct {
	ID           string    `json:"id"`
	AgentID      string    `json:"agent_id"`
	AgentName    string    `json:"agent_name"`
	Prompt       string    `json:"prompt"`
	Response     string    `json:"response"`
	Status       string    `json:"status"`
	Provider     string    `json:"provider"`
	Model        string    `json:"model"`
	TokensUsed   int       `json:"tokens_used"`
	CostUSD      float64   `json:"cost_usd"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

var (
	agents     = make(map[string]*Agent)
	executions = make(map[string]*Execution)
	providers  = make(map[string]AIProvider)
	logger     *zap.SugaredLogger
)

func initProviders() {
	openaiKey := os.Getenv("OPENAI_API_KEY")
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")

	if openaiKey != "" {
		providers["openai"] = NewOpenAIProvider(openaiKey, "gpt-4o")
		logger.Info("OpenAI provider initialized")
	}

	if anthropicKey != "" {
		providers["anthropic"] = NewAnthropicProvider(anthropicKey, "claude-sonnet-4-20250514")
		logger.Info("Anthropic provider initialized")
	}

	// Initialize default agents
	defaultAgents := []*Agent{
		{
			ID:            "agent-1",
			Name:          "Code Oracle",
			Description:   "Expert code generation and review agent",
			Purpose:       "coding",
			Goal:          "Automate code generation and review for all repositories",
			ModelProvider: "openai",
			Model:         "gpt-4o",
			Status:        "ready",
			SystemPrompt:  "You are an expert software engineer specializing in Go, TypeScript, React, and Python. You write clean, efficient, well-documented code. Always explain your reasoning and provide complete, working solutions.",
			OrgID:         "org-1",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "agent-2",
			Name:          "Marketing Guru",
			Description:   "Creates engaging marketing content and campaigns",
			Purpose:       "content",
			Goal:          "Increase brand awareness and social engagement",
			ModelProvider: "anthropic",
			Model:         "claude-sonnet-4-20250514",
			Status:        "ready",
			SystemPrompt:  "You are a creative marketing specialist with expertise in viral content, social media strategies, and brand storytelling. Create engaging, memorable content that resonates with audiences.",
			OrgID:         "org-1",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "agent-3",
			Name:          "Financial Analyst",
			Description:   "Monitors financial data and generates reports",
			Purpose:       "analysis",
			Goal:          "Provide accurate financial insights and forecasts",
			ModelProvider: "openai",
			Model:         "gpt-4o",
			Status:        "ready",
			SystemPrompt:  "You are a meticulous financial analyst with expertise in startups, SaaS metrics, and gaming industry economics. Provide data-driven insights and actionable recommendations.",
			OrgID:         "org-1",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "agent-4",
			Name:          "DevOps Engineer",
			Description:   "Manages CI/CD pipelines and infrastructure",
			Purpose:       "devops",
			Goal:          "Ensure 99.9% uptime and fast deployments",
			ModelProvider: "anthropic",
			Model:         "claude-sonnet-4-20250514",
			Status:        "ready",
			SystemPrompt:  "You are an expert DevOps engineer with deep knowledge of Kubernetes, Docker, Terraform, GitHub Actions, and cloud platforms (AWS, GCP, Fly.io). Provide production-ready configurations and best practices.",
			OrgID:         "org-1",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "agent-5",
			Name:          "Product Visionary",
			Description:   "Develops product roadmaps and feature ideas",
			Purpose:       "product",
			Goal:          "Drive product innovation and user growth",
			ModelProvider: "openai",
			Model:         "gpt-4o",
			Status:        "ready",
			SystemPrompt:  "You are a visionary product manager with experience in mobile games, SaaS, and consumer apps. You think strategically about user needs, market trends, and competitive positioning.",
			OrgID:         "org-1",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	for _, agent := range defaultAgents {
		agents[agent.ID] = agent
	}
}

func main() {
	// Initialize logger
	zapLogger, _ := zap.NewProduction()
	defer zapLogger.Sync()
	logger = zapLogger.Sugar()

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Initialize AI providers
	initProviders()

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS configuration - allow all origins for now
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	// Health check endpoints
	r.Get("/health", handleHealth)
	r.Get("/ready", handleReady)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth endpoints
		r.Post("/auth/login", handleLogin)
		r.Post("/auth/register", handleRegister)

		// Agents
		r.Get("/agents", handleListAgents)
		r.Post("/agents", handleCreateAgent)
		r.Get("/agents/{agentID}", handleGetAgent)
		r.Patch("/agents/{agentID}", handleUpdateAgent)
		r.Delete("/agents/{agentID}", handleDeleteAgent)
		r.Post("/agents/{agentID}/launch", handleLaunchAgent)
		r.Post("/agents/{agentID}/pause", handlePauseAgent)
		r.Post("/agents/{agentID}/terminate", handleTerminateAgent)

		// Executions - the main AI interaction endpoint
		r.Post("/execute", handleExecute)
		r.Get("/executions", handleListExecutions)
		r.Get("/executions/{executionID}", handleGetExecution)

		// Dashboard
		r.Get("/dashboard/overview", handleDashboardOverview)

		// Other endpoints
		r.Get("/repositories", handleListRepositories)
		r.Get("/knowledge", handleListKnowledgeBases)
		r.Get("/businesses", handleListBusinesses)
		r.Get("/costs/summary", handleCostsSummary)

		// Provider status
		r.Get("/providers/status", handleProviderStatus)
	})

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 120 * time.Second, // Longer for AI responses
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Infof("Starting Delphi API server on port %s", port)
		logger.Infof("Providers available: %d", len(providers))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server stopped")
}

// ============================================================================
// Handlers
// ============================================================================

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func jsonError(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"service":   "delphi-api",
		"version":   "1.0.0",
		"providers": len(providers),
	})
}

func handleReady(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]string{"status": "ready"})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// For now, accept any credentials (would validate against DB in production)
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"token": "delphi-jwt-token-" + time.Now().Format("20060102150405"),
		"user": map[string]interface{}{
			"id":         "user-1",
			"email":      req.Email,
			"name":       strings.Split(req.Email, "@")[0],
			"tenant_id":  "tenant-1",
			"role":       "admin",
			"created_at": time.Now().Format(time.RFC3339),
		},
	})
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"token": "delphi-jwt-token-" + time.Now().Format("20060102150405"),
		"user": map[string]interface{}{
			"id":         "user-" + time.Now().Format("150405"),
			"email":      req.Email,
			"name":       req.Name,
			"tenant_id":  "tenant-1",
			"role":       "admin",
			"created_at": time.Now().Format(time.RFC3339),
		},
	})
}

func handleListAgents(w http.ResponseWriter, r *http.Request) {
	agentList := make([]*Agent, 0, len(agents))
	for _, agent := range agents {
		agentList = append(agentList, agent)
	}
	jsonResponse(w, http.StatusOK, agentList)
}

func handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	var req Agent
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	req.ID = fmt.Sprintf("agent-%d", time.Now().UnixNano())
	req.Status = "configured"
	req.OrgID = "org-1"
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	agents[req.ID] = &req
	jsonResponse(w, http.StatusCreated, req)
}

func handleGetAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	agent, ok := agents[agentID]
	if !ok {
		jsonError(w, http.StatusNotFound, "Agent not found")
		return
	}
	jsonResponse(w, http.StatusOK, agent)
}

func handleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	agent, ok := agents[agentID]
	if !ok {
		jsonError(w, http.StatusNotFound, "Agent not found")
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Apply updates
	if name, ok := updates["name"].(string); ok {
		agent.Name = name
	}
	if desc, ok := updates["description"].(string); ok {
		agent.Description = desc
	}
	if purpose, ok := updates["purpose"].(string); ok {
		agent.Purpose = purpose
	}
	if goal, ok := updates["goal"].(string); ok {
		agent.Goal = goal
	}
	if provider, ok := updates["model_provider"].(string); ok {
		agent.ModelProvider = provider
	}
	if model, ok := updates["model"].(string); ok {
		agent.Model = model
	}
	if prompt, ok := updates["system_prompt"].(string); ok {
		agent.SystemPrompt = prompt
	}

	agent.UpdatedAt = time.Now()
	jsonResponse(w, http.StatusOK, agent)
}

func handleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	if _, ok := agents[agentID]; !ok {
		jsonError(w, http.StatusNotFound, "Agent not found")
		return
	}
	delete(agents, agentID)
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Agent deleted"})
}

func handleLaunchAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	agent, ok := agents[agentID]
	if !ok {
		jsonError(w, http.StatusNotFound, "Agent not found")
		return
	}
	agent.Status = "ready"
	agent.UpdatedAt = time.Now()
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Agent launched", "status": "ready"})
}

func handlePauseAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	agent, ok := agents[agentID]
	if !ok {
		jsonError(w, http.StatusNotFound, "Agent not found")
		return
	}
	agent.Status = "paused"
	agent.UpdatedAt = time.Now()
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Agent paused", "status": "paused"})
}

func handleTerminateAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	agent, ok := agents[agentID]
	if !ok {
		jsonError(w, http.StatusNotFound, "Agent not found")
		return
	}
	agent.Status = "terminated"
	agent.UpdatedAt = time.Now()
	jsonResponse(w, http.StatusOK, map[string]string{"message": "Agent terminated", "status": "terminated"})
}

// handleExecute - The main AI execution endpoint
func handleExecute(w http.ResponseWriter, r *http.Request) {
	var req struct {
		AgentID string `json:"agent_id"`
		Prompt  string `json:"prompt"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AgentID == "" || req.Prompt == "" {
		jsonError(w, http.StatusBadRequest, "agent_id and prompt are required")
		return
	}

	agent, ok := agents[req.AgentID]
	if !ok {
		jsonError(w, http.StatusNotFound, "Agent not found")
		return
	}

	// Get the appropriate provider
	provider, ok := providers[agent.ModelProvider]
	if !ok {
		jsonError(w, http.StatusBadRequest, fmt.Sprintf("Provider '%s' not configured. Please set %s_API_KEY environment variable.", agent.ModelProvider, strings.ToUpper(agent.ModelProvider)))
		return
	}

	// Create execution record
	execution := &Execution{
		ID:        fmt.Sprintf("exec-%d", time.Now().UnixNano()),
		AgentID:   agent.ID,
		AgentName: agent.Name,
		Prompt:    req.Prompt,
		Status:    "running",
		Provider:  agent.ModelProvider,
		Model:     agent.Model,
		StartTime: time.Now(),
	}
	executions[execution.ID] = execution

	// Update agent status
	agent.Status = "executing"

	// Call AI provider
	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()

	response, err := provider.Complete(ctx, agent.SystemPrompt, req.Prompt)
	execution.EndTime = time.Now()

	if err != nil {
		execution.Status = "failed"
		execution.ErrorMessage = err.Error()
		agent.Status = "error"
		logger.Errorw("AI execution failed", "agent", agent.Name, "error", err)
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("AI execution failed: %v", err))
		return
	}

	execution.Status = "completed"
	execution.Response = response
	execution.TokensUsed = len(response) / 4 // Rough estimate
	execution.CostUSD = float64(execution.TokensUsed) * 0.00003 // Rough estimate

	agent.Status = "ready"

	logger.Infow("AI execution completed", "agent", agent.Name, "tokens", execution.TokensUsed)

	jsonResponse(w, http.StatusOK, execution)
}

func handleListExecutions(w http.ResponseWriter, r *http.Request) {
	execList := make([]*Execution, 0, len(executions))
	for _, exec := range executions {
		execList = append(execList, exec)
	}
	jsonResponse(w, http.StatusOK, execList)
}

func handleGetExecution(w http.ResponseWriter, r *http.Request) {
	execID := chi.URLParam(r, "executionID")
	exec, ok := executions[execID]
	if !ok {
		jsonError(w, http.StatusNotFound, "Execution not found")
		return
	}
	jsonResponse(w, http.StatusOK, exec)
}

func handleProviderStatus(w http.ResponseWriter, r *http.Request) {
	status := make(map[string]interface{})
	for name := range providers {
		status[name] = map[string]interface{}{
			"configured": true,
			"name":       name,
		}
	}

	// Check for unconfigured providers
	if _, ok := providers["openai"]; !ok {
		status["openai"] = map[string]interface{}{
			"configured": false,
			"message":    "OPENAI_API_KEY not set",
		}
	}
	if _, ok := providers["anthropic"]; !ok {
		status["anthropic"] = map[string]interface{}{
			"configured": false,
			"message":    "ANTHROPIC_API_KEY not set",
		}
	}

	jsonResponse(w, http.StatusOK, status)
}

func handleDashboardOverview(w http.ResponseWriter, r *http.Request) {
	// Calculate real stats
	totalAgents := len(agents)
	activeAgents := 0
	for _, agent := range agents {
		if agent.Status == "ready" || agent.Status == "executing" {
			activeAgents++
		}
	}

	totalExecutions := len(executions)
	var totalCost float64
	var totalTokens int
	recentExecutions := make([]*Execution, 0)
	for _, exec := range executions {
		totalCost += exec.CostUSD
		totalTokens += exec.TokensUsed
		if len(recentExecutions) < 10 {
			recentExecutions = append(recentExecutions, exec)
		}
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"totalAgents":        totalAgents,
		"activeAgents":       activeAgents,
		"totalExecutions":    totalExecutions,
		"totalSpendYTD":      totalCost,
		"monthlySpend":       totalCost,
		"totalTokens":        totalTokens,
		"providersConfigured": len(providers),
		"recentExecutions":   recentExecutions,
	})
}

func handleListRepositories(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, []map[string]interface{}{
		{"id": "repo-1", "name": "Delphi", "url": "https://github.com/devitsbeka/Delphi", "status": "connected", "branch_strategy": "gitflow"},
	})
}

func handleListKnowledgeBases(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, []map[string]interface{}{
		{"id": "kb-1", "name": "Company Documentation", "type": "document", "document_count": 0},
	})
}

func handleListBusinesses(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, []map[string]interface{}{
		{"id": "biz-1", "name": "Game Studio", "description": "Mobile game development studio", "industry": "Gaming"},
		{"id": "biz-2", "name": "SaaS Ventures", "description": "Multiple SaaS startup projects", "industry": "Software"},
	})
}

func handleCostsSummary(w http.ResponseWriter, r *http.Request) {
	var totalCost float64
	var totalTokens int
	for _, exec := range executions {
		totalCost += exec.CostUSD
		totalTokens += exec.TokensUsed
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"totalSpendYTD": totalCost,
		"monthlySpend":  totalCost,
		"totalTokens":   totalTokens,
		"byProvider": map[string]float64{
			"openai":    totalCost * 0.6,
			"anthropic": totalCost * 0.4,
		},
	})
}
