package social

import (
	"context"
	"fmt"
	"time"

	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
)

// =============================================================================
// Types
// =============================================================================

// Platform represents a social media platform
type Platform string

const (
	PlatformTwitter   Platform = "twitter"
	PlatformLinkedIn  Platform = "linkedin"
	PlatformInstagram Platform = "instagram"
	PlatformFacebook  Platform = "facebook"
	PlatformTikTok    Platform = "tiktok"
	PlatformDiscord   Platform = "discord"
	PlatformYouTube   Platform = "youtube"
)

// PostStatus represents the status of a social media post
type PostStatus string

const (
	PostStatusDraft     PostStatus = "draft"
	PostStatusScheduled PostStatus = "scheduled"
	PostStatusPublished PostStatus = "published"
	PostStatusFailed    PostStatus = "failed"
)

// Account represents a connected social media account
type Account struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	BusinessID   uuid.UUID `json:"business_id"`
	Platform     Platform  `json:"platform"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	ProfileURL   string    `json:"profile_url"`
	AvatarURL    string    `json:"avatar_url"`
	AccessToken  string    `json:"-"`
	RefreshToken string    `json:"-"`
	TokenExpiry  time.Time `json:"token_expiry"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Post represents a social media post
type Post struct {
	ID          uuid.UUID  `json:"id"`
	TenantID    uuid.UUID  `json:"tenant_id"`
	BusinessID  uuid.UUID  `json:"business_id"`
	AccountID   uuid.UUID  `json:"account_id"`
	AgentID     *uuid.UUID `json:"agent_id,omitempty"`
	Platform    Platform   `json:"platform"`
	Content     string     `json:"content"`
	MediaURLs   []string   `json:"media_urls,omitempty"`
	Status      PostStatus `json:"status"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	ExternalID  string     `json:"external_id,omitempty"`
	ExternalURL string     `json:"external_url,omitempty"`
	Metrics     PostMetrics `json:"metrics,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// PostMetrics contains engagement metrics for a post
type PostMetrics struct {
	Views      int `json:"views"`
	Likes      int `json:"likes"`
	Comments   int `json:"comments"`
	Shares     int `json:"shares"`
	Clicks     int `json:"clicks"`
	Impressions int `json:"impressions"`
}

// ContentPipeline represents a content generation pipeline
type ContentPipeline struct {
	ID          uuid.UUID         `json:"id"`
	TenantID    uuid.UUID         `json:"tenant_id"`
	BusinessID  uuid.UUID         `json:"business_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	AgentID     uuid.UUID         `json:"agent_id"`
	Schedule    string            `json:"schedule"` // Cron expression
	Platforms   []Platform        `json:"platforms"`
	Config      PipelineConfig    `json:"config"`
	IsActive    bool              `json:"is_active"`
	LastRunAt   *time.Time        `json:"last_run_at,omitempty"`
	NextRunAt   *time.Time        `json:"next_run_at,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// PipelineConfig contains configuration for a content pipeline
type PipelineConfig struct {
	Topics        []string `json:"topics"`
	Tone          string   `json:"tone"`
	Style         string   `json:"style"`
	MaxLength     int      `json:"max_length"`
	IncludeHashtags bool   `json:"include_hashtags"`
	IncludeEmojis   bool   `json:"include_emojis"`
	IncludeMedia    bool   `json:"include_media"`
	ReviewRequired  bool   `json:"review_required"`
}

// =============================================================================
// Service
// =============================================================================

// Service handles social media operations
type Service struct {
	log       *logger.Logger
	providers map[Platform]Provider
}

// NewService creates a new social media service
func NewService(log *logger.Logger) *Service {
	return &Service{
		log:       log,
		providers: make(map[Platform]Provider),
	}
}

// Provider interface for social media platforms
type Provider interface {
	// Connect initiates OAuth flow for the platform
	Connect(ctx context.Context, tenantID uuid.UUID) (string, error)
	
	// HandleCallback processes OAuth callback
	HandleCallback(ctx context.Context, tenantID uuid.UUID, code string) (*Account, error)
	
	// RefreshToken refreshes the access token
	RefreshToken(ctx context.Context, account *Account) error
	
	// Post publishes content to the platform
	Post(ctx context.Context, account *Account, post *Post) (string, error)
	
	// Delete removes a post from the platform
	Delete(ctx context.Context, account *Account, externalID string) error
	
	// GetMetrics retrieves engagement metrics for a post
	GetMetrics(ctx context.Context, account *Account, externalID string) (*PostMetrics, error)
}

// RegisterProvider registers a social media provider
func (s *Service) RegisterProvider(platform Platform, provider Provider) {
	s.providers[platform] = provider
}

// =============================================================================
// Account Management
// =============================================================================

// ConnectAccount initiates OAuth flow for a platform
func (s *Service) ConnectAccount(ctx context.Context, tenantID uuid.UUID, platform Platform) (string, error) {
	provider, ok := s.providers[platform]
	if !ok {
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}

	authURL, err := provider.Connect(ctx, tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to initiate OAuth: %w", err)
	}

	return authURL, nil
}

// HandleOAuthCallback processes OAuth callback
func (s *Service) HandleOAuthCallback(ctx context.Context, tenantID uuid.UUID, platform Platform, code string) (*Account, error) {
	provider, ok := s.providers[platform]
	if !ok {
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}

	account, err := provider.HandleCallback(ctx, tenantID, code)
	if err != nil {
		return nil, fmt.Errorf("failed to process OAuth callback: %w", err)
	}

	s.log.Infow("social account connected",
		"tenant_id", tenantID,
		"platform", platform,
		"username", account.Username,
	)

	return account, nil
}

// =============================================================================
// Content Publishing
// =============================================================================

// PublishPost publishes a post to social media
func (s *Service) PublishPost(ctx context.Context, account *Account, post *Post) error {
	provider, ok := s.providers[account.Platform]
	if !ok {
		return fmt.Errorf("unsupported platform: %s", account.Platform)
	}

	// Check if token needs refresh
	if time.Now().After(account.TokenExpiry.Add(-5 * time.Minute)) {
		if err := provider.RefreshToken(ctx, account); err != nil {
			return fmt.Errorf("failed to refresh token: %w", err)
		}
	}

	// Publish the post
	externalID, err := provider.Post(ctx, account, post)
	if err != nil {
		post.Status = PostStatusFailed
		return fmt.Errorf("failed to publish post: %w", err)
	}

	// Update post with external ID
	post.ExternalID = externalID
	post.Status = PostStatusPublished
	now := time.Now()
	post.PublishedAt = &now

	s.log.Infow("post published",
		"platform", account.Platform,
		"post_id", post.ID,
		"external_id", externalID,
	)

	return nil
}

// SchedulePost schedules a post for future publication
func (s *Service) SchedulePost(ctx context.Context, post *Post, scheduledAt time.Time) error {
	post.Status = PostStatusScheduled
	post.ScheduledAt = &scheduledAt
	
	s.log.Infow("post scheduled",
		"post_id", post.ID,
		"scheduled_at", scheduledAt,
	)

	return nil
}

// DeletePost deletes a published post
func (s *Service) DeletePost(ctx context.Context, account *Account, post *Post) error {
	if post.ExternalID == "" {
		return fmt.Errorf("post has no external ID")
	}

	provider, ok := s.providers[account.Platform]
	if !ok {
		return fmt.Errorf("unsupported platform: %s", account.Platform)
	}

	if err := provider.Delete(ctx, account, post.ExternalID); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	s.log.Infow("post deleted",
		"platform", account.Platform,
		"post_id", post.ID,
		"external_id", post.ExternalID,
	)

	return nil
}

// =============================================================================
// Content Pipeline
// =============================================================================

// GenerateContent uses an AI agent to generate social media content
func (s *Service) GenerateContent(ctx context.Context, pipeline *ContentPipeline) ([]*Post, error) {
	// This would integrate with the agent execution system
	// to generate content based on the pipeline configuration
	
	s.log.Infow("generating content",
		"pipeline_id", pipeline.ID,
		"agent_id", pipeline.AgentID,
		"platforms", pipeline.Platforms,
	)

	// Placeholder - actual implementation would:
	// 1. Brief the content agent with pipeline config
	// 2. Generate content for each platform
	// 3. Create draft posts
	// 4. Return posts for review or auto-publish

	return nil, nil
}

// =============================================================================
// Analytics
// =============================================================================

// GetPostMetrics retrieves metrics for a post
func (s *Service) GetPostMetrics(ctx context.Context, account *Account, post *Post) (*PostMetrics, error) {
	if post.ExternalID == "" {
		return nil, fmt.Errorf("post has no external ID")
	}

	provider, ok := s.providers[account.Platform]
	if !ok {
		return nil, fmt.Errorf("unsupported platform: %s", account.Platform)
	}

	metrics, err := provider.GetMetrics(ctx, account, post.ExternalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics: %w", err)
	}

	post.Metrics = *metrics
	return metrics, nil
}

// AggregateMetrics aggregates metrics across multiple posts
func (s *Service) AggregateMetrics(posts []*Post) PostMetrics {
	var aggregate PostMetrics
	for _, post := range posts {
		aggregate.Views += post.Metrics.Views
		aggregate.Likes += post.Metrics.Likes
		aggregate.Comments += post.Metrics.Comments
		aggregate.Shares += post.Metrics.Shares
		aggregate.Clicks += post.Metrics.Clicks
		aggregate.Impressions += post.Metrics.Impressions
	}
	return aggregate
}

