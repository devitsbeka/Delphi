import { create } from 'zustand'
import { agentsAPI, executeAPI } from '../services/api'
import { Agent } from '../types'
import { toast } from 'react-hot-toast'

export interface Execution {
  id: string
  agent_id: string
  agent_name: string
  prompt: string
  response: string
  status: 'running' | 'completed' | 'failed'
  provider: string
  model: string
  tokens_used: number
  cost_usd: number
  start_time: string
  end_time: string
  error_message?: string
}

interface AgentsState {
  agents: Agent[]
  executions: Execution[]
  loading: boolean
  executing: boolean
  error: string | null
  fetchAgents: () => Promise<void>
  createAgent: (agent: Partial<Agent>) => Promise<void>
  updateAgent: (id: string, agent: Partial<Agent>) => Promise<void>
  deleteAgent: (id: string) => Promise<void>
  launchAgent: (id: string) => Promise<void>
  pauseAgent: (id: string) => Promise<void>
  terminateAgent: (id: string) => Promise<void>
  executeAgent: (agentId: string, prompt: string) => Promise<Execution>
  fetchExecutions: () => Promise<void>
}

const useAgentsStore = create<AgentsState>((set) => ({
  agents: [],
  executions: [],
  loading: false,
  executing: false,
  error: null,

  fetchAgents: async () => {
    set({ loading: true, error: null })
    try {
      const response = await agentsAPI.list()
      set({ agents: response.data, loading: false })
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Unknown error'
      set({ error: message, loading: false })
    }
  },

  createAgent: async (newAgentData) => {
    set({ loading: true, error: null })
    try {
      const response = await agentsAPI.create(newAgentData)
      set((state) => ({
        agents: [...state.agents, response.data],
        loading: false,
      }))
      toast.success('Oracle created!')
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Unknown error'
      set({ error: message, loading: false })
      throw error
    }
  },

  updateAgent: async (id, updatedAgentData) => {
    set({ loading: true, error: null })
    try {
      const response = await agentsAPI.update(id, updatedAgentData)
      set((state) => ({
        agents: state.agents.map((agent) => (agent.id === id ? response.data : agent)),
        loading: false,
      }))
      toast.success('Oracle updated!')
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Unknown error'
      set({ error: message, loading: false })
      throw error
    }
  },

  deleteAgent: async (id) => {
    set({ loading: true, error: null })
    try {
      await agentsAPI.delete(id)
      set((state) => ({
        agents: state.agents.filter((agent) => agent.id !== id),
        loading: false,
      }))
      toast.success('Oracle deleted!')
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Unknown error'
      set({ error: message, loading: false })
      throw error
    }
  },

  launchAgent: async (id) => {
    try {
      await agentsAPI.launch(id)
      set((state) => ({
        agents: state.agents.map((agent) => 
          agent.id === id ? { ...agent, status: 'ready' as const } : agent
        ),
      }))
      toast.success('Oracle launched!')
    } catch {
      toast.error('Failed to launch oracle')
    }
  },

  pauseAgent: async (id) => {
    try {
      await agentsAPI.pause(id)
      set((state) => ({
        agents: state.agents.map((agent) => 
          agent.id === id ? { ...agent, status: 'paused' as const } : agent
        ),
      }))
      toast.success('Oracle paused')
    } catch {
      toast.error('Failed to pause oracle')
    }
  },

  terminateAgent: async (id) => {
    try {
      await agentsAPI.terminate(id)
      set((state) => ({
        agents: state.agents.map((agent) => 
          agent.id === id ? { ...agent, status: 'terminated' as const } : agent
        ),
      }))
      toast.success('Oracle terminated')
    } catch {
      toast.error('Failed to terminate oracle')
    }
  },

  executeAgent: async (agentId: string, prompt: string) => {
    set({ executing: true, error: null })
    try {
      const response = await executeAPI.run(agentId, prompt)
      const execution = response.data as Execution
      
      set((state) => ({
        executions: [execution, ...state.executions],
        executing: false,
      }))
      
      if (execution.status === 'completed') {
        toast.success('Execution completed!')
      }
      
      return execution
    } catch (error: unknown) {
      const message = error instanceof Error ? error.message : 'Unknown error'
      set({ executing: false, error: message })
      throw error
    }
  },

  fetchExecutions: async () => {
    try {
      const response = await executeAPI.list()
      set({ executions: response.data })
    } catch (error: unknown) {
      console.error('Failed to fetch executions:', error)
    }
  },
}))

export default useAgentsStore
