# Delphi User Guide

Welcome to Delphi, your AI Agent Command Center. This guide will help you get started and make the most of the platform.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Dashboard Overview](#dashboard-overview)
3. [Creating Oracles (Agents)](#creating-oracles-agents)
4. [Executing Tasks](#executing-tasks)
5. [Knowledge Base](#knowledge-base)
6. [Repository Management](#repository-management)
7. [Business Management](#business-management)
8. [Cost Tracking](#cost-tracking)
9. [Settings & Configuration](#settings--configuration)
10. [Best Practices](#best-practices)

---

## Getting Started

### 1. Create Your Account

1. Visit [delphi.dev](https://delphi.dev)
2. Click "Sign Up"
3. Enter your email and create a password
4. Verify your email address

### 2. Configure API Keys

Before using AI oracles, you need to add your API keys:

1. Go to **Settings** → **API Keys**
2. Click **Configure** next to your preferred provider
3. Enter your API key (e.g., OpenAI: `sk-...`)
4. Click **Save**

Supported providers:
- **OpenAI**: GPT-4, GPT-4 Turbo, GPT-3.5
- **Anthropic**: Claude 3 Opus, Sonnet, Haiku
- **Google AI**: Gemini Pro
- **Ollama**: Any locally hosted model

### 3. Connect Your First Repository

1. Go to **Repositories**
2. Click **+ Connect Repository**
3. Authorize Delphi with GitHub
4. Select a repository to connect
5. Configure branch strategy:
   - **Main branch**: Production code (e.g., `main`)
   - **Staging branch**: Review before production (e.g., `staging`)
   - **Dev branch**: Where oracles commit (e.g., `dev`)

---

## Dashboard Overview

The dashboard is your command center, showing:

### Metrics Cards
- **Active Oracles**: Currently running agents
- **Today's Executions**: Tasks completed today
- **Today's Cost**: API spending for the day
- **Monthly Budget**: Budget utilization

### Activity Chart
Shows execution volume and cost over time. Toggle between 7, 14, and 30-day views.

### Cost Breakdown
Pie chart showing spending by AI provider.

### Recent Executions
Live feed of oracle activity with status, duration, and cost.

### Business Overview
Quick stats for each connected business.

---

## Creating Oracles (Agents)

Oracles are AI agents specialized for specific tasks.

### Create an Oracle

1. Go to **Oracles** → **+ Create Oracle**
2. Fill in the details:

| Field | Description | Example |
|-------|-------------|---------|
| Name | Descriptive name | "Code Review Oracle" |
| Description | What it does | "Reviews PRs for code quality" |
| Purpose | Category | Coding, Content, DevOps, Analysis |
| Model Provider | AI provider | OpenAI, Anthropic, Google |
| Model | Specific model | gpt-4-turbo, claude-3-opus |
| System Prompt | Instructions | "You are an expert code reviewer..." |
| Goal | Main objective | "Identify bugs and suggest improvements" |
| Business | Associated business | Mobile Game Studio |

### System Prompt Best Practices

A good system prompt includes:
- **Role definition**: Who the oracle is
- **Capabilities**: What it can do
- **Constraints**: What it should avoid
- **Output format**: How to structure responses

Example:
```
You are Code Review Oracle, an expert software engineer specializing in 
TypeScript and React applications.

Your responsibilities:
- Review code for bugs, security issues, and best practices
- Suggest performance improvements
- Ensure consistent coding style
- Be constructive and educational

Output Format:
- Start with a summary (1-2 sentences)
- List issues by severity (Critical, High, Medium, Low)
- Provide specific code suggestions with examples
- End with positive feedback on what's done well

Constraints:
- Don't rewrite entire files unless necessary
- Focus on the most impactful improvements
- Respect existing coding conventions
```

### Oracle Templates

Use pre-built templates for common use cases:
- **Code Reviewer**: PR review and code quality
- **Bug Fixer**: Automatic bug resolution
- **Documentation Writer**: README and API docs
- **Content Creator**: Blog posts and marketing
- **Data Analyst**: Analytics and reporting
- **DevOps Agent**: CI/CD and infrastructure

---

## Executing Tasks

### Launch a Single Oracle

1. Go to **Execute**
2. Select an oracle from the list
3. Enter your task description
4. (Optional) Add context or constraints
5. Click **Launch Oracle**

### Parallel Execution

Launch multiple oracles simultaneously:

1. Select multiple oracles (checkboxes)
2. Enter a task they all work on
3. Click **Launch X Oracles**

Each oracle receives the same task but approaches it based on their specialization.

### Task Examples

**For Code Review Oracle:**
```
Review PR #42 in puzzle-blast repository. Focus on:
- Memory management in the game loop
- React hooks usage
- TypeScript type safety
```

**For Content Writer:**
```
Write a blog post about our new puzzle game launch.
Target audience: casual mobile gamers
Tone: Friendly, exciting
Length: 800-1000 words
Include: Feature highlights, download CTA
```

**For DevOps Agent:**
```
Set up GitHub Actions workflow for the crash-royale repository.
Requirements:
- Run tests on every PR
- Build Docker image on main branch
- Deploy to staging on successful build
```

### Monitoring Executions

The execution log shows real-time progress:
- `[SYS]` - System messages
- `[AGT]` - Oracle activity
- `[TL]` - Tool usage (knowledge base, repository access)
- `[OK]` - Success messages
- `[ERR]` - Errors

---

## Knowledge Base

The knowledge base provides oracles with context about your codebase and documentation.

### Types of Knowledge Bases

1. **Repository**: Automatically indexed from connected repos
2. **Documents**: Uploaded PDFs, docs, and text files
3. **Custom**: API-ingested content

### Uploading Documents

1. Go to **Knowledge Base**
2. Click **Upload Documents**
3. Select a target knowledge base
4. Drag and drop files or click to browse
5. Supported formats: PDF, DOCX, MD, TXT, code files

### Querying the Knowledge Base

Test queries before oracle execution:

1. Go to **Knowledge Base**
2. Enter a question in the search box
3. View retrieved context chunks
4. Adjust your oracle's system prompt if needed

### Indexing Status

Check if repositories are fully indexed:
- ✓ **Indexed**: Ready for use
- ◐ **Indexing**: In progress
- ✗ **Failed**: Check connection

---

## Repository Management

### Branch Strategy

Delphi uses a three-branch strategy:

```
main (production)
  ↑ merge
staging (review)
  ↑ merge
dev (oracle commits)
```

1. Oracles commit to the **dev** branch
2. You review and merge to **staging**
3. After testing, merge to **main**

### Syncing Repositories

Repositories sync automatically, but you can force a sync:

1. Go to **Repositories**
2. Click on a repository
3. Click **Sync Now**

### Viewing Oracle Activity

See what oracles have done:
- Commits created
- Pull requests opened
- Code reviews submitted

---

## Business Management

Organize your work by business unit.

### Creating a Business

1. Go to **Businesses** → **+ Add Business**
2. Enter details:
   - Name
   - Description
   - Industry
   - Monthly AI Budget

### Assigning Resources

Associate oracles and repositories with businesses:
- Oracles can belong to one business
- Repositories can belong to one business
- Costs are tracked per business

### Budget Controls

Set spending limits per business:
1. Click on a business
2. Set **Monthly Budget**
3. Set **Alert Threshold** (e.g., 80%)
4. Enable **Hard Limit** to stop executions when exceeded

---

## Cost Tracking

### Understanding Costs

Costs are based on:
- **Tokens Used**: Input + output tokens
- **Provider Rates**: Each AI provider has different pricing
- **Execution Count**: Number of tasks run

### Cost Dashboard

The **Costs** page shows:
- Total spend over time
- Cost by provider
- Cost by oracle
- Cost by business
- Token usage breakdown

### Optimizing Costs

1. **Choose the right model**: GPT-3.5 is cheaper than GPT-4
2. **Write efficient prompts**: Be concise but clear
3. **Use knowledge base**: Reduces token usage by providing targeted context
4. **Set budgets**: Prevent runaway costs
5. **Monitor regularly**: Check the cost dashboard weekly

---

## Settings & Configuration

### Notification Preferences

Configure how you receive updates:
- **Email**: Execution failures, budget alerts
- **Slack**: Real-time execution updates
- **Discord**: Team notifications

### Security

- **Two-Factor Authentication**: Enable for added security
- **Active Sessions**: View and revoke sessions
- **Audit Log**: Review security events

### Billing

Manage your subscription:
- View current plan
- Upgrade/downgrade
- Download invoices
- Update payment method

---

## Best Practices

### 1. Start Small
Begin with one oracle and simple tasks. Expand as you learn.

### 2. Clear System Prompts
The better your instructions, the better the results.

### 3. Use Knowledge Base
Index your documentation for better context.

### 4. Review Before Merge
Always review oracle commits before merging to production.

### 5. Set Budgets
Prevent unexpected costs with spending limits.

### 6. Monitor Performance
Track success rates and adjust prompts accordingly.

### 7. Iterate
Refine your oracles based on results.

---

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `⌘ + K` | Quick search |
| `⌘ + N` | New oracle |
| `⌘ + E` | Execute task |
| `⌘ + /` | Show shortcuts |
| `Esc` | Close modal |

---

## Getting Help

- **Documentation**: [docs.delphi.dev](https://docs.delphi.dev)
- **API Reference**: [docs.delphi.dev/api](https://docs.delphi.dev/api)
- **Discord Community**: [discord.gg/delphi](https://discord.gg/delphi)
- **Email Support**: support@delphi.dev

---

<p align="center">
  Happy orchestrating! 🔮
</p>

