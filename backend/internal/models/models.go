package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// Core Entity Types
// =============================================================================

// Tenant represents a customer organization
type Tenant struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	Name      string          `json:"name" db:"name"`
	Slug      string          `json:"slug" db:"slug"`
	Plan      TenantPlan      `json:"plan" db:"plan"`
	Settings  json.RawMessage `json:"settings" db:"settings"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

type TenantPlan string

const (
	PlanFree       TenantPlan = "free"
	PlanPro        TenantPlan = "pro"
	PlanEnterprise TenantPlan = "enterprise"
)

// User represents a platform user
type User struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	TenantID    uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Email       string          `json:"email" db:"email"`
	Name        string          `json:"name" db:"name"`
	Role        UserRole        `json:"role" db:"role"`
	Preferences json.RawMessage `json:"preferences" db:"preferences"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
	LastLoginAt *time.Time      `json:"last_login_at" db:"last_login_at"`
}

type UserRole string

const (
	RoleOwner     UserRole = "owner"
	RoleAdmin     UserRole = "admin"
	RoleDeveloper UserRole = "developer"
	RoleViewer    UserRole = "viewer"
	RoleBilling   UserRole = "billing"
)

// =============================================================================
// API Keys (Provider Credentials)
// =============================================================================

// APIKey stores encrypted provider API keys
type APIKey struct {
	ID           uuid.UUID     `json:"id" db:"id"`
	TenantID     uuid.UUID     `json:"tenant_id" db:"tenant_id"`
	Provider     AIProvider    `json:"provider" db:"provider"`
	Name         string        `json:"name" db:"name"`
	EncryptedKey string        `json:"-" db:"encrypted_key"`
	IsValid      bool          `json:"is_valid" db:"is_valid"`
	LastUsedAt   *time.Time    `json:"last_used_at" db:"last_used_at"`
	CreatedAt    time.Time     `json:"created_at" db:"created_at"`
}

type AIProvider string

const (
	ProviderOpenAI    AIProvider = "openai"
	ProviderAnthropic AIProvider = "anthropic"
	ProviderGoogle    AIProvider = "google"
	ProviderOllama    AIProvider = "ollama"
	ProviderCustom    AIProvider = "custom"
)

// =============================================================================
// Agents
// =============================================================================

// Agent represents an AI agent configuration
type Agent struct {
	ID             uuid.UUID       `json:"id" db:"id"`
	TenantID       uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name           string          `json:"name" db:"name"`
	Description    string          `json:"description" db:"description"`
	Type           AgentType       `json:"type" db:"type"`
	Provider       AIProvider      `json:"provider" db:"provider"`
	Model          string          `json:"model" db:"model"`
	SystemPrompt   string          `json:"system_prompt" db:"system_prompt"`
	Tools          json.RawMessage `json:"tools" db:"tools"`
	KnowledgeBases []uuid.UUID     `json:"knowledge_bases" db:"knowledge_bases"`
	Config         AgentConfig     `json:"config" db:"config"`
	Status         AgentStatus     `json:"status" db:"status"`
	CreatedAt      time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at" db:"updated_at"`
}

type AgentType string

const (
	AgentTypeCoding     AgentType = "coding"
	AgentTypeBusiness   AgentType = "business"
	AgentTypeAccounting AgentType = "accounting"
	AgentTypeMarketing  AgentType = "marketing"
	AgentTypeProduct    AgentType = "product"
	AgentTypeAssistant  AgentType = "assistant"
	AgentTypeCustom     AgentType = "custom"
)

type AgentStatus string

const (
	AgentStatusConfigured AgentStatus = "configured"
	AgentStatusBriefing   AgentStatus = "briefing"
	AgentStatusReady      AgentStatus = "ready"
	AgentStatusExecuting  AgentStatus = "executing"
	AgentStatusPaused     AgentStatus = "paused"
	AgentStatusError      AgentStatus = "error"
	AgentStatusTerminated AgentStatus = "terminated"
)

type AgentConfig struct {
	Temperature      float64     `json:"temperature"`
	MaxTokens        int         `json:"max_tokens"`
	BudgetLimit      float64     `json:"budget_limit"`
	TimeoutSeconds   int         `json:"timeout_seconds"`
	RetryPolicy      RetryPolicy `json:"retry_policy"`
	BriefingRequired bool        `json:"briefing_required"`
	BriefingDepth    string      `json:"briefing_depth"` // quick, standard, full
}

type RetryPolicy struct {
	MaxRetries  int `json:"max_retries"`
	BackoffMs   int `json:"backoff_ms"`
	MaxBackoffMs int `json:"max_backoff_ms"`
}

// AgentTemplate provides pre-configured agent templates
type AgentTemplate struct {
	ID            uuid.UUID       `json:"id" db:"id"`
	Name          string          `json:"name" db:"name"`
	Description   string          `json:"description" db:"description"`
	Type          AgentType       `json:"type" db:"type"`
	SystemPrompt  string          `json:"system_prompt" db:"system_prompt"`
	DefaultConfig AgentConfig     `json:"default_config" db:"default_config"`
	Tools         json.RawMessage `json:"tools" db:"tools"`
	Category      string          `json:"category" db:"category"`
	IsPublic      bool            `json:"is_public" db:"is_public"`
}

// =============================================================================
// Agent Runs and Logs
// =============================================================================

// AgentRun represents a single execution of an agent
type AgentRun struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	AgentID     uuid.UUID       `json:"agent_id" db:"agent_id"`
	TenantID    uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Prompt      string          `json:"prompt" db:"prompt"`
	Status      RunStatus       `json:"status" db:"status"`
	Result      json.RawMessage `json:"result" db:"result"`
	TokensUsed  int             `json:"tokens_used" db:"tokens_used"`
	Cost        float64         `json:"cost" db:"cost"`
	MachineID   string          `json:"machine_id" db:"machine_id"`
	StartedAt   time.Time       `json:"started_at" db:"started_at"`
	CompletedAt *time.Time      `json:"completed_at" db:"completed_at"`
	Error       string          `json:"error,omitempty" db:"error"`
}

type RunStatus string

const (
	RunStatusPending    RunStatus = "pending"
	RunStatusBriefing   RunStatus = "briefing"
	RunStatusRunning    RunStatus = "running"
	RunStatusCompleted  RunStatus = "completed"
	RunStatusFailed     RunStatus = "failed"
	RunStatusCancelled  RunStatus = "cancelled"
)

// AgentLog represents a log entry from an agent run
type AgentLog struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	RunID     uuid.UUID       `json:"run_id" db:"run_id"`
	Level     LogLevel        `json:"level" db:"level"`
	Message   string          `json:"message" db:"message"`
	Metadata  json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
}

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

// =============================================================================
// Knowledge Base
// =============================================================================

type KnowledgeBase struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	TenantID  uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name      string          `json:"name" db:"name"`
	Type      KnowledgeType   `json:"type" db:"type"`
	Config    json.RawMessage `json:"config" db:"config"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

type KnowledgeType string

const (
	KnowledgeTypeGeneral    KnowledgeType = "general"
	KnowledgeTypeRepository KnowledgeType = "repository"
	KnowledgeTypeProject    KnowledgeType = "project"
)

type KnowledgeDocument struct {
	ID              uuid.UUID       `json:"id" db:"id"`
	KnowledgeBaseID uuid.UUID       `json:"knowledge_base_id" db:"knowledge_base_id"`
	Source          string          `json:"source" db:"source"`
	SourceType      string          `json:"source_type" db:"source_type"`
	ContentHash     string          `json:"content_hash" db:"content_hash"`
	Metadata        json.RawMessage `json:"metadata" db:"metadata"`
	ChunkCount      int             `json:"chunk_count" db:"chunk_count"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at" db:"updated_at"`
}

type KnowledgeChunk struct {
	ID         uuid.UUID `json:"id" db:"id"`
	DocumentID uuid.UUID `json:"document_id" db:"document_id"`
	Content    string    `json:"content" db:"content"`
	Embedding  []float32 `json:"embedding" db:"embedding"`
	Metadata   json.RawMessage `json:"metadata" db:"metadata"`
	ChunkIndex int       `json:"chunk_index" db:"chunk_index"`
}

// =============================================================================
// Repositories
// =============================================================================

type Repository struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	TenantID     uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name         string          `json:"name" db:"name"`
	FullName     string          `json:"full_name" db:"full_name"`
	URL          string          `json:"url" db:"url"`
	DefaultBranch string         `json:"default_branch" db:"default_branch"`
	IsPrivate    bool            `json:"is_private" db:"is_private"`
	LastSyncAt   *time.Time      `json:"last_sync_at" db:"last_sync_at"`
	Metadata     json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

// =============================================================================
// Business & Projects
// =============================================================================

type Business struct {
	ID        uuid.UUID       `json:"id" db:"id"`
	TenantID  uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name      string          `json:"name" db:"name"`
	Type      BusinessType    `json:"type" db:"type"`
	Settings  json.RawMessage `json:"settings" db:"settings"`
	Metadata  json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

type BusinessType string

const (
	BusinessTypeGameStudio BusinessType = "game_studio"
	BusinessTypeSaaS       BusinessType = "saas"
	BusinessTypeCrashGames BusinessType = "crash_games"
	BusinessTypeGeneral    BusinessType = "general"
)

type Project struct {
	ID           uuid.UUID   `json:"id" db:"id"`
	BusinessID   uuid.UUID   `json:"business_id" db:"business_id"`
	Name         string      `json:"name" db:"name"`
	Description  string      `json:"description" db:"description"`
	Repositories []uuid.UUID `json:"repositories" db:"repositories"`
	Agents       []uuid.UUID `json:"agents" db:"agents"`
	Status       string      `json:"status" db:"status"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
}

// =============================================================================
// Financial
// =============================================================================

type FinancialAccount struct {
	ID         uuid.UUID `json:"id" db:"id"`
	BusinessID uuid.UUID `json:"business_id" db:"business_id"`
	Name       string    `json:"name" db:"name"`
	Type       string    `json:"type" db:"type"`
	Currency   string    `json:"currency" db:"currency"`
	Balance    float64   `json:"balance" db:"balance"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

type Transaction struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	AccountID   uuid.UUID       `json:"account_id" db:"account_id"`
	Amount      float64         `json:"amount" db:"amount"`
	Category    string          `json:"category" db:"category"`
	Description string          `json:"description" db:"description"`
	Date        time.Time       `json:"date" db:"date"`
	Metadata    json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type Budget struct {
	ID         uuid.UUID `json:"id" db:"id"`
	BusinessID uuid.UUID `json:"business_id" db:"business_id"`
	Category   string    `json:"category" db:"category"`
	Amount     float64   `json:"amount" db:"amount"`
	Period     string    `json:"period" db:"period"` // monthly, quarterly, yearly
	StartDate  time.Time `json:"start_date" db:"start_date"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// =============================================================================
// Social Media
// =============================================================================

type SocialAccount struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	TenantID    uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Platform    string          `json:"platform" db:"platform"`
	AccountID   string          `json:"account_id" db:"account_id"`
	AccountName string          `json:"account_name" db:"account_name"`
	Credentials string          `json:"-" db:"credentials"` // encrypted
	Metadata    json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type SocialPost struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	AccountID   uuid.UUID       `json:"account_id" db:"account_id"`
	Content     string          `json:"content" db:"content"`
	MediaURLs   []string        `json:"media_urls" db:"media_urls"`
	Status      string          `json:"status" db:"status"` // draft, scheduled, published
	ScheduledAt *time.Time      `json:"scheduled_at" db:"scheduled_at"`
	PublishedAt *time.Time      `json:"published_at" db:"published_at"`
	Analytics   json.RawMessage `json:"analytics" db:"analytics"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

// =============================================================================
// IoT
// =============================================================================

type IoTDevice struct {
	ID          uuid.UUID       `json:"id" db:"id"`
	TenantID    uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	Name        string          `json:"name" db:"name"`
	Type        string          `json:"type" db:"type"`
	Protocol    string          `json:"protocol" db:"protocol"` // mqtt, http, websocket
	Endpoint    string          `json:"endpoint" db:"endpoint"`
	Credentials string          `json:"-" db:"credentials"` // encrypted
	Status      string          `json:"status" db:"status"`
	LastSeenAt  *time.Time      `json:"last_seen_at" db:"last_seen_at"`
	Metadata    json.RawMessage `json:"metadata" db:"metadata"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type IoTCommand struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	DeviceID   uuid.UUID       `json:"device_id" db:"device_id"`
	Command    string          `json:"command" db:"command"`
	Parameters json.RawMessage `json:"parameters" db:"parameters"`
	Status     string          `json:"status" db:"status"`
	Result     json.RawMessage `json:"result" db:"result"`
	ExecutedAt *time.Time      `json:"executed_at" db:"executed_at"`
	CreatedAt  time.Time       `json:"created_at" db:"created_at"`
}

type IoTTelemetry struct {
	ID         uuid.UUID       `json:"id" db:"id"`
	DeviceID   uuid.UUID       `json:"device_id" db:"device_id"`
	Data       json.RawMessage `json:"data" db:"data"`
	ReceivedAt time.Time       `json:"received_at" db:"received_at"`
}

// =============================================================================
// Audit Logging
// =============================================================================

type AuditLog struct {
	ID           uuid.UUID       `json:"id" db:"id"`
	TenantID     uuid.UUID       `json:"tenant_id" db:"tenant_id"`
	UserID       *uuid.UUID      `json:"user_id" db:"user_id"`
	AgentID      *uuid.UUID      `json:"agent_id" db:"agent_id"`
	Action       string          `json:"action" db:"action"`
	ResourceType string          `json:"resource_type" db:"resource_type"`
	ResourceID   string          `json:"resource_id" db:"resource_id"`
	OldValue     json.RawMessage `json:"old_value" db:"old_value"`
	NewValue     json.RawMessage `json:"new_value" db:"new_value"`
	IPAddress    string          `json:"ip_address" db:"ip_address"`
	UserAgent    string          `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
}

// =============================================================================
// Cost Tracking
// =============================================================================

type CostRecord struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	TenantID   uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	AgentID    *uuid.UUID `json:"agent_id" db:"agent_id"`
	RunID      *uuid.UUID `json:"run_id" db:"run_id"`
	Provider   AIProvider `json:"provider" db:"provider"`
	Model      string     `json:"model" db:"model"`
	InputTokens  int      `json:"input_tokens" db:"input_tokens"`
	OutputTokens int      `json:"output_tokens" db:"output_tokens"`
	Cost       float64    `json:"cost" db:"cost"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

type CostLimit struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	TenantID  uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	AgentID   *uuid.UUID `json:"agent_id" db:"agent_id"`
	LimitType string     `json:"limit_type" db:"limit_type"` // daily, monthly
	Amount    float64    `json:"amount" db:"amount"`
	AlertAt   float64    `json:"alert_at" db:"alert_at"` // percentage
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

