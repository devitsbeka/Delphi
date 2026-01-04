package execution

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/delphi-platform/delphi/backend/internal/models"
	"github.com/delphi-platform/delphi/backend/internal/providers"
	"github.com/delphi-platform/delphi/backend/pkg/logger"
)

// BriefingEngine handles the agent briefing/grooming phase
type BriefingEngine struct {
	log *logger.Logger
}

// NewBriefingEngine creates a new briefing engine
func NewBriefingEngine(log *logger.Logger) *BriefingEngine {
	return &BriefingEngine{log: log}
}

// BriefingContext contains context information for briefing
type BriefingContext struct {
	TenantContext    *TenantBriefing
	ProjectContext   *ProjectBriefing
	RecentActivity   *ActivityBriefing
	KnowledgeContext *KnowledgeBriefing
}

// TenantBriefing contains tenant-wide context
type TenantBriefing struct {
	TenantName     string
	BusinessCount  int
	ProjectCount   int
	TotalAgents    int
	ActiveAgents   int
	CustomRules    []string
}

// ProjectBriefing contains project-specific context
type ProjectBriefing struct {
	ProjectName    string
	Description    string
	Repositories   []string
	RecentCommits  []string
	OpenPRs        int
	ActiveBranches []string
}

// ActivityBriefing contains recent activity
type ActivityBriefing struct {
	RecentRuns        []RunSummary
	RecentConversations []ConversationSummary
	PendingTasks      []string
}

// RunSummary summarizes a recent run
type RunSummary struct {
	AgentName string
	Prompt    string
	Status    string
	Time      time.Time
}

// ConversationSummary summarizes a recent conversation
type ConversationSummary struct {
	Topic    string
	Summary  string
	Time     time.Time
}

// KnowledgeBriefing contains knowledge base context
type KnowledgeBriefing struct {
	RelevantDocuments []DocumentSummary
	RecentUpdates     []string
}

// DocumentSummary summarizes a relevant document
type DocumentSummary struct {
	Title   string
	Summary string
	Source  string
}

// BriefingResult contains the result of the briefing process
type BriefingResult struct {
	Success          bool
	EnhancedPrompt   string
	ContextSummary   string
	EstimatedTokens  int
	Duration         time.Duration
	Warnings         []string
}

// Brief performs the briefing process for an agent
func (e *BriefingEngine) Brief(ctx context.Context, agent *models.Agent, briefingContext *BriefingContext) (*BriefingResult, error) {
	start := time.Now()
	e.log.Infow("starting briefing", "agent_id", agent.ID, "depth", agent.Config.BriefingDepth)

	result := &BriefingResult{
		Success: true,
	}

	// Build the enhanced system prompt based on briefing depth
	var enhancedPrompt strings.Builder
	
	// Start with base system prompt
	enhancedPrompt.WriteString(agent.SystemPrompt)
	enhancedPrompt.WriteString("\n\n")

	// Add context based on depth
	switch agent.Config.BriefingDepth {
	case "quick":
		e.addQuickContext(&enhancedPrompt, briefingContext)
	case "full":
		e.addFullContext(&enhancedPrompt, briefingContext)
	default: // standard
		e.addStandardContext(&enhancedPrompt, briefingContext)
	}

	// Add universal guidelines
	e.addUniversalGuidelines(&enhancedPrompt, agent)

	result.EnhancedPrompt = enhancedPrompt.String()
	result.ContextSummary = e.generateContextSummary(briefingContext)
	result.EstimatedTokens = len(result.EnhancedPrompt) / 4 // Approximate
	result.Duration = time.Since(start)

	e.log.Infow("briefing complete", 
		"agent_id", agent.ID, 
		"duration_ms", result.Duration.Milliseconds(),
		"estimated_tokens", result.EstimatedTokens,
	)

	return result, nil
}

// addQuickContext adds minimal context for quick briefings
func (e *BriefingEngine) addQuickContext(b *strings.Builder, ctx *BriefingContext) {
	b.WriteString("## Current Context (Quick Briefing)\n\n")
	
	if ctx.TenantContext != nil {
		b.WriteString(fmt.Sprintf("Organization: %s\n", ctx.TenantContext.TenantName))
	}
	
	if ctx.ProjectContext != nil && ctx.ProjectContext.ProjectName != "" {
		b.WriteString(fmt.Sprintf("Project: %s\n", ctx.ProjectContext.ProjectName))
	}
	
	b.WriteString("\n")
}

// addStandardContext adds standard context for normal briefings
func (e *BriefingEngine) addStandardContext(b *strings.Builder, ctx *BriefingContext) {
	b.WriteString("## Current Context (Standard Briefing)\n\n")
	
	// Tenant context
	if ctx.TenantContext != nil {
		b.WriteString(fmt.Sprintf("### Organization: %s\n", ctx.TenantContext.TenantName))
		b.WriteString(fmt.Sprintf("- Businesses: %d\n", ctx.TenantContext.BusinessCount))
		b.WriteString(fmt.Sprintf("- Projects: %d\n", ctx.TenantContext.ProjectCount))
		b.WriteString(fmt.Sprintf("- Active Agents: %d/%d\n", ctx.TenantContext.ActiveAgents, ctx.TenantContext.TotalAgents))
		b.WriteString("\n")
	}
	
	// Project context
	if ctx.ProjectContext != nil && ctx.ProjectContext.ProjectName != "" {
		b.WriteString(fmt.Sprintf("### Current Project: %s\n", ctx.ProjectContext.ProjectName))
		if ctx.ProjectContext.Description != "" {
			b.WriteString(fmt.Sprintf("%s\n", ctx.ProjectContext.Description))
		}
		if len(ctx.ProjectContext.Repositories) > 0 {
			b.WriteString("Repositories:\n")
			for _, repo := range ctx.ProjectContext.Repositories {
				b.WriteString(fmt.Sprintf("- %s\n", repo))
			}
		}
		b.WriteString("\n")
	}
	
	// Recent activity
	if ctx.RecentActivity != nil && len(ctx.RecentActivity.RecentRuns) > 0 {
		b.WriteString("### Recent Activity\n")
		for _, run := range ctx.RecentActivity.RecentRuns[:min(3, len(ctx.RecentActivity.RecentRuns))] {
			b.WriteString(fmt.Sprintf("- [%s] %s: %s\n", run.Status, run.AgentName, truncate(run.Prompt, 50)))
		}
		b.WriteString("\n")
	}
}

// addFullContext adds comprehensive context for full briefings
func (e *BriefingEngine) addFullContext(b *strings.Builder, ctx *BriefingContext) {
	b.WriteString("## Current Context (Full Briefing)\n\n")
	
	// Start with standard context
	e.addStandardContext(b, ctx)
	
	// Add knowledge context
	if ctx.KnowledgeContext != nil {
		if len(ctx.KnowledgeContext.RelevantDocuments) > 0 {
			b.WriteString("### Relevant Knowledge\n")
			for _, doc := range ctx.KnowledgeContext.RelevantDocuments[:min(5, len(ctx.KnowledgeContext.RelevantDocuments))] {
				b.WriteString(fmt.Sprintf("**%s** (%s)\n", doc.Title, doc.Source))
				b.WriteString(fmt.Sprintf("%s\n\n", doc.Summary))
			}
		}
		
		if len(ctx.KnowledgeContext.RecentUpdates) > 0 {
			b.WriteString("### Recent Updates\n")
			for _, update := range ctx.KnowledgeContext.RecentUpdates[:min(5, len(ctx.KnowledgeContext.RecentUpdates))] {
				b.WriteString(fmt.Sprintf("- %s\n", update))
			}
			b.WriteString("\n")
		}
	}
	
	// Add project details
	if ctx.ProjectContext != nil {
		if len(ctx.ProjectContext.RecentCommits) > 0 {
			b.WriteString("### Recent Commits\n")
			for _, commit := range ctx.ProjectContext.RecentCommits[:min(5, len(ctx.ProjectContext.RecentCommits))] {
				b.WriteString(fmt.Sprintf("- %s\n", commit))
			}
			b.WriteString("\n")
		}
		
		if ctx.ProjectContext.OpenPRs > 0 {
			b.WriteString(fmt.Sprintf("Open PRs: %d\n\n", ctx.ProjectContext.OpenPRs))
		}
	}
	
	// Add pending tasks
	if ctx.RecentActivity != nil && len(ctx.RecentActivity.PendingTasks) > 0 {
		b.WriteString("### Pending Tasks\n")
		for _, task := range ctx.RecentActivity.PendingTasks {
			b.WriteString(fmt.Sprintf("- [ ] %s\n", task))
		}
		b.WriteString("\n")
	}
}

// addUniversalGuidelines adds guidelines that apply to all agents
func (e *BriefingEngine) addUniversalGuidelines(b *strings.Builder, agent *models.Agent) {
	b.WriteString("## Guidelines\n\n")
	
	// Type-specific guidelines
	switch agent.Type {
	case models.AgentTypeCoding:
		b.WriteString("- Always commit to the dev or staging branch, never directly to main\n")
		b.WriteString("- Write clear, descriptive commit messages\n")
		b.WriteString("- Follow the existing code style and conventions\n")
		b.WriteString("- Add appropriate tests for new functionality\n")
		b.WriteString("- Create a PR with a clear description when done\n")
	case models.AgentTypeBusiness:
		b.WriteString("- Provide data-driven insights with sources when possible\n")
		b.WriteString("- Consider both short-term and long-term implications\n")
		b.WriteString("- Be specific and actionable in recommendations\n")
	case models.AgentTypeAccounting:
		b.WriteString("- Ensure accuracy in all calculations\n")
		b.WriteString("- Follow standard accounting practices\n")
		b.WriteString("- Maintain clear audit trails\n")
	case models.AgentTypeMarketing:
		b.WriteString("- Align content with brand voice and guidelines\n")
		b.WriteString("- Consider the target audience for all content\n")
		b.WriteString("- Optimize for engagement and conversions\n")
	case models.AgentTypeAssistant:
		b.WriteString("- Be concise and professional in communications\n")
		b.WriteString("- Verify important details before taking action\n")
		b.WriteString("- Prioritize urgent matters appropriately\n")
	}
	
	b.WriteString("\n")
}

// generateContextSummary creates a brief summary of the context
func (e *BriefingEngine) generateContextSummary(ctx *BriefingContext) string {
	var parts []string
	
	if ctx.TenantContext != nil {
		parts = append(parts, ctx.TenantContext.TenantName)
	}
	
	if ctx.ProjectContext != nil && ctx.ProjectContext.ProjectName != "" {
		parts = append(parts, ctx.ProjectContext.ProjectName)
	}
	
	if ctx.RecentActivity != nil && len(ctx.RecentActivity.RecentRuns) > 0 {
		parts = append(parts, fmt.Sprintf("%d recent runs", len(ctx.RecentActivity.RecentRuns)))
	}
	
	if len(parts) == 0 {
		return "No context loaded"
	}
	
	return strings.Join(parts, " | ")
}

// BuildCompletionRequest builds a completion request with briefing context
func (e *BriefingEngine) BuildCompletionRequest(agent *models.Agent, briefingResult *BriefingResult, userPrompt string) *providers.CompletionRequest {
	return providers.NewRequestBuilder(agent.Model).
		WithSystemPrompt(briefingResult.EnhancedPrompt).
		WithUserMessage(userPrompt).
		WithTemperature(agent.Config.Temperature).
		WithMaxTokens(agent.Config.MaxTokens).
		Build()
}

// Helper functions

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

