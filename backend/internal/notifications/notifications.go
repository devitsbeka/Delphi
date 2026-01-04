package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"time"

	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
)

// Service handles notifications across multiple channels
type Service struct {
	emailConfig  *EmailConfig
	slackConfig  *SlackConfig
	discordConfig *DiscordConfig
	httpClient   *http.Client
	log          *logger.Logger
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// SlackConfig holds Slack configuration
type SlackConfig struct {
	WebhookURL string
}

// DiscordConfig holds Discord configuration
type DiscordConfig struct {
	WebhookURL string
}

// NewService creates a new notification service
func NewService(emailConfig *EmailConfig, slackConfig *SlackConfig, discordConfig *DiscordConfig, log *logger.Logger) *Service {
	return &Service{
		emailConfig:   emailConfig,
		slackConfig:   slackConfig,
		discordConfig: discordConfig,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		log: log,
	}
}

// =============================================================================
// Notification Types
// =============================================================================

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationExecutionComplete NotificationType = "execution_complete"
	NotificationExecutionFailed   NotificationType = "execution_failed"
	NotificationBudgetAlert       NotificationType = "budget_alert"
	NotificationBudgetExceeded    NotificationType = "budget_exceeded"
	NotificationAgentError        NotificationType = "agent_error"
	NotificationPRCreated         NotificationType = "pr_created"
	NotificationWeeklyDigest      NotificationType = "weekly_digest"
)

// NotificationChannel represents a notification channel
type NotificationChannel string

const (
	ChannelEmail   NotificationChannel = "email"
	ChannelSlack   NotificationChannel = "slack"
	ChannelDiscord NotificationChannel = "discord"
	ChannelPush    NotificationChannel = "push"
)

// Notification represents a notification to send
type Notification struct {
	ID        uuid.UUID
	TenantID  uuid.UUID
	UserID    *uuid.UUID
	Type      NotificationType
	Title     string
	Message   string
	Data      map[string]interface{}
	Channels  []NotificationChannel
	CreatedAt time.Time
}

// =============================================================================
// Send Notifications
// =============================================================================

// Send sends a notification through all specified channels
func (s *Service) Send(ctx context.Context, notification *Notification) error {
	s.log.Infow("sending notification",
		"type", notification.Type,
		"channels", notification.Channels,
	)

	var errors []error

	for _, channel := range notification.Channels {
		var err error
		switch channel {
		case ChannelEmail:
			err = s.sendEmail(ctx, notification)
		case ChannelSlack:
			err = s.sendSlack(ctx, notification)
		case ChannelDiscord:
			err = s.sendDiscord(ctx, notification)
		case ChannelPush:
			err = s.sendPush(ctx, notification)
		}

		if err != nil {
			s.log.Warnw("failed to send notification", "channel", channel, "error", err)
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to send to %d channels", len(errors))
	}

	return nil
}

// =============================================================================
// Email
// =============================================================================

func (s *Service) sendEmail(ctx context.Context, notification *Notification) error {
	if s.emailConfig == nil || s.emailConfig.Host == "" {
		return fmt.Errorf("email not configured")
	}

	// Get recipient email from notification.Data or lookup by UserID
	to, ok := notification.Data["email"].(string)
	if !ok || to == "" {
		return fmt.Errorf("no email recipient specified")
	}

	// Build email
	subject := notification.Title
	body := notification.Message

	// Simple HTML template
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: 'JetBrains Mono', monospace; background-color: #0a0a0f; color: #e8e8ec; padding: 20px; }
        .container { max-width: 600px; margin: 0 auto; }
        .header { color: #4a9eff; font-size: 24px; margin-bottom: 20px; }
        .content { background-color: #12121a; border: 1px solid #2a2a3a; border-radius: 8px; padding: 20px; }
        .footer { color: #606070; font-size: 12px; margin-top: 20px; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">Delphi</div>
        <div class="content">
            <h2>%s</h2>
            <p>%s</p>
        </div>
        <div class="footer">Delphi AI Command Center</div>
    </div>
</body>
</html>
`, subject, body)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", s.emailConfig.From, to, subject, htmlBody)

	auth := smtp.PlainAuth("", s.emailConfig.User, s.emailConfig.Password, s.emailConfig.Host)
	addr := fmt.Sprintf("%s:%d", s.emailConfig.Host, s.emailConfig.Port)

	err := smtp.SendMail(addr, auth, s.emailConfig.From, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.log.Infow("email sent", "to", to, "subject", subject)
	return nil
}

// =============================================================================
// Slack
// =============================================================================

type slackMessage struct {
	Text        string            `json:"text,omitempty"`
	Blocks      []slackBlock      `json:"blocks,omitempty"`
	Attachments []slackAttachment `json:"attachments,omitempty"`
}

type slackBlock struct {
	Type string      `json:"type"`
	Text *slackText  `json:"text,omitempty"`
}

type slackText struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type slackAttachment struct {
	Color  string `json:"color"`
	Title  string `json:"title"`
	Text   string `json:"text"`
	Footer string `json:"footer"`
}

func (s *Service) sendSlack(ctx context.Context, notification *Notification) error {
	if s.slackConfig == nil || s.slackConfig.WebhookURL == "" {
		return fmt.Errorf("Slack not configured")
	}

	// Determine color based on notification type
	color := "#4a9eff" // blue
	switch notification.Type {
	case NotificationExecutionComplete, NotificationPRCreated:
		color = "#00d68f" // green
	case NotificationExecutionFailed, NotificationAgentError:
		color = "#ff4757" // red
	case NotificationBudgetAlert, NotificationBudgetExceeded:
		color = "#ffaa00" // yellow
	}

	msg := slackMessage{
		Attachments: []slackAttachment{
			{
				Color:  color,
				Title:  notification.Title,
				Text:   notification.Message,
				Footer: "Delphi AI Command Center",
			},
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.slackConfig.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack webhook returned %d", resp.StatusCode)
	}

	s.log.Infow("Slack notification sent", "title", notification.Title)
	return nil
}

// =============================================================================
// Discord
// =============================================================================

type discordMessage struct {
	Content string         `json:"content,omitempty"`
	Embeds  []discordEmbed `json:"embeds,omitempty"`
}

type discordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
	Footer      *discordFooter `json:"footer,omitempty"`
	Timestamp   string `json:"timestamp,omitempty"`
}

type discordFooter struct {
	Text string `json:"text"`
}

func (s *Service) sendDiscord(ctx context.Context, notification *Notification) error {
	if s.discordConfig == nil || s.discordConfig.WebhookURL == "" {
		return fmt.Errorf("Discord not configured")
	}

	// Determine color based on notification type (Discord uses decimal colors)
	color := 4889855 // blue
	switch notification.Type {
	case NotificationExecutionComplete, NotificationPRCreated:
		color = 54927 // green
	case NotificationExecutionFailed, NotificationAgentError:
		color = 16729943 // red
	case NotificationBudgetAlert, NotificationBudgetExceeded:
		color = 16755200 // yellow
	}

	msg := discordMessage{
		Embeds: []discordEmbed{
			{
				Title:       notification.Title,
				Description: notification.Message,
				Color:       color,
				Footer: &discordFooter{
					Text: "Delphi AI Command Center",
				},
				Timestamp: time.Now().Format(time.RFC3339),
			},
		},
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.discordConfig.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Discord webhook returned %d", resp.StatusCode)
	}

	s.log.Infow("Discord notification sent", "title", notification.Title)
	return nil
}

// =============================================================================
// Push (placeholder for web push or mobile push)
// =============================================================================

func (s *Service) sendPush(ctx context.Context, notification *Notification) error {
	// In production, implement web push via Service Workers
	// or mobile push via FCM/APNs
	s.log.Debugw("push notification not implemented", "title", notification.Title)
	return nil
}

// =============================================================================
// Notification Templates
// =============================================================================

// ExecutionCompleteNotification creates a notification for completed execution
func ExecutionCompleteNotification(tenantID uuid.UUID, agentName string, runID uuid.UUID, duration time.Duration) *Notification {
	return &Notification{
		ID:       uuid.New(),
		TenantID: tenantID,
		Type:     NotificationExecutionComplete,
		Title:    fmt.Sprintf("Oracle '%s' completed task", agentName),
		Message:  fmt.Sprintf("Execution %s completed in %s", runID.String()[:8], duration.Round(time.Second)),
		Data: map[string]interface{}{
			"agent_name": agentName,
			"run_id":     runID.String(),
			"duration":   duration.String(),
		},
		Channels:  []NotificationChannel{ChannelSlack, ChannelDiscord},
		CreatedAt: time.Now(),
	}
}

// ExecutionFailedNotification creates a notification for failed execution
func ExecutionFailedNotification(tenantID uuid.UUID, agentName string, runID uuid.UUID, errorMsg string) *Notification {
	return &Notification{
		ID:       uuid.New(),
		TenantID: tenantID,
		Type:     NotificationExecutionFailed,
		Title:    fmt.Sprintf("Oracle '%s' failed", agentName),
		Message:  fmt.Sprintf("Execution %s failed: %s", runID.String()[:8], errorMsg),
		Data: map[string]interface{}{
			"agent_name": agentName,
			"run_id":     runID.String(),
			"error":      errorMsg,
		},
		Channels:  []NotificationChannel{ChannelSlack, ChannelDiscord, ChannelEmail},
		CreatedAt: time.Now(),
	}
}

// BudgetAlertNotification creates a notification for budget alerts
func BudgetAlertNotification(tenantID uuid.UUID, spent float64, limit float64, percentage float64) *Notification {
	return &Notification{
		ID:       uuid.New(),
		TenantID: tenantID,
		Type:     NotificationBudgetAlert,
		Title:    "Budget Alert",
		Message:  fmt.Sprintf("You've used %.0f%% of your budget ($%.2f of $%.2f)", percentage, spent, limit),
		Data: map[string]interface{}{
			"spent":      spent,
			"limit":      limit,
			"percentage": percentage,
		},
		Channels:  []NotificationChannel{ChannelSlack, ChannelEmail},
		CreatedAt: time.Now(),
	}
}

// PRCreatedNotification creates a notification for PR creation
func PRCreatedNotification(tenantID uuid.UUID, agentName, repoName string, prNumber int, prURL string) *Notification {
	return &Notification{
		ID:       uuid.New(),
		TenantID: tenantID,
		Type:     NotificationPRCreated,
		Title:    fmt.Sprintf("Oracle '%s' created PR #%d", agentName, prNumber),
		Message:  fmt.Sprintf("New pull request on %s: %s", repoName, prURL),
		Data: map[string]interface{}{
			"agent_name": agentName,
			"repo_name":  repoName,
			"pr_number":  prNumber,
			"pr_url":     prURL,
		},
		Channels:  []NotificationChannel{ChannelSlack, ChannelDiscord},
		CreatedAt: time.Now(),
	}
}

