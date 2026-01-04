# Delphi Deployment Guide

This guide covers deploying Delphi to production environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Infrastructure Setup](#infrastructure-setup)
3. [Backend Deployment](#backend-deployment)
4. [Frontend Deployment](#frontend-deployment)
5. [Database Setup](#database-setup)
6. [Agent Execution Environment](#agent-execution-environment)
7. [Monitoring & Observability](#monitoring--observability)
8. [Security Checklist](#security-checklist)
9. [Troubleshooting](#troubleshooting)

---

## Prerequisites

### Required Services

- **Fly.io Account**: For backend and agent execution
- **Vercel Account**: For frontend hosting
- **Supabase or PostgreSQL**: Database with pgvector extension
- **Redis**: For caching and job queues
- **Stripe Account**: For billing (optional)
- **GitHub OAuth App**: For repository integration

### CLI Tools

```bash
# Install required CLIs
brew install flyctl
npm i -g vercel
```

---

## Infrastructure Setup

### 1. Create Fly.io Application

```bash
cd backend

# Create the app
fly launch --name delphi-api --region sjc --no-deploy

# Set secrets
fly secrets set \
  DATABASE_URL="postgres://..." \
  REDIS_URL="redis://..." \
  JWT_SECRET="your-256-bit-secret" \
  GITHUB_CLIENT_ID="..." \
  GITHUB_CLIENT_SECRET="..." \
  STRIPE_SECRET_KEY="..." \
  STRIPE_WEBHOOK_SECRET="..."
```

### 2. Database Setup (Supabase)

```sql
-- Enable pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Run migrations
-- See migrations/ directory for schema
```

Or use Fly Postgres:

```bash
fly postgres create --name delphi-db --region sjc
fly postgres attach delphi-db
```

### 3. Redis Setup

```bash
fly redis create --name delphi-redis --region sjc
```

---

## Backend Deployment

### Dockerfile Configuration

The production Dockerfile (`backend/Dockerfile`):

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /api .
EXPOSE 8080

CMD ["./api"]
```

### Deploy to Fly.io

```bash
cd backend

# Deploy
fly deploy

# Scale as needed
fly scale count 2

# View logs
fly logs
```

### fly.toml Configuration

```toml
app = "delphi-api"
primary_region = "sjc"

[build]
  dockerfile = "Dockerfile"

[env]
  PORT = "8080"
  GIN_MODE = "release"

[http_service]
  internal_port = 8080
  force_https = true
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 1

[[services]]
  protocol = "tcp"
  internal_port = 8080

  [[services.ports]]
    port = 80
    handlers = ["http"]

  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]

  [[services.http_checks]]
    interval = 10000
    grace_period = "5s"
    method = "get"
    path = "/health"
    protocol = "http"
    timeout = 2000
```

---

## Frontend Deployment

### Build Configuration

```bash
cd frontend

# Build for production
npm run build

# Preview locally
npm run preview
```

### Deploy to Vercel

```bash
cd frontend

# First deployment
vercel

# Production deployment
vercel --prod
```

### Environment Variables (Vercel)

Set in Vercel dashboard or via CLI:

```bash
vercel env add VITE_API_URL production
# Enter: https://delphi-api.fly.dev
```

### vercel.json Configuration

```json
{
  "buildCommand": "npm run build",
  "outputDirectory": "dist",
  "framework": "vite",
  "rewrites": [
    { "source": "/(.*)", "destination": "/index.html" }
  ],
  "headers": [
    {
      "source": "/assets/(.*)",
      "headers": [
        { "key": "Cache-Control", "value": "public, max-age=31536000, immutable" }
      ]
    }
  ]
}
```

---

## Database Setup

### Run Migrations

```bash
# Using Go migrate
go run cmd/migrate/main.go up

# Or using psql
psql $DATABASE_URL -f migrations/001_initial_schema.sql
```

### Backup Strategy

```bash
# Fly Postgres backup
fly postgres backup create -a delphi-db

# List backups
fly postgres backup list -a delphi-db

# Restore
fly postgres backup restore <backup-id> -a delphi-db
```

### Connection Pooling

For high traffic, use PgBouncer:

```bash
fly postgres connect -a delphi-db
# Configure pooling mode in pg_hba.conf
```

---

## Agent Execution Environment

### Fly Machines for Sandboxed Execution

Agents execute in isolated Fly Machines:

```go
// Create machine for agent execution
machine, err := flyClient.CreateMachine(ctx, &fly.MachineConfig{
    App:    "delphi-agents",
    Region: "sjc",
    Config: fly.MachineCreateRequest{
        Image: "delphi-agent-runtime:latest",
        Guest: fly.MachineGuest{
            CPUKind:  "shared",
            CPUs:     1,
            MemoryMB: 256,
        },
        Env: map[string]string{
            "AGENT_ID": agentID,
            "TASK":     taskPayload,
        },
    },
    AutoDestroy: true,
})
```

### Agent Runtime Image

```dockerfile
# agent-runtime/Dockerfile
FROM python:3.11-slim

# Install common tools
RUN pip install openai anthropic langchain

# Security hardening
RUN useradd -m agent
USER agent

WORKDIR /workspace
COPY entrypoint.sh .
ENTRYPOINT ["./entrypoint.sh"]
```

---

## Monitoring & Observability

### Prometheus Metrics

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'delphi-api'
    static_configs:
      - targets: ['delphi-api.fly.dev:443']
    scheme: https
```

### Key Metrics to Monitor

| Metric | Alert Threshold |
|--------|----------------|
| `delphi_api_latency_p99` | > 2s |
| `delphi_error_rate` | > 1% |
| `delphi_agent_execution_duration` | > 5m |
| `delphi_queue_depth` | > 100 |
| `delphi_token_usage_rate` | > budget |

### Logging (Fly.io)

```bash
# Stream logs
fly logs -a delphi-api

# Search logs
fly logs -a delphi-api --filter "level=error"
```

### Alerting

Configure PagerDuty/Opsgenie webhooks:

```yaml
# alertmanager.yml
receivers:
  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: '<key>'

route:
  receiver: 'pagerduty'
  routes:
    - match:
        severity: critical
      receiver: 'pagerduty'
```

---

## Security Checklist

### Pre-Launch Checklist

- [ ] **JWT Secret**: 256-bit random secret
- [ ] **Database**: SSL required, connection encrypted
- [ ] **API Keys**: Encrypted at rest (AES-256)
- [ ] **HTTPS**: Forced for all endpoints
- [ ] **CORS**: Properly configured for frontend domain
- [ ] **Rate Limiting**: Enabled on all endpoints
- [ ] **Input Validation**: All inputs sanitized
- [ ] **SQL Injection**: Parameterized queries only
- [ ] **XSS**: Content Security Policy headers
- [ ] **Audit Logging**: All security events logged
- [ ] **Secrets Rotation**: 90-day rotation policy
- [ ] **Dependency Scanning**: Snyk/Dependabot enabled
- [ ] **Container Scanning**: Trivy on CI/CD

### Security Headers

```go
// middleware/security.go
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000")
        c.Header("Content-Security-Policy", "default-src 'self'")
        c.Next()
    }
}
```

### Secrets Management

```bash
# Rotate secrets
fly secrets set JWT_SECRET="new-secret"

# Unset old secrets
fly secrets unset OLD_SECRET
```

---

## Troubleshooting

### Common Issues

#### 1. Database Connection Failed

```bash
# Check connection
fly postgres connect -a delphi-db

# Verify DATABASE_URL
fly ssh console -a delphi-api
echo $DATABASE_URL
```

#### 2. Agent Execution Timeout

```bash
# Check machine logs
fly logs -a delphi-agents

# Increase timeout
fly machine update <id> --cpus 2 --memory 512
```

#### 3. High Latency

```bash
# Check metrics
fly dashboard -a delphi-api

# Scale up
fly scale count 4
```

#### 4. Redis Connection Issues

```bash
# Check Redis status
fly redis status -a delphi-redis

# Verify connection
fly ssh console -a delphi-api
redis-cli -u $REDIS_URL ping
```

### Support Channels

- **GitHub Issues**: For bugs and feature requests
- **Discord**: Community support
- **Email**: support@delphi.dev (Enterprise)

---

## Rollback Procedure

```bash
# List deployments
fly releases -a delphi-api

# Rollback to previous version
fly releases rollback -a delphi-api

# Or specific version
fly releases rollback v42 -a delphi-api
```

---

## Scaling Guide

### Horizontal Scaling

```bash
# Scale API servers
fly scale count 4 -a delphi-api

# Scale agent machines
fly scale count 10 -a delphi-agents
```

### Vertical Scaling

```bash
# Increase resources
fly scale vm shared-cpu-2x -a delphi-api
fly scale memory 1024 -a delphi-api
```

### Auto-scaling

```toml
# fly.toml
[http_service]
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 2
  max_machines_running = 10
```

---

## Cost Optimization

| Resource | Dev | Staging | Production |
|----------|-----|---------|------------|
| API Machines | 1x shared-cpu-1x | 2x shared-cpu-1x | 4x shared-cpu-2x |
| Database | 256MB | 1GB | 4GB |
| Redis | 256MB | 512MB | 1GB |
| Agent Machines | On-demand | On-demand | On-demand |

Estimated monthly costs:
- **Development**: ~$20/mo
- **Staging**: ~$50/mo
- **Production**: ~$200/mo (base) + usage

---

<p align="center">
  Need help? Contact support@delphi.dev
</p>

