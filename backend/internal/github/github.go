package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/models"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
	"github.com/google/uuid"
)

const githubAPIURL = "https://api.github.com"

// Client handles GitHub API operations
type Client struct {
	httpClient *http.Client
	log        *logger.Logger
}

// NewClient creates a new GitHub client
func NewClient(log *logger.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		log: log,
	}
}

// =============================================================================
// Authentication
// =============================================================================

// Installation represents a GitHub App installation
type Installation struct {
	ID          int64  `json:"id"`
	Account     Account `json:"account"`
	AccessToken string `json:"-"` // Not from API, set after authentication
}

// Account represents a GitHub account
type Account struct {
	ID    int64  `json:"id"`
	Login string `json:"login"`
	Type  string `json:"type"` // User or Organization
}

// GetInstallationToken gets an access token for a GitHub App installation
func (c *Client) GetInstallationToken(ctx context.Context, appID string, privateKey []byte, installationID int64) (string, error) {
	// In production, this would:
	// 1. Generate a JWT using the app private key
	// 2. Exchange it for an installation access token
	// For now, return a placeholder
	c.log.Infow("getting installation token", "installation_id", installationID)
	return "", fmt.Errorf("GitHub App authentication not implemented - configure via OAuth")
}

// =============================================================================
// Repository Operations
// =============================================================================

// Repository represents a GitHub repository
type Repository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	FullName      string `json:"full_name"`
	Private       bool   `json:"private"`
	Description   string `json:"description"`
	HTMLURL       string `json:"html_url"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

// ListRepositories lists repositories accessible to the authenticated user
func (c *Client) ListRepositories(ctx context.Context, token string) ([]Repository, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", githubAPIURL+"/user/repos?per_page=100", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var repos []Repository
	if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
		return nil, err
	}

	return repos, nil
}

// GetRepository gets a specific repository
func (c *Client) GetRepository(ctx context.Context, token, owner, repo string) (*Repository, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", githubAPIURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, err
	}

	return &repository, nil
}

// =============================================================================
// Branch Operations
// =============================================================================

// Branch represents a GitHub branch
type Branch struct {
	Name   string `json:"name"`
	Commit struct {
		SHA string `json:"sha"`
	} `json:"commit"`
	Protected bool `json:"protected"`
}

// ListBranches lists branches in a repository
func (c *Client) ListBranches(ctx context.Context, token, owner, repo string) ([]Branch, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/branches", githubAPIURL, owner, repo)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var branches []Branch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}

	return branches, nil
}

// CreateBranch creates a new branch
func (c *Client) CreateBranch(ctx context.Context, token, owner, repo, branchName, baseSHA string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/git/refs", githubAPIURL, owner, repo)
	
	body := map[string]string{
		"ref": "refs/heads/" + branchName,
		"sha": baseSHA,
	}
	
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// =============================================================================
// File Operations
// =============================================================================

// FileContent represents file content from GitHub
type FileContent struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	SHA         string `json:"sha"`
	Size        int    `json:"size"`
	Content     string `json:"content"`
	Encoding    string `json:"encoding"`
	DownloadURL string `json:"download_url"`
}

// GetFileContent gets the content of a file
func (c *Client) GetFileContent(ctx context.Context, token, owner, repo, path, ref string) (*FileContent, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", githubAPIURL, owner, repo, path)
	if ref != "" {
		url += "?ref=" + ref
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var content FileContent
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, err
	}

	return &content, nil
}

// CreateOrUpdateFile creates or updates a file in the repository
func (c *Client) CreateOrUpdateFile(ctx context.Context, token, owner, repo, path, message, content, branch string, sha *string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/contents/%s", githubAPIURL, owner, repo, path)
	
	body := map[string]interface{}{
		"message": message,
		"content": content, // Must be base64 encoded
		"branch":  branch,
	}
	
	if sha != nil {
		body["sha"] = *sha
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(jsonBody))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// =============================================================================
// Pull Request Operations
// =============================================================================

// PullRequest represents a GitHub pull request
type PullRequest struct {
	ID        int64  `json:"id"`
	Number    int    `json:"number"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	State     string `json:"state"`
	HTMLURL   string `json:"html_url"`
	Head      struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"head"`
	Base struct {
		Ref string `json:"ref"`
		SHA string `json:"sha"`
	} `json:"base"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreatePullRequest creates a new pull request
func (c *Client) CreatePullRequest(ctx context.Context, token, owner, repo, title, body, head, base string) (*PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls", githubAPIURL, owner, repo)
	
	reqBody := map[string]string{
		"title": title,
		"body":  body,
		"head":  head,
		"base":  base,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(respBody))
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, err
	}

	return &pr, nil
}

// ListPullRequests lists pull requests
func (c *Client) ListPullRequests(ctx context.Context, token, owner, repo, state string) ([]PullRequest, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=%s", githubAPIURL, owner, repo, state)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: %d", resp.StatusCode)
	}

	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, err
	}

	return prs, nil
}

// =============================================================================
// Webhook Handling
// =============================================================================

// WebhookPayload represents a GitHub webhook payload
type WebhookPayload struct {
	Action       string          `json:"action"`
	Repository   Repository      `json:"repository"`
	Sender       Account         `json:"sender"`
	Installation *Installation   `json:"installation,omitempty"`
	PullRequest  *PullRequest    `json:"pull_request,omitempty"`
}

// WebhookHandler handles GitHub webhooks
type WebhookHandler struct {
	secret string
	log    *logger.Logger
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(secret string, log *logger.Logger) *WebhookHandler {
	return &WebhookHandler{
		secret: secret,
		log:    log,
	}
}

// HandleWebhook processes a GitHub webhook
func (h *WebhookHandler) HandleWebhook(eventType string, payload []byte) error {
	h.log.Infow("received webhook", "event", eventType)

	var webhookPayload WebhookPayload
	if err := json.Unmarshal(payload, &webhookPayload); err != nil {
		return fmt.Errorf("failed to parse payload: %w", err)
	}

	switch eventType {
	case "push":
		return h.handlePush(webhookPayload)
	case "pull_request":
		return h.handlePullRequest(webhookPayload)
	case "issues":
		return h.handleIssue(webhookPayload)
	case "issue_comment":
		return h.handleIssueComment(webhookPayload)
	default:
		h.log.Debugw("ignoring unhandled event", "event", eventType)
	}

	return nil
}

func (h *WebhookHandler) handlePush(payload WebhookPayload) error {
	h.log.Infow("processing push event", "repo", payload.Repository.FullName)
	// Trigger knowledge base update for the repository
	return nil
}

func (h *WebhookHandler) handlePullRequest(payload WebhookPayload) error {
	h.log.Infow("processing pull request event", 
		"repo", payload.Repository.FullName, 
		"action", payload.Action,
		"pr", payload.PullRequest.Number,
	)
	// Track PR status for agent-created PRs
	return nil
}

func (h *WebhookHandler) handleIssue(payload WebhookPayload) error {
	h.log.Infow("processing issue event", 
		"repo", payload.Repository.FullName, 
		"action", payload.Action,
	)
	// Could trigger agent to handle assigned issues
	return nil
}

func (h *WebhookHandler) handleIssueComment(payload WebhookPayload) error {
	h.log.Infow("processing issue comment event", 
		"repo", payload.Repository.FullName, 
		"action", payload.Action,
	)
	// Could enable human-agent interaction via PR comments
	return nil
}

// =============================================================================
// Repository Manager
// =============================================================================

// RepositoryManager manages repository connections
type RepositoryManager struct {
	client *Client
	log    *logger.Logger
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(log *logger.Logger) *RepositoryManager {
	return &RepositoryManager{
		client: NewClient(log),
		log:    log,
	}
}

// ConnectRepository connects a repository to the platform
func (m *RepositoryManager) ConnectRepository(ctx context.Context, token string, repoURL string) (*models.Repository, error) {
	// Parse owner and repo from URL
	// In production, properly parse the URL
	
	repo, err := m.client.GetRepository(ctx, token, "owner", "repo")
	if err != nil {
		return nil, err
	}

	return &models.Repository{
		ID:            uuid.New(),
		Name:          repo.Name,
		FullName:      repo.FullName,
		URL:           repo.HTMLURL,
		DefaultBranch: repo.DefaultBranch,
		IsPrivate:     repo.Private,
		CreatedAt:     time.Now(),
	}, nil
}

