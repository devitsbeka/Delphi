package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

func main() {
	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	log := logger.Sugar()

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Get frontend URL for CORS
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "*"
	}

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Compress(5))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{frontendURL, "http://localhost:5173", "https://*.vercel.app"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check endpoints
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"delphi-api","version":"1.0.0"}`))
	})

	r.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ready"}`))
	})

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth endpoints (simplified for demo)
		r.Post("/auth/login", handleLogin)
		r.Post("/auth/register", handleRegister)
		r.Post("/auth/refresh", handleRefresh)

		// Demo data endpoints
		r.Get("/agents", handleListAgents)
		r.Post("/agents", handleCreateAgent)
		r.Get("/agents/{agentID}", handleGetAgent)
		r.Patch("/agents/{agentID}", handleUpdateAgent)
		r.Delete("/agents/{agentID}", handleDeleteAgent)
		r.Post("/agents/{agentID}/launch", handleLaunchAgent)
		r.Post("/agents/{agentID}/pause", handlePauseAgent)
		r.Post("/agents/{agentID}/terminate", handleTerminateAgent)

		r.Get("/dashboard/overview", handleDashboardOverview)
		r.Get("/repositories", handleListRepositories)
		r.Get("/knowledge", handleListKnowledgeBases)
		r.Get("/businesses", handleListBusinesses)
		r.Get("/costs/summary", handleCostsSummary)
	})

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Infof("Starting Delphi API server on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Info("Server stopped")
}

// ============================================================================
// Handlers - Simplified for initial deployment
// ============================================================================

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if str, ok := data.(string); ok {
			w.Write([]byte(str))
		}
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{
		"token": "demo-jwt-token-abc123",
		"user": {
			"id": "user-1",
			"email": "demo@delphi.io",
			"name": "Demo User",
			"tenant_id": "tenant-1",
			"role": "admin",
			"preferences": {},
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		}
	}`)
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusCreated, `{
		"token": "demo-jwt-token-abc123",
		"user": {
			"id": "user-new",
			"email": "newuser@delphi.io",
			"name": "New User",
			"tenant_id": "tenant-1",
			"role": "admin",
			"preferences": {},
			"created_at": "2024-01-01T00:00:00Z",
			"updated_at": "2024-01-01T00:00:00Z"
		}
	}`)
}

func handleRefresh(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{"token": "demo-jwt-token-refreshed"}`)
}

func handleListAgents(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `[
		{
			"id": "agent-1",
			"name": "Code Oracle",
			"description": "Expert code generation and review agent",
			"purpose": "coding",
			"goal": "Automate code generation and review for all repositories",
			"model_provider": "openai",
			"model": "gpt-4o",
			"status": "ready",
			"system_prompt": "You are an expert Go and TypeScript developer.",
			"organization_id": "org-1",
			"created_at": "2024-01-01T12:00:00Z",
			"updated_at": "2024-01-01T12:00:00Z"
		},
		{
			"id": "agent-2",
			"name": "Marketing Guru",
			"description": "Creates engaging marketing content and campaigns",
			"purpose": "content",
			"goal": "Increase brand awareness and social engagement",
			"model_provider": "anthropic",
			"model": "claude-3-opus-20240229",
			"status": "executing",
			"system_prompt": "You are a creative marketing specialist focused on viral content.",
			"organization_id": "org-1",
			"created_at": "2024-01-02T12:00:00Z",
			"updated_at": "2024-01-02T12:00:00Z"
		},
		{
			"id": "agent-3",
			"name": "Financial Analyst",
			"description": "Monitors financial data and generates reports",
			"purpose": "analysis",
			"goal": "Provide accurate financial insights and forecasts",
			"model_provider": "google",
			"model": "gemini-pro",
			"status": "paused",
			"system_prompt": "You are a meticulous financial analyst with expertise in startups.",
			"organization_id": "org-1",
			"created_at": "2024-01-03T12:00:00Z",
			"updated_at": "2024-01-03T12:00:00Z"
		},
		{
			"id": "agent-4",
			"name": "DevOps Engineer",
			"description": "Manages CI/CD pipelines and infrastructure",
			"purpose": "devops",
			"goal": "Ensure 99.9% uptime and fast deployments",
			"model_provider": "openai",
			"model": "gpt-4o",
			"status": "briefing",
			"system_prompt": "You are an expert DevOps engineer with Kubernetes and Terraform expertise.",
			"organization_id": "org-1",
			"created_at": "2024-01-04T12:00:00Z",
			"updated_at": "2024-01-04T12:00:00Z"
		},
		{
			"id": "agent-5",
			"name": "Support Assistant",
			"description": "Handles customer inquiries and provides 24/7 support",
			"purpose": "support",
			"goal": "Maintain 95% customer satisfaction rate",
			"model_provider": "ollama",
			"model": "llama2",
			"status": "ready",
			"system_prompt": "You are a helpful and empathetic customer support agent.",
			"organization_id": "org-1",
			"created_at": "2024-01-05T12:00:00Z",
			"updated_at": "2024-01-05T12:00:00Z"
		}
	]`)
}

func handleCreateAgent(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusCreated, `{
		"id": "agent-new",
		"name": "New Agent",
		"description": "Newly created agent",
		"purpose": "custom",
		"goal": "Custom goal",
		"model_provider": "openai",
		"model": "gpt-4o",
		"status": "configured",
		"system_prompt": "You are a helpful assistant.",
		"organization_id": "org-1",
		"created_at": "2024-01-06T12:00:00Z",
		"updated_at": "2024-01-06T12:00:00Z"
	}`)
}

func handleGetAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	jsonResponse(w, http.StatusOK, fmt.Sprintf(`{
		"id": "%s",
		"name": "Code Oracle",
		"description": "Expert code generation and review agent",
		"purpose": "coding",
		"goal": "Automate code generation and review",
		"model_provider": "openai",
		"model": "gpt-4o",
		"status": "ready",
		"system_prompt": "You are an expert developer.",
		"organization_id": "org-1",
		"created_at": "2024-01-01T12:00:00Z",
		"updated_at": "2024-01-01T12:00:00Z"
	}`, agentID))
}

func handleUpdateAgent(w http.ResponseWriter, r *http.Request) {
	agentID := chi.URLParam(r, "agentID")
	jsonResponse(w, http.StatusOK, fmt.Sprintf(`{
		"id": "%s",
		"name": "Updated Agent",
		"description": "Updated description",
		"purpose": "coding",
		"goal": "Updated goal",
		"model_provider": "openai",
		"model": "gpt-4o",
		"status": "configured",
		"system_prompt": "Updated system prompt.",
		"organization_id": "org-1",
		"created_at": "2024-01-01T12:00:00Z",
		"updated_at": "2024-01-06T12:00:00Z"
	}`, agentID))
}

func handleDeleteAgent(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{"message": "Agent deleted successfully"}`)
}

func handleLaunchAgent(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{"message": "Agent launched successfully", "status": "ready"}`)
}

func handlePauseAgent(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{"message": "Agent paused", "status": "paused"}`)
}

func handleTerminateAgent(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{"message": "Agent terminated", "status": "terminated"}`)
}

func handleDashboardOverview(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{
		"totalAgents": 15,
		"activeAgents": 8,
		"totalSpendYTD": 2847.56,
		"monthlySpend": 487.32,
		"agentStatusDistribution": {
			"ready": 5,
			"executing": 3,
			"paused": 2,
			"error": 1,
			"configured": 3,
			"briefing": 1
		},
		"recentActivities": [
			{"id": "act1", "agentName": "Code Oracle", "type": "commit", "description": "Pushed 3 commits to dev/feature-auth", "timestamp": "2024-01-06T10:00:00Z", "status": "success"},
			{"id": "act2", "agentName": "Marketing Guru", "type": "post", "description": "Published Twitter thread", "timestamp": "2024-01-06T09:30:00Z", "status": "success"},
			{"id": "act3", "agentName": "Financial Analyst", "type": "report", "description": "Generated Q4 financial report", "timestamp": "2024-01-06T09:00:00Z", "status": "success"},
			{"id": "act4", "agentName": "DevOps Engineer", "type": "deploy", "description": "Deployed v2.1.0 to production", "timestamp": "2024-01-06T08:30:00Z", "status": "success"},
			{"id": "act5", "agentName": "Support Assistant", "type": "ticket", "description": "Resolved 12 support tickets", "timestamp": "2024-01-06T08:00:00Z", "status": "success"}
		]
	}`)
}

func handleListRepositories(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `[
		{"id": "repo-1", "name": "delphi-platform", "url": "https://github.com/devitsbeka/Delphi", "status": "connected", "branch_strategy": "gitflow"},
		{"id": "repo-2", "name": "game-studio-apps", "url": "https://github.com/gamestudio/apps", "status": "connected", "branch_strategy": "github-flow"},
		{"id": "repo-3", "name": "saas-backend", "url": "https://github.com/company/saas-backend", "status": "syncing", "branch_strategy": "trunk-based"}
	]`)
}

func handleListKnowledgeBases(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `[
		{"id": "kb-1", "name": "Company Documentation", "type": "document", "document_count": 156},
		{"id": "kb-2", "name": "Codebase Context", "type": "repository", "document_count": 2340},
		{"id": "kb-3", "name": "Product Requirements", "type": "manual", "document_count": 45}
	]`)
}

func handleListBusinesses(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `[
		{"id": "biz-1", "name": "Game Studio", "description": "Mobile game development studio with 30+ apps", "industry": "Gaming"},
		{"id": "biz-2", "name": "Crash Games Inc", "description": "Crash game development company", "industry": "Gaming"},
		{"id": "biz-3", "name": "SaaS Ventures", "description": "Multiple SaaS startup projects", "industry": "Software"}
	]`)
}

func handleCostsSummary(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, `{
		"totalSpendYTD": 2847.56,
		"monthlySpend": 487.32,
		"dailyAverage": 15.73,
		"byProvider": {
			"openai": 1523.45,
			"anthropic": 876.23,
			"google": 312.88,
			"ollama": 135.00
		},
		"byAgent": {
			"Code Oracle": 892.34,
			"Marketing Guru": 567.23,
			"Financial Analyst": 445.67,
			"DevOps Engineer": 534.12,
			"Support Assistant": 408.20
		}
	}`)
}
