# Delphi 🔮

**The AI Agent Command Center**

Delphi is a powerful platform for orchestrating dozens of AI agents across multiple businesses, repositories, and workflows. Think of it as mission control for your AI workforce.

![Delphi Dashboard](docs/images/dashboard-preview.png)

## ✨ Features

### 🤖 Multi-Agent Orchestration
- **Multiple AI Providers**: OpenAI, Anthropic, Google AI, and local LLMs via Ollama
- **Agent Briefing System**: Automatic context provisioning before each execution
- **Parallel Execution**: Run multiple agents simultaneously
- **Cost Tracking**: Per-agent, per-business spending analytics

### 📁 Repository Management
- **GitHub Integration**: Full OAuth-based repository connection
- **Branch Strategy**: Agents commit to dev/staging, you review and merge to main
- **Code Indexing**: Semantic search across your entire codebase
- **Webhook Support**: Trigger agents on PRs, commits, and issues

### 🧠 Central Knowledge Base
- **Vector Search**: Semantic search powered by pgvector
- **Multi-Source Ingestion**: Code, documents, PDFs, and more
- **Context System**: Agents automatically receive relevant context

### 📊 Palantir-Style Dashboards
- **Real-Time Visualizations**: Beautiful, data-rich dashboards
- **Cost Analytics**: Track spending across providers and businesses
- **Execution Metrics**: Success rates, response times, token usage

### 🏢 Multi-Business Support
- **Organization Management**: Multiple businesses under one account
- **Budget Controls**: Per-business spending limits
- **Access Control**: Role-based permissions (RBAC)

### 🔐 Enterprise Security
- **Encryption at Rest**: AES-256 for all sensitive data
- **Audit Logging**: Complete activity trail
- **API Key Management**: Secure storage with rotation support
- **Sandboxed Execution**: Isolated Fly.io machines for agent execution

## 🚀 Quick Start

### Prerequisites

- Docker & Docker Compose
- Node.js 18+ (for frontend development)
- Go 1.21+ (for backend development)
- GitHub OAuth App credentials

### 1. Clone the Repository

```bash
git clone https://github.com/delphi-platform/delphi.git
cd delphi
```

### 2. Configure Environment

```bash
cp env.example .env
# Edit .env with your configuration
```

Required environment variables:
```env
# Database
DATABASE_URL=postgres://postgres:postgres@localhost:5432/delphi?sslmode=disable

# Authentication
JWT_SECRET=your-secure-secret-key

# GitHub OAuth
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# AI Providers (add your API keys in the app)
# Keys are managed per-user in the application

# Stripe (for billing)
STRIPE_SECRET_KEY=sk_...
STRIPE_WEBHOOK_SECRET=whsec_...
```

### 3. Start Development Environment

```bash
docker-compose up -d
```

This starts:
- PostgreSQL database with pgvector
- Redis for caching and job queues
- Backend API (Go, hot-reloading)
- Frontend (React/Vite, hot-reloading)

### 4. Access the Application

- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **API Docs**: http://localhost:8080/docs

### 5. Run Database Migrations

```bash
docker-compose exec backend go run cmd/migrate/main.go up
```

## 📁 Project Structure

```
delphi/
├── backend/                 # Go backend
│   ├── cmd/
│   │   └── api/            # Main API entrypoint
│   ├── internal/
│   │   ├── config/         # Configuration
│   │   ├── handlers/       # HTTP handlers
│   │   ├── middleware/     # Auth, logging, etc.
│   │   ├── models/         # Data models
│   │   ├── providers/      # AI provider adapters
│   │   ├── repository/     # Database layer
│   │   ├── services/       # Business logic
│   │   ├── execution/      # Agent execution engine
│   │   ├── knowledge/      # Knowledge base
│   │   ├── github/         # GitHub integration
│   │   ├── billing/        # Stripe integration
│   │   ├── social/         # Social media integration
│   │   ├── iot/            # IoT device management
│   │   ├── security/       # Audit, RBAC
│   │   └── notifications/  # Email, Slack, Discord
│   └── pkg/
│       ├── auth/           # JWT utilities
│       ├── crypto/         # Encryption
│       └── logger/         # Structured logging
│
├── frontend/               # React frontend
│   ├── src/
│   │   ├── components/     # Reusable components
│   │   ├── pages/          # Page components
│   │   ├── stores/         # Zustand stores
│   │   ├── services/       # API client
│   │   └── types/          # TypeScript types
│   └── ...
│
├── migrations/             # SQL migrations
├── docs/                   # Documentation
├── docker-compose.yml      # Local development
└── env.example            # Environment template
```

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend (React)                      │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │
│  │Dashboard│ │ Agents  │ │Knowledge│ │Settings │           │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘           │
└───────┼──────────┼─────────┼─────────┼─────────────────────┘
        │          │         │         │
        └──────────┴────┬────┴─────────┘
                        │ REST API
        ┌───────────────▼───────────────┐
        │       Backend (Go)            │
        │  ┌─────────────────────────┐  │
        │  │     API Handlers        │  │
        │  └──────────┬──────────────┘  │
        │  ┌──────────▼──────────────┐  │
        │  │     Services Layer      │  │
        │  │ ┌─────┐ ┌─────┐ ┌─────┐ │  │
        │  │ │Agent│ │Auth │ │Cost │ │  │
        │  │ └─────┘ └─────┘ └─────┘ │  │
        │  └──────────┬──────────────┘  │
        │  ┌──────────▼──────────────┐  │
        │  │   Execution Engine      │  │
        │  │  (Fly.io Sandboxes)     │  │
        │  └──────────┬──────────────┘  │
        └─────────────┼─────────────────┘
                      │
    ┌─────────────────┼─────────────────────┐
    │                 │                      │
    ▼                 ▼                      ▼
┌────────┐     ┌──────────┐          ┌────────────┐
│ OpenAI │     │PostgreSQL│          │   Redis    │
│Anthropic│    │ pgvector │          │   Queue    │
│ Google │     │          │          │            │
│ Ollama │     └──────────┘          └────────────┘
└────────┘
```

## 🔧 Configuration

### AI Providers

Configure AI providers through the Settings page or API:

```typescript
// API Example
await delphi.apiKeys.create({
  provider: 'openai',
  key: 'sk-...'
});
```

### Agent Templates

Create reusable agent templates:

```yaml
# templates/code-reviewer.yaml
name: Code Review Oracle
purpose: coding
model_provider: openai
model: gpt-4-turbo
system_prompt: |
  You are an expert code reviewer. Your responsibilities:
  - Check for bugs and security vulnerabilities
  - Ensure code follows best practices
  - Suggest improvements for readability and performance
  - Be constructive and educational in your feedback
goal: Review code for quality, security, and best practices
```

### Knowledge Base Sources

Configure knowledge ingestion:

```yaml
# knowledge/config.yaml
sources:
  - type: repository
    url: https://github.com/org/repo
    patterns:
      - "**/*.ts"
      - "**/*.go"
      - "**/*.md"
  
  - type: documents
    path: /docs
    formats: [pdf, docx, md]
  
  - type: api
    endpoint: https://api.example.com/docs
    auth: bearer
```

## 📊 Monitoring

### Metrics

Prometheus metrics are exposed at `/metrics`:

```
delphi_agent_executions_total{agent="code-review",status="success"}
delphi_tokens_used_total{provider="openai"}
delphi_execution_duration_seconds{agent="code-review"}
```

### Logging

Structured JSON logs for easy aggregation:

```json
{
  "level": "info",
  "msg": "agent execution completed",
  "agent_id": "uuid",
  "execution_id": "uuid",
  "duration_ms": 2340,
  "tokens_used": 1500
}
```

## 🧪 Testing

```bash
# Backend tests
cd backend && go test ./...

# Frontend tests
cd frontend && npm test

# E2E tests
npm run test:e2e
```

## 📦 Deployment

### Docker Deployment

```bash
docker build -t delphi-backend ./backend
docker build -t delphi-frontend ./frontend
```

### Fly.io Deployment

```bash
# Deploy backend
cd backend && fly deploy

# Deploy frontend to Vercel
cd frontend && vercel --prod
```

### Kubernetes

Helm charts available in `deploy/helm/`:

```bash
helm install delphi ./deploy/helm/delphi \
  --set database.url=$DATABASE_URL \
  --set jwt.secret=$JWT_SECRET
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see [LICENSE](LICENSE) for details.

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/), [React](https://reactjs.org/), and [PostgreSQL](https://www.postgresql.org/)
- UI inspired by [Palantir](https://www.palantir.com/) and [Linear](https://linear.app/)
- Icons from [Lucide](https://lucide.dev/)

---

<p align="center">
  Built with 🔮 by the Delphi Team
</p>
