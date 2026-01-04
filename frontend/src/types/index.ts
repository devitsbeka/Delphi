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
  | 'product'

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
  goal?: string
  model_provider: ModelProvider
  model: string
  status: AgentStatus
  system_prompt?: string
  organization_id?: string
  created_at?: string
  updated_at?: string
}

export interface AgentExecution {
  id: string
  agent_id: string
  agent_name: string
  prompt: string
  response: string
  status: 'pending' | 'running' | 'completed' | 'failed'
  provider: string
  model: string
  tokens_used?: number
  cost_usd?: number
  start_time?: string
  end_time?: string
  error_message?: string
}

// =============================================================================
// Repository Types
// =============================================================================

export interface Repository {
  id: string
  name: string
  url: string
  status: string
  branch_strategy: string
  provider?: string
  last_sync_at?: string
}

// =============================================================================
// Knowledge Base Types
// =============================================================================

export interface KnowledgeBase {
  id: string
  name: string
  description?: string
  type: 'repository' | 'document' | 'api' | 'custom'
  document_count: number
}

// =============================================================================
// Business Types
// =============================================================================

export interface Business {
  id: string
  name: string
  description?: string
  industry?: string
}

// =============================================================================
// Cost Types
// =============================================================================

export interface CostMetric {
  id: string
  agent_id?: string
  provider: ModelProvider
  model: string
  input_tokens: number
  output_tokens: number
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
  is_set: boolean
  last_used?: string
}
