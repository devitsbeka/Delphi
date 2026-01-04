// =============================================================================
// User & Auth Types
// =============================================================================

export interface User {
  id: string
  email: string
  name: string
  tenant_id?: string
  role?: string
  preferences?: Record<string, unknown>
  created_at?: string
  updated_at?: string
}

export interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
}

// =============================================================================
// Agent Types
// =============================================================================

export type AgentStatus = 
  | 'configured'
  | 'ready'
  | 'briefing'
  | 'executing'
  | 'paused'
  | 'error'
  | 'terminated'

export type AgentPurpose = 
  | 'coding'
  | 'content'
  | 'devops'
  | 'analysis'
  | 'support'
  | 'custom'

export type ModelProvider = 
  | 'openai'
  | 'anthropic'
  | 'google'
  | 'ollama'

export interface Agent {
  id: string
  name: string
  description?: string
  purpose: AgentPurpose
  status: AgentStatus
  modelProvider: ModelProvider
  model: string
  systemPrompt?: string
  goal?: string
  organizationId?: string
  businessId?: string
  createdAt?: string
  updatedAt?: string
}

export interface AgentExecution {
  id: string
  agentId: string
  task: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  output?: string
  tokensUsed?: number
  cost?: number
  startedAt?: string
  completedAt?: string
}

// =============================================================================
// Repository Types
// =============================================================================

export interface Repository {
  id: string
  name: string
  fullName: string
  description?: string
  language?: string
  defaultBranch: string
  devBranch: string
  stagingBranch: string
  lastSync?: string
  indexed: boolean
  businessId?: string
}

// =============================================================================
// Knowledge Base Types
// =============================================================================

export interface KnowledgeBase {
  id: string
  name: string
  description?: string
  type: 'repository' | 'document' | 'api' | 'custom'
  documentCount: number
  vectorCount: number
  lastUpdated?: string
  size?: string
}

// =============================================================================
// Business Types
// =============================================================================

export interface Business {
  id: string
  name: string
  description?: string
  industry?: string
  status: 'active' | 'paused' | 'archived'
  agents: number
  repositories: number
  monthlyBudget: number
  monthlySpent: number
  createdAt?: string
}

// =============================================================================
// Cost Types
// =============================================================================

export interface CostMetric {
  id: string
  agentId?: string
  provider: ModelProvider
  model: string
  inputTokens: number
  outputTokens: number
  cost: number
  timestamp: string
}

// =============================================================================
// API Key Types
// =============================================================================

export interface APIKey {
  id: string
  provider: ModelProvider
  label: string
  isSet: boolean
  lastUsed?: string
  usageThisMonth?: number
}
