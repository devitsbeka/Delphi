package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repositories contains all repository instances
type Repositories struct {
	db          *PostgresDB
	Tenants     *TenantRepository
	Users       *UserRepository
	APIKeys     *APIKeyRepository
	Agents      *AgentRepository
	AgentRuns   *AgentRunRepository
	Knowledge   *KnowledgeRepository
	Repositories *RepositoryRepository
	Businesses  *BusinessRepository
	Projects    *ProjectRepository
	Financial   *FinancialRepository
	Social      *SocialRepository
	IoT         *IoTRepository
	Audit       *AuditRepository
	Costs       *CostRepository
}

// NewRepositories creates all repository instances
func NewRepositories(db *PostgresDB) *Repositories {
	return &Repositories{
		db:           db,
		Tenants:      &TenantRepository{db: db},
		Users:        &UserRepository{db: db},
		APIKeys:      &APIKeyRepository{db: db},
		Agents:       &AgentRepository{db: db},
		AgentRuns:    &AgentRunRepository{db: db},
		Knowledge:    &KnowledgeRepository{db: db},
		Repositories: &RepositoryRepository{db: db},
		Businesses:   &BusinessRepository{db: db},
		Projects:     &ProjectRepository{db: db},
		Financial:    &FinancialRepository{db: db},
		Social:       &SocialRepository{db: db},
		IoT:          &IoTRepository{db: db},
		Audit:        &AuditRepository{db: db},
		Costs:        &CostRepository{db: db},
	}
}

// =============================================================================
// Tenant Repository
// =============================================================================

type TenantRepository struct {
	db *PostgresDB
}

func (r *TenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, slug, plan, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.pool.Exec(ctx, query,
		tenant.ID, tenant.Name, tenant.Slug, tenant.Plan, tenant.Settings,
		tenant.CreatedAt, tenant.UpdatedAt)
	return err
}

func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	query := `SELECT id, name, slug, plan, settings, created_at, updated_at FROM tenants WHERE id = $1`
	var tenant models.Tenant
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Plan, &tenant.Settings,
		&tenant.CreatedAt, &tenant.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &tenant, err
}

func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	query := `SELECT id, name, slug, plan, settings, created_at, updated_at FROM tenants WHERE slug = $1`
	var tenant models.Tenant
	err := r.db.pool.QueryRow(ctx, query, slug).Scan(
		&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Plan, &tenant.Settings,
		&tenant.CreatedAt, &tenant.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &tenant, err
}

func (r *TenantRepository) Update(ctx context.Context, tenant *models.Tenant) error {
	query := `
		UPDATE tenants SET name = $2, plan = $3, settings = $4, updated_at = $5
		WHERE id = $1
	`
	_, err := r.db.pool.Exec(ctx, query,
		tenant.ID, tenant.Name, tenant.Plan, tenant.Settings, time.Now())
	return err
}

// =============================================================================
// User Repository
// =============================================================================

type UserRepository struct {
	db *PostgresDB
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, tenant_id, email, name, role, preferences, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.pool.Exec(ctx, query,
		user.ID, user.TenantID, user.Email, user.Name, user.Role, user.Preferences,
		user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `SELECT id, tenant_id, email, name, role, preferences, created_at, updated_at, last_login_at 
			  FROM users WHERE id = $1`
	var user models.User
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.Name, &user.Role, &user.Preferences,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, tenant_id, email, name, role, preferences, created_at, updated_at, last_login_at 
			  FROM users WHERE email = $1`
	var user models.User
	err := r.db.pool.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.Name, &user.Role, &user.Preferences,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &user, err
}

func (r *UserRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.User, error) {
	query := `SELECT id, tenant_id, email, name, role, preferences, created_at, updated_at, last_login_at 
			  FROM users WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID, &user.TenantID, &user.Email, &user.Name, &user.Role, &user.Preferences,
			&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	return users, rows.Err()
}

func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE users SET last_login_at = $2 WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, userID, time.Now())
	return err
}

// =============================================================================
// API Key Repository
// =============================================================================

type APIKeyRepository struct {
	db *PostgresDB
}

func (r *APIKeyRepository) Create(ctx context.Context, key *models.APIKey) error {
	query := `
		INSERT INTO api_keys (id, tenant_id, provider, name, encrypted_key, is_valid, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.pool.Exec(ctx, query,
		key.ID, key.TenantID, key.Provider, key.Name, key.EncryptedKey, key.IsValid, key.CreatedAt)
	return err
}

func (r *APIKeyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.APIKey, error) {
	query := `SELECT id, tenant_id, provider, name, encrypted_key, is_valid, last_used_at, created_at 
			  FROM api_keys WHERE id = $1`
	var key models.APIKey
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&key.ID, &key.TenantID, &key.Provider, &key.Name, &key.EncryptedKey,
		&key.IsValid, &key.LastUsedAt, &key.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &key, err
}

func (r *APIKeyRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.APIKey, error) {
	query := `SELECT id, tenant_id, provider, name, is_valid, last_used_at, created_at 
			  FROM api_keys WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*models.APIKey
	for rows.Next() {
		var key models.APIKey
		if err := rows.Scan(
			&key.ID, &key.TenantID, &key.Provider, &key.Name,
			&key.IsValid, &key.LastUsedAt, &key.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, &key)
	}
	return keys, rows.Err()
}

func (r *APIKeyRepository) GetByTenantAndProvider(ctx context.Context, tenantID uuid.UUID, provider models.AIProvider) (*models.APIKey, error) {
	query := `SELECT id, tenant_id, provider, name, encrypted_key, is_valid, last_used_at, created_at 
			  FROM api_keys WHERE tenant_id = $1 AND provider = $2 AND is_valid = true
			  ORDER BY created_at DESC LIMIT 1`
	var key models.APIKey
	err := r.db.pool.QueryRow(ctx, query, tenantID, provider).Scan(
		&key.ID, &key.TenantID, &key.Provider, &key.Name, &key.EncryptedKey,
		&key.IsValid, &key.LastUsedAt, &key.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &key, err
}

func (r *APIKeyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM api_keys WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id)
	return err
}

func (r *APIKeyRepository) UpdateLastUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE api_keys SET last_used_at = $2 WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id, time.Now())
	return err
}

// =============================================================================
// Agent Repository
// =============================================================================

type AgentRepository struct {
	db *PostgresDB
}

func (r *AgentRepository) Create(ctx context.Context, agent *models.Agent) error {
	configJSON, _ := json.Marshal(agent.Config)
	kbJSON, _ := json.Marshal(agent.KnowledgeBases)
	query := `
		INSERT INTO agents (id, tenant_id, name, description, type, provider, model, system_prompt, 
						   tools, knowledge_bases, config, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.db.pool.Exec(ctx, query,
		agent.ID, agent.TenantID, agent.Name, agent.Description, agent.Type,
		agent.Provider, agent.Model, agent.SystemPrompt, agent.Tools, kbJSON, configJSON,
		agent.Status, agent.CreatedAt, agent.UpdatedAt)
	return err
}

func (r *AgentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Agent, error) {
	query := `SELECT id, tenant_id, name, description, type, provider, model, system_prompt, 
					 tools, knowledge_bases, config, status, created_at, updated_at 
			  FROM agents WHERE id = $1`
	var agent models.Agent
	var configJSON, kbJSON []byte
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&agent.ID, &agent.TenantID, &agent.Name, &agent.Description, &agent.Type,
		&agent.Provider, &agent.Model, &agent.SystemPrompt, &agent.Tools, &kbJSON, &configJSON,
		&agent.Status, &agent.CreatedAt, &agent.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	json.Unmarshal(configJSON, &agent.Config)
	json.Unmarshal(kbJSON, &agent.KnowledgeBases)
	return &agent, nil
}

func (r *AgentRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID) ([]*models.Agent, error) {
	query := `SELECT id, tenant_id, name, description, type, provider, model, system_prompt, 
					 tools, knowledge_bases, config, status, created_at, updated_at 
			  FROM agents WHERE tenant_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*models.Agent
	for rows.Next() {
		var agent models.Agent
		var configJSON, kbJSON []byte
		if err := rows.Scan(
			&agent.ID, &agent.TenantID, &agent.Name, &agent.Description, &agent.Type,
			&agent.Provider, &agent.Model, &agent.SystemPrompt, &agent.Tools, &kbJSON, &configJSON,
			&agent.Status, &agent.CreatedAt, &agent.UpdatedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(configJSON, &agent.Config)
		json.Unmarshal(kbJSON, &agent.KnowledgeBases)
		agents = append(agents, &agent)
	}
	return agents, rows.Err()
}

func (r *AgentRepository) Update(ctx context.Context, agent *models.Agent) error {
	configJSON, _ := json.Marshal(agent.Config)
	kbJSON, _ := json.Marshal(agent.KnowledgeBases)
	query := `
		UPDATE agents SET name = $2, description = $3, type = $4, provider = $5, model = $6,
						  system_prompt = $7, tools = $8, knowledge_bases = $9, config = $10,
						  status = $11, updated_at = $12
		WHERE id = $1
	`
	_, err := r.db.pool.Exec(ctx, query,
		agent.ID, agent.Name, agent.Description, agent.Type, agent.Provider, agent.Model,
		agent.SystemPrompt, agent.Tools, kbJSON, configJSON, agent.Status, time.Now())
	return err
}

func (r *AgentRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.AgentStatus) error {
	query := `UPDATE agents SET status = $2, updated_at = $3 WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id, status, time.Now())
	return err
}

func (r *AgentRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM agents WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id)
	return err
}

// =============================================================================
// Agent Run Repository
// =============================================================================

type AgentRunRepository struct {
	db *PostgresDB
}

func (r *AgentRunRepository) Create(ctx context.Context, run *models.AgentRun) error {
	query := `
		INSERT INTO agent_runs (id, agent_id, tenant_id, prompt, status, machine_id, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.pool.Exec(ctx, query,
		run.ID, run.AgentID, run.TenantID, run.Prompt, run.Status, run.MachineID, run.StartedAt)
	return err
}

func (r *AgentRunRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AgentRun, error) {
	query := `SELECT id, agent_id, tenant_id, prompt, status, result, tokens_used, cost, 
					 machine_id, started_at, completed_at, error 
			  FROM agent_runs WHERE id = $1`
	var run models.AgentRun
	err := r.db.pool.QueryRow(ctx, query, id).Scan(
		&run.ID, &run.AgentID, &run.TenantID, &run.Prompt, &run.Status, &run.Result,
		&run.TokensUsed, &run.Cost, &run.MachineID, &run.StartedAt, &run.CompletedAt, &run.Error)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &run, err
}

func (r *AgentRunRepository) ListByAgent(ctx context.Context, agentID uuid.UUID, limit int) ([]*models.AgentRun, error) {
	query := `SELECT id, agent_id, tenant_id, prompt, status, result, tokens_used, cost, 
					 machine_id, started_at, completed_at, error 
			  FROM agent_runs WHERE agent_id = $1 ORDER BY started_at DESC LIMIT $2`
	rows, err := r.db.pool.Query(ctx, query, agentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []*models.AgentRun
	for rows.Next() {
		var run models.AgentRun
		if err := rows.Scan(
			&run.ID, &run.AgentID, &run.TenantID, &run.Prompt, &run.Status, &run.Result,
			&run.TokensUsed, &run.Cost, &run.MachineID, &run.StartedAt, &run.CompletedAt, &run.Error); err != nil {
			return nil, err
		}
		runs = append(runs, &run)
	}
	return runs, rows.Err()
}

func (r *AgentRunRepository) Complete(ctx context.Context, id uuid.UUID, result json.RawMessage, tokensUsed int, cost float64) error {
	query := `UPDATE agent_runs SET status = $2, result = $3, tokens_used = $4, cost = $5, completed_at = $6 WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id, models.RunStatusCompleted, result, tokensUsed, cost, time.Now())
	return err
}

func (r *AgentRunRepository) Fail(ctx context.Context, id uuid.UUID, errMsg string) error {
	query := `UPDATE agent_runs SET status = $2, error = $3, completed_at = $4 WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id, models.RunStatusFailed, errMsg, time.Now())
	return err
}

func (r *AgentRunRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.RunStatus) error {
	query := `UPDATE agent_runs SET status = $2 WHERE id = $1`
	_, err := r.db.pool.Exec(ctx, query, id, status)
	return err
}

// =============================================================================
// Placeholder repositories for other entities
// =============================================================================

type KnowledgeRepository struct {
	db *PostgresDB
}

type RepositoryRepository struct {
	db *PostgresDB
}

type BusinessRepository struct {
	db *PostgresDB
}

type ProjectRepository struct {
	db *PostgresDB
}

type FinancialRepository struct {
	db *PostgresDB
}

type SocialRepository struct {
	db *PostgresDB
}

type IoTRepository struct {
	db *PostgresDB
}

type AuditRepository struct {
	db *PostgresDB
}

func (r *AuditRepository) Create(ctx context.Context, log *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (id, tenant_id, user_id, agent_id, action, resource_type, resource_id,
							   old_value, new_value, ip_address, user_agent, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.db.pool.Exec(ctx, query,
		log.ID, log.TenantID, log.UserID, log.AgentID, log.Action, log.ResourceType,
		log.ResourceID, log.OldValue, log.NewValue, log.IPAddress, log.UserAgent, log.CreatedAt)
	return err
}

type CostRepository struct {
	db *PostgresDB
}

func (r *CostRepository) RecordCost(ctx context.Context, record *models.CostRecord) error {
	query := `
		INSERT INTO cost_records (id, tenant_id, agent_id, run_id, provider, model, 
								 input_tokens, output_tokens, cost, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.db.pool.Exec(ctx, query,
		record.ID, record.TenantID, record.AgentID, record.RunID, record.Provider,
		record.Model, record.InputTokens, record.OutputTokens, record.Cost, record.CreatedAt)
	return err
}

func (r *CostRepository) GetTotalByTenant(ctx context.Context, tenantID uuid.UUID, since time.Time) (float64, error) {
	query := `SELECT COALESCE(SUM(cost), 0) FROM cost_records WHERE tenant_id = $1 AND created_at >= $2`
	var total float64
	err := r.db.pool.QueryRow(ctx, query, tenantID, since).Scan(&total)
	return total, err
}

func (r *CostRepository) GetTotalByAgent(ctx context.Context, agentID uuid.UUID, since time.Time) (float64, error) {
	query := `SELECT COALESCE(SUM(cost), 0) FROM cost_records WHERE agent_id = $1 AND created_at >= $2`
	var total float64
	err := r.db.pool.QueryRow(ctx, query, agentID, since).Scan(&total)
	return total, err
}

// GetLimit retrieves cost limit for tenant or agent
func (r *CostRepository) GetLimit(ctx context.Context, tenantID uuid.UUID, agentID *uuid.UUID, limitType string) (*models.CostLimit, error) {
	var query string
	var args []interface{}
	
	if agentID != nil {
		query = `SELECT id, tenant_id, agent_id, limit_type, amount, alert_at, created_at 
				 FROM cost_limits WHERE tenant_id = $1 AND agent_id = $2 AND limit_type = $3`
		args = []interface{}{tenantID, agentID, limitType}
	} else {
		query = `SELECT id, tenant_id, agent_id, limit_type, amount, alert_at, created_at 
				 FROM cost_limits WHERE tenant_id = $1 AND agent_id IS NULL AND limit_type = $2`
		args = []interface{}{tenantID, limitType}
	}
	
	var limit models.CostLimit
	err := r.db.pool.QueryRow(ctx, query, args...).Scan(
		&limit.ID, &limit.TenantID, &limit.AgentID, &limit.LimitType,
		&limit.Amount, &limit.AlertAt, &limit.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &limit, err
}

// SetLimit creates or updates a cost limit
func (r *CostRepository) SetLimit(ctx context.Context, limit *models.CostLimit) error {
	query := `
		INSERT INTO cost_limits (id, tenant_id, agent_id, limit_type, amount, alert_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (tenant_id, agent_id, limit_type) 
		DO UPDATE SET amount = EXCLUDED.amount, alert_at = EXCLUDED.alert_at
	`
	_, err := r.db.pool.Exec(ctx, query,
		limit.ID, limit.TenantID, limit.AgentID, limit.LimitType,
		limit.Amount, limit.AlertAt, limit.CreatedAt)
	return err
}

// Health check for repositories
func (r *Repositories) Ping(ctx context.Context) error {
	return r.db.Ping(ctx)
}

// Helper function to generate error messages
func ErrNotFound(entity string, id interface{}) error {
	return fmt.Errorf("%s not found: %v", entity, id)
}

