import { create } from 'zustand'
import { apiClient } from '../services/api'
import type { Agent } from '../types'

interface AgentState {
  agents: Agent[]
  currentAgent: Agent | null
  loading: boolean
  error: string | null

  fetchAgents: () => Promise<void>
  fetchAgent: (id: string) => Promise<void>
  createAgent: (data: Partial<Agent>) => Promise<void>
  updateAgent: (id: string, data: Partial<Agent>) => Promise<void>
  deleteAgent: (id: string) => Promise<void>
  clearError: () => void
}

export const useAgentStore = create<AgentState>((set) => ({
  agents: [],
  currentAgent: null,
  loading: false,
  error: null,

  fetchAgents: async () => {
    set({ loading: true, error: null })
    try {
      const response = await apiClient.get('/agents')
      const agents = Array.isArray(response.data) ? response.data : response.data.agents || []
      set({ agents, loading: false })
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } } }
      set({
        error: err.response?.data?.message || 'Failed to fetch agents',
        loading: false,
      })
    }
  },

  fetchAgent: async (id: string) => {
    set({ loading: true, error: null })
    try {
      const response = await apiClient.get(`/agents/${id}`)
      set({ currentAgent: response.data, loading: false })
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } } }
      set({
        error: err.response?.data?.message || 'Failed to fetch agent',
        loading: false,
      })
    }
  },

  createAgent: async (data: Partial<Agent>) => {
    set({ loading: true, error: null })
    try {
      const response = await apiClient.post('/agents', data as Record<string, unknown>)
      const newAgent = response.data
      set((state) => ({
        agents: [...state.agents, newAgent],
        loading: false,
      }))
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } } }
      set({
        error: err.response?.data?.message || 'Failed to create agent',
        loading: false,
      })
      throw error
    }
  },

  updateAgent: async (id: string, data: Partial<Agent>) => {
    set({ loading: true, error: null })
    try {
      const response = await apiClient.put(`/agents/${id}`, data as Record<string, unknown>)
      const updatedAgent = response.data
      set((state) => ({
        agents: state.agents.map((a) => (a.id === id ? updatedAgent : a)),
        currentAgent: state.currentAgent?.id === id ? updatedAgent : state.currentAgent,
        loading: false,
      }))
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } } }
      set({
        error: err.response?.data?.message || 'Failed to update agent',
        loading: false,
      })
      throw error
    }
  },

  deleteAgent: async (id: string) => {
    set({ loading: true, error: null })
    try {
      await apiClient.delete(`/agents/${id}`)
      set((state) => ({
        agents: state.agents.filter((a) => a.id !== id),
        currentAgent: state.currentAgent?.id === id ? null : state.currentAgent,
        loading: false,
      }))
    } catch (error: unknown) {
      const err = error as { response?: { data?: { message?: string } } }
      set({
        error: err.response?.data?.message || 'Failed to delete agent',
        loading: false,
      })
      throw error
    }
  },

  clearError: () => set({ error: null }),
}))
