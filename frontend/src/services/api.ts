import axios from 'axios'

// Check if we're in demo mode
const isDemoMode = import.meta.env.VITE_DEMO_MODE === 'true' || !import.meta.env.VITE_API_URL

const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

// =============================================================================
// Demo Mode - Mock API responses for testing UI
// =============================================================================

const mockUser = {
  id: 'demo-user-id',
  email: 'demo@delphi.dev',
  name: 'Demo User',
}

const mockAgents = [
  { id: '1', name: 'Code Review Oracle', description: 'Automated code review for PRs', purpose: 'coding', status: 'ready', modelProvider: 'openai', model: 'gpt-4-turbo' },
  { id: '2', name: 'Bug Fixer', description: 'Automatically fixes bugs from issues', purpose: 'coding', status: 'executing', modelProvider: 'anthropic', model: 'claude-3-opus' },
  { id: '3', name: 'Content Writer', description: 'Blog posts and documentation', purpose: 'content', status: 'ready', modelProvider: 'openai', model: 'gpt-4-turbo' },
  { id: '4', name: 'DevOps Agent', description: 'CI/CD and infrastructure', purpose: 'devops', status: 'paused', modelProvider: 'google', model: 'gemini-pro' },
  { id: '5', name: 'Data Analyst', description: 'Analytics and reporting', purpose: 'analysis', status: 'briefing', modelProvider: 'openai', model: 'gpt-4-turbo' },
]

// Demo API wrapper
export const demoApi = {
  // Auth
  login: async (_email: string, _password: string) => {
    await new Promise(r => setTimeout(r, 500))
    return { token: 'demo-token', user: mockUser }
  },
  
  register: async (email: string, _password: string, name: string) => {
    await new Promise(r => setTimeout(r, 500))
    return { token: 'demo-token', user: { ...mockUser, email, name } }
  },
  
  // Agents
  getAgents: async () => {
    await new Promise(r => setTimeout(r, 300))
    return mockAgents
  },
  
  getAgent: async (id: string) => {
    await new Promise(r => setTimeout(r, 200))
    return mockAgents.find(a => a.id === id) || mockAgents[0]
  },
  
  createAgent: async (data: Record<string, unknown>) => {
    await new Promise(r => setTimeout(r, 500))
    return { id: Date.now().toString(), ...data, status: 'configured' }
  },
  
  updateAgent: async (id: string, data: Record<string, unknown>) => {
    await new Promise(r => setTimeout(r, 500))
    return { id, ...data }
  },
  
  deleteAgent: async (_id: string) => {
    await new Promise(r => setTimeout(r, 300))
    return { success: true }
  },
  
  // User
  getUser: async () => {
    return mockUser
  },
}

// Export the appropriate API based on mode
export const apiClient = isDemoMode ? {
  get: async (url: string) => {
    if (url.includes('/agents')) {
      const id = url.split('/').pop()
      if (id && id !== 'agents') {
        return { data: await demoApi.getAgent(id) }
      }
      return { data: await demoApi.getAgents() }
    }
    if (url.includes('/user')) {
      return { data: await demoApi.getUser() }
    }
    return { data: {} }
  },
  post: async (url: string, data?: Record<string, unknown>) => {
    if (url.includes('/login')) {
      return { data: await demoApi.login(data?.email as string, data?.password as string) }
    }
    if (url.includes('/register')) {
      return { data: await demoApi.register(data?.email as string, data?.password as string, data?.name as string) }
    }
    if (url.includes('/agents')) {
      return { data: await demoApi.createAgent(data || {}) }
    }
    return { data: {} }
  },
  put: async (url: string, data?: Record<string, unknown>) => {
    const id = url.split('/').pop()
    return { data: await demoApi.updateAgent(id!, data || {}) }
  },
  delete: async (url: string) => {
    const id = url.split('/').pop()
    return { data: await demoApi.deleteAgent(id!) }
  },
} : api

export default api
