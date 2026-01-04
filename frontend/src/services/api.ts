import axios from 'axios'
import { toast } from 'react-hot-toast'

// API Base URL - use environment variable or default to Fly.io backend
const API_URL = import.meta.env.VITE_API_URL || 'https://delphi-api.fly.dev/api/v1'

const api = axios.create({
  baseURL: API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 120000, // 2 minute timeout for AI responses
})

// Request interceptor for adding JWT token
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  },
)

// Response interceptor for error handling
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      let errorMessage = data.error || data.message || 'An unexpected error occurred.'

      switch (status) {
        case 400:
          errorMessage = data.error || 'Bad Request.'
          break
        case 401:
          errorMessage = data.error || 'Unauthorized. Please log in again.'
          localStorage.removeItem('token')
          localStorage.removeItem('user')
          break
        case 403:
          errorMessage = data.error || 'Forbidden.'
          break
        case 404:
          errorMessage = data.error || 'Resource not found.'
          break
        case 500:
          errorMessage = data.error || 'Internal Server Error.'
          break
        default:
          break
      }
      toast.error(errorMessage)
    } else if (error.request) {
      toast.error('No response from server. Please check your connection.')
    } else {
      toast.error('Error: ' + error.message)
    }
    return Promise.reject(error)
  },
)

// Auth API
export const authAPI = {
  login: (email: string, password: string) => 
    api.post('/auth/login', { email, password }),
  register: (name: string, email: string, password: string) => 
    api.post('/auth/register', { name, email, password }),
}

// Agents API
export const agentsAPI = {
  list: () => api.get('/agents'),
  get: (id: string) => api.get(`/agents/${id}`),
  create: (agent: any) => api.post('/agents', agent),
  update: (id: string, agent: any) => api.patch(`/agents/${id}`, agent),
  delete: (id: string) => api.delete(`/agents/${id}`),
  launch: (id: string) => api.post(`/agents/${id}/launch`),
  pause: (id: string) => api.post(`/agents/${id}/pause`),
  terminate: (id: string) => api.post(`/agents/${id}/terminate`),
}

// Execute API - Main AI interaction
export const executeAPI = {
  run: (agentId: string, prompt: string) => 
    api.post('/execute', { agent_id: agentId, prompt }),
  list: () => api.get('/executions'),
  get: (id: string) => api.get(`/executions/${id}`),
}

// Dashboard API
export const dashboardAPI = {
  overview: () => api.get('/dashboard/overview'),
}

// Providers API
export const providersAPI = {
  status: () => api.get('/providers/status'),
}

// Other APIs
export const repositoriesAPI = {
  list: () => api.get('/repositories'),
}

export const knowledgeAPI = {
  list: () => api.get('/knowledge'),
}

export const businessesAPI = {
  list: () => api.get('/businesses'),
}

export const costsAPI = {
  summary: () => api.get('/costs/summary'),
}

export default api
