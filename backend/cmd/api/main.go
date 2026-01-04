package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/config"
	"github.com/delphi-platform/delphi/backend/internal/handlers"
	"github.com/delphi-platform/delphi/backend/internal/middleware"
	"github.com/delphi-platform/delphi/backend/internal/repository"
	"github.com/delphi-platform/delphi/backend/internal/services"
	"github.com/delphi-platform/delphi/backend/pkg/logger"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

func main() {
	// Initialize logger
	log := logger.New()
	defer log.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
	}

	// Initialize database connection
	db, err := repository.NewPostgresDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database", "error", err)
	}
	defer db.Close()

	// Initialize Redis
	redis, err := repository.NewRedisClient(cfg.RedisURL)
	if err != nil {
		log.Fatal("Failed to connect to Redis", "error", err)
	}
	defer redis.Close()

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize services
	svc := services.NewServices(cfg, repos, redis, log)

	// Initialize handlers
	h := handlers.NewHandlers(svc, log)

	// Setup router
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.Logger(log))
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.Compress(5))

	// CORS configuration
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Rate limiting
	r.Use(httprate.LimitByIP(100, time.Minute))

	// Health check (no auth required)
	r.Get("/health", h.Health.Check)
	r.Get("/ready", h.Health.Ready)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes
		r.Group(func(r chi.Router) {
			r.Post("/auth/login", h.Auth.Login)
			r.Post("/auth/register", h.Auth.Register)
			r.Post("/auth/refresh", h.Auth.RefreshToken)
			r.Post("/auth/forgot-password", h.Auth.ForgotPassword)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate(svc.Auth))
			r.Use(middleware.TenantContext())

			// User profile
			r.Get("/me", h.User.GetProfile)
			r.Patch("/me", h.User.UpdateProfile)
			r.Post("/me/password", h.User.ChangePassword)

			// Tenants
			r.Route("/tenants", func(r chi.Router) {
				r.Get("/", h.Tenant.List)
				r.Post("/", h.Tenant.Create)
				r.Get("/{tenantID}", h.Tenant.Get)
				r.Patch("/{tenantID}", h.Tenant.Update)
			})

			// API Keys (Provider credentials)
			r.Route("/api-keys", func(r chi.Router) {
				r.Get("/", h.APIKey.List)
				r.Post("/", h.APIKey.Create)
				r.Delete("/{keyID}", h.APIKey.Delete)
				r.Post("/{keyID}/validate", h.APIKey.Validate)
			})

			// Agents
			r.Route("/agents", func(r chi.Router) {
				r.Get("/", h.Agent.List)
				r.Post("/", h.Agent.Create)
				r.Get("/templates", h.Agent.ListTemplates)
				r.Get("/{agentID}", h.Agent.Get)
				r.Patch("/{agentID}", h.Agent.Update)
				r.Delete("/{agentID}", h.Agent.Delete)
				r.Post("/{agentID}/launch", h.Agent.Launch)
				r.Post("/{agentID}/pause", h.Agent.Pause)
				r.Post("/{agentID}/terminate", h.Agent.Terminate)
				r.Get("/{agentID}/runs", h.Agent.ListRuns)
				r.Get("/{agentID}/runs/{runID}", h.Agent.GetRun)
				r.Get("/{agentID}/runs/{runID}/logs", h.Agent.GetRunLogs)
			})

			// Agent execution (prompts/tasks)
			r.Route("/execute", func(r chi.Router) {
				r.Post("/", h.Execute.Create)
				r.Get("/{executionID}", h.Execute.Get)
				r.Post("/{executionID}/cancel", h.Execute.Cancel)
			})

			// Knowledge bases
			r.Route("/knowledge", func(r chi.Router) {
				r.Get("/", h.Knowledge.List)
				r.Post("/", h.Knowledge.Create)
				r.Get("/{kbID}", h.Knowledge.Get)
				r.Patch("/{kbID}", h.Knowledge.Update)
				r.Delete("/{kbID}", h.Knowledge.Delete)
				r.Post("/{kbID}/documents", h.Knowledge.UploadDocument)
				r.Get("/{kbID}/documents", h.Knowledge.ListDocuments)
				r.Delete("/{kbID}/documents/{docID}", h.Knowledge.DeleteDocument)
				r.Post("/{kbID}/query", h.Knowledge.Query)
			})

			// Repositories
			r.Route("/repositories", func(r chi.Router) {
				r.Get("/", h.Repository.List)
				r.Post("/", h.Repository.Connect)
				r.Get("/{repoID}", h.Repository.Get)
				r.Delete("/{repoID}", h.Repository.Disconnect)
				r.Post("/{repoID}/sync", h.Repository.Sync)
				r.Get("/{repoID}/branches", h.Repository.ListBranches)
				r.Get("/{repoID}/commits", h.Repository.ListCommits)
				r.Get("/{repoID}/prs", h.Repository.ListPRs)
			})

			// Businesses
			r.Route("/businesses", func(r chi.Router) {
				r.Get("/", h.Business.List)
				r.Post("/", h.Business.Create)
				r.Get("/{businessID}", h.Business.Get)
				r.Patch("/{businessID}", h.Business.Update)
				r.Delete("/{businessID}", h.Business.Delete)

				// Projects under business
				r.Route("/{businessID}/projects", func(r chi.Router) {
					r.Get("/", h.Project.List)
					r.Post("/", h.Project.Create)
					r.Get("/{projectID}", h.Project.Get)
					r.Patch("/{projectID}", h.Project.Update)
					r.Delete("/{projectID}", h.Project.Delete)
				})
			})

			// Financial
			r.Route("/financial", func(r chi.Router) {
				r.Get("/accounts", h.Financial.ListAccounts)
				r.Post("/accounts", h.Financial.CreateAccount)
				r.Get("/transactions", h.Financial.ListTransactions)
				r.Post("/transactions", h.Financial.CreateTransaction)
				r.Get("/reports", h.Financial.GetReports)
				r.Get("/budgets", h.Financial.ListBudgets)
				r.Post("/budgets", h.Financial.CreateBudget)
			})

			// Social media
			r.Route("/social", func(r chi.Router) {
				r.Get("/accounts", h.Social.ListAccounts)
				r.Post("/accounts/connect", h.Social.ConnectAccount)
				r.Delete("/accounts/{accountID}", h.Social.DisconnectAccount)
				r.Get("/posts", h.Social.ListPosts)
				r.Post("/posts", h.Social.CreatePost)
				r.Patch("/posts/{postID}", h.Social.UpdatePost)
				r.Post("/posts/{postID}/schedule", h.Social.SchedulePost)
				r.Post("/posts/{postID}/publish", h.Social.PublishPost)
				r.Get("/analytics", h.Social.GetAnalytics)
			})

			// IoT
			r.Route("/iot", func(r chi.Router) {
				r.Get("/devices", h.IoT.ListDevices)
				r.Post("/devices", h.IoT.RegisterDevice)
				r.Get("/devices/{deviceID}", h.IoT.GetDevice)
				r.Patch("/devices/{deviceID}", h.IoT.UpdateDevice)
				r.Delete("/devices/{deviceID}", h.IoT.DeleteDevice)
				r.Post("/devices/{deviceID}/command", h.IoT.SendCommand)
				r.Get("/devices/{deviceID}/telemetry", h.IoT.GetTelemetry)
			})

			// Cost tracking
			r.Route("/costs", func(r chi.Router) {
				r.Get("/summary", h.Cost.GetSummary)
				r.Get("/by-agent", h.Cost.ByAgent)
				r.Get("/by-provider", h.Cost.ByProvider)
				r.Get("/history", h.Cost.GetHistory)
				r.Get("/limits", h.Cost.GetLimits)
				r.Patch("/limits", h.Cost.UpdateLimits)
			})

			// Dashboard data
			r.Route("/dashboard", func(r chi.Router) {
				r.Get("/overview", h.Dashboard.Overview)
				r.Get("/agents-status", h.Dashboard.AgentsStatus)
				r.Get("/recent-activity", h.Dashboard.RecentActivity)
				r.Get("/cost-trends", h.Dashboard.CostTrends)
				r.Get("/widgets/{widgetID}", h.Dashboard.GetWidget)
			})

			// Audit logs
			r.Route("/audit", func(r chi.Router) {
				r.Get("/", h.Audit.List)
				r.Get("/export", h.Audit.Export)
			})

			// Settings
			r.Route("/settings", func(r chi.Router) {
				r.Get("/", h.Settings.Get)
				r.Patch("/", h.Settings.Update)
				r.Get("/integrations", h.Settings.ListIntegrations)
				r.Post("/integrations/{integration}", h.Settings.ConfigureIntegration)
			})
		})

		// Webhook endpoints (authenticated via signature)
		r.Route("/webhooks", func(r chi.Router) {
			r.Post("/github", h.Webhook.GitHub)
			r.Post("/stripe", h.Webhook.Stripe)
		})
	})

	// WebSocket endpoint for real-time updates
	r.Get("/ws", h.WebSocket.Handle)

	// Create server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.APIPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		log.Info("Starting server", "port", cfg.APIPort, "env", cfg.Environment)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed", "error", err)
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
		log.Fatal("Server forced to shutdown", "error", err)
	}

	log.Info("Server stopped")
}

