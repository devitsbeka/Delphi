package security

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
)

// =============================================================================
// Types
// =============================================================================

// AuditAction represents an action type for audit logging
type AuditAction string

const (
	// Authentication actions
	AuditActionLogin          AuditAction = "auth.login"
	AuditActionLogout         AuditAction = "auth.logout"
	AuditActionLoginFailed    AuditAction = "auth.login_failed"
	AuditActionPasswordChange AuditAction = "auth.password_change"
	AuditActionMFAEnabled     AuditAction = "auth.mfa_enabled"
	AuditActionMFADisabled    AuditAction = "auth.mfa_disabled"

	// User actions
	AuditActionUserCreated  AuditAction = "user.created"
	AuditActionUserUpdated  AuditAction = "user.updated"
	AuditActionUserDeleted  AuditAction = "user.deleted"
	AuditActionRoleAssigned AuditAction = "user.role_assigned"

	// Agent actions
	AuditActionAgentCreated   AuditAction = "agent.created"
	AuditActionAgentUpdated   AuditAction = "agent.updated"
	AuditActionAgentDeleted   AuditAction = "agent.deleted"
	AuditActionAgentExecuted  AuditAction = "agent.executed"

	// API key actions
	AuditActionAPIKeyCreated  AuditAction = "apikey.created"
	AuditActionAPIKeyUpdated  AuditAction = "apikey.updated"
	AuditActionAPIKeyRevoked  AuditAction = "apikey.revoked"

	// Repository actions
	AuditActionRepoConnected    AuditAction = "repo.connected"
	AuditActionRepoDisconnected AuditAction = "repo.disconnected"
	AuditActionCommitCreated    AuditAction = "repo.commit_created"
	AuditActionPRCreated        AuditAction = "repo.pr_created"

	// Billing actions
	AuditActionPlanChanged     AuditAction = "billing.plan_changed"
	AuditActionPaymentReceived AuditAction = "billing.payment_received"
	AuditActionPaymentFailed   AuditAction = "billing.payment_failed"

	// Settings actions
	AuditActionSettingsChanged AuditAction = "settings.changed"

	// Data access actions
	AuditActionDataExported AuditAction = "data.exported"
	AuditActionDataDeleted  AuditAction = "data.deleted"
)

// AuditSeverity represents the severity of an audit event
type AuditSeverity string

const (
	SeverityInfo     AuditSeverity = "info"
	SeverityWarning  AuditSeverity = "warning"
	SeverityCritical AuditSeverity = "critical"
)

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ID          uuid.UUID              `json:"id"`
	TenantID    uuid.UUID              `json:"tenant_id"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	Action      AuditAction            `json:"action"`
	Severity    AuditSeverity          `json:"severity"`
	ResourceType string                `json:"resource_type"`
	ResourceID  *uuid.UUID             `json:"resource_id,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
	IPAddress   string                 `json:"ip_address,omitempty"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	SessionID   string                 `json:"session_id,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

// =============================================================================
// Audit Service
// =============================================================================

// AuditService handles audit logging
type AuditService struct {
	log     *logger.Logger
	buffer  chan *AuditEntry
	storage AuditStorage
}

// AuditStorage interface for storing audit logs
type AuditStorage interface {
	Store(ctx context.Context, entry *AuditEntry) error
	Query(ctx context.Context, query AuditQuery) ([]*AuditEntry, error)
	GetByID(ctx context.Context, id uuid.UUID) (*AuditEntry, error)
}

// AuditQuery defines parameters for querying audit logs
type AuditQuery struct {
	TenantID     uuid.UUID
	UserID       *uuid.UUID
	Actions      []AuditAction
	ResourceType string
	ResourceID   *uuid.UUID
	Severity     *AuditSeverity
	StartTime    *time.Time
	EndTime      *time.Time
	Limit        int
	Offset       int
}

// NewAuditService creates a new audit service
func NewAuditService(log *logger.Logger, storage AuditStorage) *AuditService {
	s := &AuditService{
		log:     log,
		buffer:  make(chan *AuditEntry, 10000),
		storage: storage,
	}

	// Start background processor
	go s.processBuffer()

	return s
}

// Log creates an audit log entry
func (s *AuditService) Log(ctx context.Context, entry *AuditEntry) {
	entry.ID = uuid.New()
	entry.Timestamp = time.Now()

	if entry.Severity == "" {
		entry.Severity = SeverityInfo
	}

	select {
	case s.buffer <- entry:
	default:
		// Buffer full, log directly
		s.log.Warnw("audit buffer full, logging directly",
			"action", entry.Action,
			"tenant_id", entry.TenantID,
		)
		if s.storage != nil {
			s.storage.Store(ctx, entry)
		}
	}
}

// processBuffer processes buffered audit entries
func (s *AuditService) processBuffer() {
	batch := make([]*AuditEntry, 0, 100)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case entry := <-s.buffer:
			batch = append(batch, entry)
			if len(batch) >= 100 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch writes a batch of audit entries to storage
func (s *AuditService) flushBatch(batch []*AuditEntry) {
	ctx := context.Background()
	for _, entry := range batch {
		if s.storage != nil {
			if err := s.storage.Store(ctx, entry); err != nil {
				s.log.Errorw("failed to store audit entry",
					"error", err,
					"entry_id", entry.ID,
				)
			}
		}

		// Also log to structured logger
		s.log.Infow("audit",
			"action", entry.Action,
			"tenant_id", entry.TenantID,
			"user_id", entry.UserID,
			"severity", entry.Severity,
			"resource_type", entry.ResourceType,
			"resource_id", entry.ResourceID,
		)
	}
}

// Query retrieves audit entries based on filters
func (s *AuditService) Query(ctx context.Context, query AuditQuery) ([]*AuditEntry, error) {
	if s.storage == nil {
		return nil, fmt.Errorf("no audit storage configured")
	}
	return s.storage.Query(ctx, query)
}

// =============================================================================
// Convenience Methods
// =============================================================================

// LogLogin logs a successful login
func (s *AuditService) LogLogin(ctx context.Context, tenantID, userID uuid.UUID, ipAddress, userAgent string) {
	s.Log(ctx, &AuditEntry{
		TenantID:     tenantID,
		UserID:       &userID,
		Action:       AuditActionLogin,
		Severity:     SeverityInfo,
		ResourceType: "user",
		ResourceID:   &userID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	})
}

// LogLoginFailed logs a failed login attempt
func (s *AuditService) LogLoginFailed(ctx context.Context, tenantID uuid.UUID, email, ipAddress, userAgent, reason string) {
	s.Log(ctx, &AuditEntry{
		TenantID:     tenantID,
		Action:       AuditActionLoginFailed,
		Severity:     SeverityWarning,
		ResourceType: "user",
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details: map[string]interface{}{
			"email":  email,
			"reason": reason,
		},
	})
}

// LogAgentExecution logs an agent execution
func (s *AuditService) LogAgentExecution(ctx context.Context, tenantID, userID, agentID, executionID uuid.UUID, task string) {
	s.Log(ctx, &AuditEntry{
		TenantID:     tenantID,
		UserID:       &userID,
		Action:       AuditActionAgentExecuted,
		Severity:     SeverityInfo,
		ResourceType: "agent",
		ResourceID:   &agentID,
		Details: map[string]interface{}{
			"execution_id": executionID,
			"task":         task,
		},
	})
}

// LogAPIKeyCreated logs API key creation
func (s *AuditService) LogAPIKeyCreated(ctx context.Context, tenantID, userID uuid.UUID, provider string) {
	s.Log(ctx, &AuditEntry{
		TenantID:     tenantID,
		UserID:       &userID,
		Action:       AuditActionAPIKeyCreated,
		Severity:     SeverityCritical,
		ResourceType: "api_key",
		Details: map[string]interface{}{
			"provider": provider,
		},
	})
}

// LogSettingsChanged logs settings changes
func (s *AuditService) LogSettingsChanged(ctx context.Context, tenantID, userID uuid.UUID, setting string, oldValue, newValue interface{}) {
	s.Log(ctx, &AuditEntry{
		TenantID:     tenantID,
		UserID:       &userID,
		Action:       AuditActionSettingsChanged,
		Severity:     SeverityInfo,
		ResourceType: "settings",
		Details: map[string]interface{}{
			"setting":   setting,
			"old_value": oldValue,
			"new_value": newValue,
		},
	})
}

// =============================================================================
// RBAC (Role-Based Access Control)
// =============================================================================

// Role represents a user role
type Role string

const (
	RoleOwner    Role = "owner"
	RoleAdmin    Role = "admin"
	RoleMember   Role = "member"
	RoleViewer   Role = "viewer"
	RoleAgent    Role = "agent" // For AI agents
)

// Permission represents a permission
type Permission string

const (
	// Agent permissions
	PermAgentCreate  Permission = "agent:create"
	PermAgentRead    Permission = "agent:read"
	PermAgentUpdate  Permission = "agent:update"
	PermAgentDelete  Permission = "agent:delete"
	PermAgentExecute Permission = "agent:execute"

	// Repository permissions
	PermRepoConnect    Permission = "repo:connect"
	PermRepoRead       Permission = "repo:read"
	PermRepoDisconnect Permission = "repo:disconnect"

	// User permissions
	PermUserCreate Permission = "user:create"
	PermUserRead   Permission = "user:read"
	PermUserUpdate Permission = "user:update"
	PermUserDelete Permission = "user:delete"

	// Settings permissions
	PermSettingsRead   Permission = "settings:read"
	PermSettingsUpdate Permission = "settings:update"

	// Billing permissions
	PermBillingRead   Permission = "billing:read"
	PermBillingUpdate Permission = "billing:update"

	// API key permissions
	PermAPIKeyCreate Permission = "apikey:create"
	PermAPIKeyRead   Permission = "apikey:read"
	PermAPIKeyRevoke Permission = "apikey:revoke"
)

// RolePermissions maps roles to their permissions
var RolePermissions = map[Role][]Permission{
	RoleOwner: {
		PermAgentCreate, PermAgentRead, PermAgentUpdate, PermAgentDelete, PermAgentExecute,
		PermRepoConnect, PermRepoRead, PermRepoDisconnect,
		PermUserCreate, PermUserRead, PermUserUpdate, PermUserDelete,
		PermSettingsRead, PermSettingsUpdate,
		PermBillingRead, PermBillingUpdate,
		PermAPIKeyCreate, PermAPIKeyRead, PermAPIKeyRevoke,
	},
	RoleAdmin: {
		PermAgentCreate, PermAgentRead, PermAgentUpdate, PermAgentDelete, PermAgentExecute,
		PermRepoConnect, PermRepoRead, PermRepoDisconnect,
		PermUserCreate, PermUserRead, PermUserUpdate,
		PermSettingsRead, PermSettingsUpdate,
		PermAPIKeyCreate, PermAPIKeyRead, PermAPIKeyRevoke,
	},
	RoleMember: {
		PermAgentRead, PermAgentExecute,
		PermRepoRead,
		PermUserRead,
		PermSettingsRead,
	},
	RoleViewer: {
		PermAgentRead,
		PermRepoRead,
		PermUserRead,
	},
}

// RBAC handles role-based access control
type RBAC struct {
	log *logger.Logger
}

// NewRBAC creates a new RBAC instance
func NewRBAC(log *logger.Logger) *RBAC {
	return &RBAC{log: log}
}

// HasPermission checks if a role has a specific permission
func (r *RBAC) HasPermission(role Role, permission Permission) bool {
	permissions, ok := RolePermissions[role]
	if !ok {
		return false
	}

	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// CanAccess checks if a user with given role can perform an action
func (r *RBAC) CanAccess(role Role, resource string, action string) bool {
	permission := Permission(fmt.Sprintf("%s:%s", resource, action))
	return r.HasPermission(role, permission)
}

// =============================================================================
// Rate Limiting
// =============================================================================

// RateLimiter implements rate limiting for API calls
type RateLimiter struct {
	log      *logger.Logger
	requests map[string]*rateLimitBucket
	mu       sync.RWMutex
}

type rateLimitBucket struct {
	tokens    float64
	lastFill  time.Time
	maxTokens float64
	fillRate  float64 // tokens per second
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(log *logger.Logger) *RateLimiter {
	return &RateLimiter{
		log:      log,
		requests: make(map[string]*rateLimitBucket),
	}
}

// Allow checks if a request is allowed
func (r *RateLimiter) Allow(key string, maxRequests int, window time.Duration) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	bucket, ok := r.requests[key]
	if !ok {
		bucket = &rateLimitBucket{
			tokens:    float64(maxRequests),
			lastFill:  time.Now(),
			maxTokens: float64(maxRequests),
			fillRate:  float64(maxRequests) / window.Seconds(),
		}
		r.requests[key] = bucket
	}

	// Refill tokens
	now := time.Now()
	elapsed := now.Sub(bucket.lastFill).Seconds()
	bucket.tokens = min(bucket.maxTokens, bucket.tokens+elapsed*bucket.fillRate)
	bucket.lastFill = now

	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}

	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

