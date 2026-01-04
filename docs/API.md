# Delphi API Documentation

## Overview

The Delphi API provides a comprehensive interface for managing AI agents (Oracles), repositories, knowledge bases, and more. All API endpoints require authentication via JWT tokens.

## Base URL

```
https://api.delphi.dev/v1
```

## Authentication

Include the JWT token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### Obtain Token

```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "your-password"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "name": "John Doe"
  }
}
```

---

## Agents (Oracles)

### List Agents

```http
GET /agents
```

Query Parameters:
- `status` - Filter by status (ready, executing, paused, error)
- `purpose` - Filter by purpose (coding, content, devops, analysis, support)
- `business_id` - Filter by business

Response:
```json
{
  "agents": [
    {
      "id": "uuid",
      "name": "Code Review Oracle",
      "description": "Automated code review",
      "purpose": "coding",
      "status": "ready",
      "model_provider": "openai",
      "model": "gpt-4-turbo",
      "created_at": "2025-01-04T10:00:00Z"
    }
  ],
  "total": 1
}
```

### Create Agent

```http
POST /agents
Content-Type: application/json

{
  "name": "Code Review Oracle",
  "description": "Automated code review for PRs",
  "purpose": "coding",
  "model_provider": "openai",
  "model": "gpt-4-turbo",
  "system_prompt": "You are an expert code reviewer...",
  "goal": "Review code for quality and best practices",
  "business_id": "uuid"
}
```

### Get Agent

```http
GET /agents/:id
```

### Update Agent

```http
PUT /agents/:id
Content-Type: application/json

{
  "name": "Updated Agent Name",
  "system_prompt": "Updated system prompt..."
}
```

### Delete Agent

```http
DELETE /agents/:id
```

### Execute Agent

```http
POST /agents/:id/execute
Content-Type: application/json

{
  "task": "Review the latest PR in the puzzle-blast repository",
  "context": {
    "repository_id": "uuid",
    "pr_number": 42
  }
}
```

Response:
```json
{
  "execution_id": "uuid",
  "status": "running",
  "started_at": "2025-01-04T10:00:00Z"
}
```

### Get Execution Status

```http
GET /executions/:id
```

Response:
```json
{
  "id": "uuid",
  "agent_id": "uuid",
  "status": "completed",
  "task": "Review the latest PR...",
  "output": "PR review completed...",
  "tokens_used": 1500,
  "cost": 0.045,
  "started_at": "2025-01-04T10:00:00Z",
  "completed_at": "2025-01-04T10:02:30Z"
}
```

---

## Repositories

### List Repositories

```http
GET /repositories
```

### Connect Repository

```http
POST /repositories
Content-Type: application/json

{
  "owner": "organization",
  "name": "repository-name",
  "default_branch": "main",
  "dev_branch": "dev",
  "staging_branch": "staging",
  "business_id": "uuid"
}
```

### Sync Repository

```http
POST /repositories/:id/sync
```

### Get Repository

```http
GET /repositories/:id
```

---

## Knowledge Base

### List Knowledge Bases

```http
GET /knowledge-bases
```

### Create Knowledge Base

```http
POST /knowledge-bases
Content-Type: application/json

{
  "name": "Project Documentation",
  "description": "All project documentation and specs",
  "type": "document",
  "business_id": "uuid"
}
```

### Upload Document

```http
POST /knowledge-bases/:id/documents
Content-Type: multipart/form-data

file: (binary)
```

### Query Knowledge Base

```http
POST /knowledge-bases/:id/query
Content-Type: application/json

{
  "query": "How does the authentication system work?",
  "limit": 5
}
```

Response:
```json
{
  "results": [
    {
      "content": "The authentication system uses JWT tokens...",
      "source": "auth/README.md",
      "score": 0.92
    }
  ]
}
```

---

## Businesses

### List Businesses

```http
GET /businesses
```

### Create Business

```http
POST /businesses
Content-Type: application/json

{
  "name": "Mobile Game Studio",
  "description": "Casual mobile games portfolio",
  "industry": "Gaming",
  "monthly_budget": 200
}
```

### Get Business

```http
GET /businesses/:id
```

### Update Business

```http
PUT /businesses/:id
```

---

## API Keys

### List API Keys

```http
GET /api-keys
```

### Create API Key

```http
POST /api-keys
Content-Type: application/json

{
  "provider": "openai",
  "key": "sk-..."
}
```

Note: API keys are encrypted before storage and cannot be retrieved in plain text.

### Delete API Key

```http
DELETE /api-keys/:id
```

---

## Cost & Usage

### Get Usage Summary

```http
GET /usage
```

Query Parameters:
- `start_date` - Start date (YYYY-MM-DD)
- `end_date` - End date (YYYY-MM-DD)
- `business_id` - Filter by business

### Get Cost Breakdown

```http
GET /costs
```

Query Parameters:
- `period` - Period (7d, 14d, 30d, 90d)
- `group_by` - Group by (agent, provider, business)

---

## Billing

### Get Current Plan

```http
GET /billing/plan
```

### Get Pricing Plans

```http
GET /billing/plans
```

### Create Subscription

```http
POST /billing/subscribe
Content-Type: application/json

{
  "plan": "pro"
}
```

### Get Invoices

```http
GET /billing/invoices
```

---

## Webhooks

### GitHub Webhook

```http
POST /webhooks/github
X-GitHub-Event: push
X-Hub-Signature-256: sha256=...

{...github payload...}
```

### Stripe Webhook

```http
POST /webhooks/stripe
Stripe-Signature: t=...,v1=...

{...stripe payload...}
```

---

## Error Responses

All errors follow this format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Agent name is required",
    "details": {
      "field": "name"
    }
  }
}
```

Common Error Codes:
- `UNAUTHORIZED` - Invalid or missing authentication
- `FORBIDDEN` - Insufficient permissions
- `NOT_FOUND` - Resource not found
- `VALIDATION_ERROR` - Invalid request data
- `RATE_LIMITED` - Too many requests
- `INTERNAL_ERROR` - Server error

---

## Rate Limits

- Free plan: 100 requests per minute
- Pro plan: 1000 requests per minute
- Enterprise: Custom limits

Rate limit headers:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1704365000
```

---

## SDKs

Official SDKs are available for:
- TypeScript/JavaScript: `npm install @delphi/sdk`
- Python: `pip install delphi-sdk`
- Go: `go get github.com/delphi-platform/delphi-go`

Example (TypeScript):
```typescript
import { Delphi } from '@delphi/sdk';

const delphi = new Delphi({ apiKey: 'your-api-key' });

const agents = await delphi.agents.list();
const execution = await delphi.agents.execute('agent-id', {
  task: 'Review latest PR'
});
```

