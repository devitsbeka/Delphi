import { useEffect, useState, useMemo } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Link } from 'react-router-dom'
import { useAgentStore } from '../stores/agents'
import StatusBadge from '../components/StatusBadge'
import { Sparkline, chartColors } from '../components/Charts'
import type { Agent, AgentStatus, ModelProvider } from '../types'

const PURPOSES = ['coding', 'content', 'devops', 'analysis', 'support', 'all'] as const
const STATUSES: (AgentStatus | 'all')[] = ['all', 'ready', 'executing', 'briefing', 'paused', 'error']
const PROVIDERS: (ModelProvider | 'all')[] = ['all', 'openai', 'anthropic', 'google', 'ollama']

// Mock agents for demonstration
const mockAgents: Agent[] = [
  {
    id: '1', name: 'Code Review Oracle', description: 'Automated code review for PRs',
    purpose: 'coding', status: 'ready', modelProvider: 'openai', model: 'gpt-4-turbo',
    systemPrompt: '', goal: 'Review code for quality and best practices', 
    organizationId: '1', businessId: '1', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString()
  },
  {
    id: '2', name: 'Bug Fixer', description: 'Automatically fixes bugs from issues',
    purpose: 'coding', status: 'executing', modelProvider: 'anthropic', model: 'claude-3-opus',
    systemPrompt: '', goal: 'Fix bugs efficiently', 
    organizationId: '1', businessId: '1', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString()
  },
  {
    id: '3', name: 'Content Writer', description: 'Blog posts and documentation',
    purpose: 'content', status: 'ready', modelProvider: 'openai', model: 'gpt-4-turbo',
    systemPrompt: '', goal: 'Create engaging content', 
    organizationId: '1', businessId: '2', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString()
  },
  {
    id: '4', name: 'DevOps Agent', description: 'CI/CD and infrastructure',
    purpose: 'devops', status: 'paused', modelProvider: 'google', model: 'gemini-pro',
    systemPrompt: '', goal: 'Manage infrastructure', 
    organizationId: '1', businessId: '1', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString()
  },
  {
    id: '5', name: 'Data Analyst', description: 'Analytics and reporting',
    purpose: 'analysis', status: 'briefing', modelProvider: 'openai', model: 'gpt-4-turbo',
    systemPrompt: '', goal: 'Provide actionable insights', 
    organizationId: '1', businessId: '3', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString()
  },
  {
    id: '6', name: 'Customer Support', description: 'Handles support tickets',
    purpose: 'support', status: 'error', modelProvider: 'anthropic', model: 'claude-3-sonnet',
    systemPrompt: '', goal: 'Resolve customer issues', 
    organizationId: '1', businessId: '2', createdAt: new Date().toISOString(), updatedAt: new Date().toISOString()
  },
]

export function Agents() {
  const { agents, fetchAgents, loading } = useAgentStore()
  const [search, setSearch] = useState('')
  const [purposeFilter, setPurposeFilter] = useState<typeof PURPOSES[number]>('all')
  const [statusFilter, setStatusFilter] = useState<AgentStatus | 'all'>('all')
  const [providerFilter, setProviderFilter] = useState<ModelProvider | 'all'>('all')
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')

  useEffect(() => {
    fetchAgents()
  }, [fetchAgents])

  // Use mock agents if no real data
  const allAgents = agents.length > 0 ? agents : mockAgents

  const filteredAgents = useMemo(() => {
    return allAgents.filter((agent) => {
      const matchesSearch = 
        agent.name.toLowerCase().includes(search.toLowerCase()) ||
        agent.description?.toLowerCase().includes(search.toLowerCase())
      const matchesPurpose = purposeFilter === 'all' || agent.purpose === purposeFilter
      const matchesStatus = statusFilter === 'all' || agent.status === statusFilter
      const matchesProvider = providerFilter === 'all' || agent.modelProvider === providerFilter
      return matchesSearch && matchesPurpose && matchesStatus && matchesProvider
    })
  }, [allAgents, search, purposeFilter, statusFilter, providerFilter])

  const stats = useMemo(() => ({
    total: allAgents.length,
    ready: allAgents.filter(a => a.status === 'ready').length,
    executing: allAgents.filter(a => a.status === 'executing').length,
    error: allAgents.filter(a => a.status === 'error').length,
  }), [allAgents])

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-delphi-text-primary">Oracles</h1>
          <p className="text-sm text-delphi-text-muted">
            Manage your AI agents and their configurations
          </p>
        </div>
        <Link to="/agents/new" className="btn-primary">
          + Create Oracle
        </Link>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        {[
          { label: 'Total', value: stats.total, color: 'text-delphi-text-primary' },
          { label: 'Ready', value: stats.ready, color: 'text-delphi-accent' },
          { label: 'Executing', value: stats.executing, color: 'text-delphi-success' },
          { label: 'Errors', value: stats.error, color: 'text-delphi-error' },
        ].map((stat) => (
          <div key={stat.label} className="card py-4">
            <p className="text-xs text-delphi-text-muted uppercase tracking-wide">{stat.label}</p>
            <p className={`text-2xl font-bold font-mono ${stat.color}`}>{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Filters */}
      <div className="card">
        <div className="flex flex-wrap items-center gap-4">
          {/* Search */}
          <div className="flex-1 min-w-[200px]">
            <input
              type="text"
              placeholder="Search oracles..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="input-primary w-full"
            />
          </div>

          {/* Purpose Filter */}
          <select
            value={purposeFilter}
            onChange={(e) => setPurposeFilter(e.target.value as typeof PURPOSES[number])}
            className="input-primary"
          >
            {PURPOSES.map((purpose) => (
              <option key={purpose} value={purpose}>
                {purpose === 'all' ? 'All Purposes' : purpose.charAt(0).toUpperCase() + purpose.slice(1)}
              </option>
            ))}
          </select>

          {/* Status Filter */}
          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as AgentStatus | 'all')}
            className="input-primary"
          >
            {STATUSES.map((status) => (
              <option key={status} value={status}>
                {status === 'all' ? 'All Statuses' : status.charAt(0).toUpperCase() + status.slice(1)}
              </option>
            ))}
          </select>

          {/* Provider Filter */}
          <select
            value={providerFilter}
            onChange={(e) => setProviderFilter(e.target.value as ModelProvider | 'all')}
            className="input-primary"
          >
            {PROVIDERS.map((provider) => (
              <option key={provider} value={provider}>
                {provider === 'all' ? 'All Providers' : provider.charAt(0).toUpperCase() + provider.slice(1)}
              </option>
            ))}
          </select>

          {/* View Toggle */}
          <div className="flex rounded-lg border border-delphi-border overflow-hidden">
            <button
              onClick={() => setViewMode('grid')}
              className={`px-3 py-2 text-sm transition-colors ${
                viewMode === 'grid'
                  ? 'bg-delphi-accent text-white'
                  : 'bg-delphi-bg-elevated text-delphi-text-muted hover:text-delphi-text-primary'
              }`}
            >
              ▦
            </button>
            <button
              onClick={() => setViewMode('list')}
              className={`px-3 py-2 text-sm transition-colors ${
                viewMode === 'list'
                  ? 'bg-delphi-accent text-white'
                  : 'bg-delphi-bg-elevated text-delphi-text-muted hover:text-delphi-text-primary'
              }`}
            >
              ≡
            </button>
          </div>
        </div>
      </div>

      {/* Agents Grid/List */}
      <AnimatePresence mode="wait">
        {loading ? (
          <div className="flex items-center justify-center py-12">
            <div className="w-8 h-8 border-2 border-delphi-accent border-t-transparent rounded-full animate-spin" />
          </div>
        ) : filteredAgents.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-12"
          >
            <p className="text-delphi-text-muted">No oracles found matching your filters</p>
          </motion.div>
        ) : viewMode === 'grid' ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4"
          >
            {filteredAgents.map((agent, index) => (
              <AgentGridCard key={agent.id} agent={agent} index={index} />
            ))}
          </motion.div>
        ) : (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="space-y-2"
          >
            {filteredAgents.map((agent, index) => (
              <AgentListRow key={agent.id} agent={agent} index={index} />
            ))}
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  )
}

function AgentGridCard({ agent, index }: { agent: Agent; index: number }) {
  const mockStats = {
    executions: Math.floor(Math.random() * 1000),
    successRate: 85 + Math.random() * 15,
    avgCost: (Math.random() * 2).toFixed(2),
    sparkline: Array.from({ length: 12 }, () => Math.random() * 100),
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ delay: index * 0.05 }}
    >
      <Link to={`/agents/${agent.id}`}>
        <div className="card hover:border-delphi-accent/50 transition-colors cursor-pointer group">
          {/* Header */}
          <div className="flex items-start justify-between mb-3">
            <div>
              <h3 className="font-semibold text-delphi-text-primary group-hover:text-delphi-accent transition-colors">
                {agent.name}
              </h3>
              <p className="text-xs text-delphi-text-muted mt-0.5">{agent.description}</p>
            </div>
            <StatusBadge status={agent.status} />
          </div>

          {/* Model Info */}
          <div className="flex items-center gap-2 mb-4">
            <span className="px-2 py-1 text-xs bg-delphi-bg-primary rounded text-delphi-text-secondary">
              {agent.modelProvider}
            </span>
            <span className="text-xs text-delphi-text-muted">{agent.model}</span>
          </div>

          {/* Stats */}
          <div className="grid grid-cols-3 gap-2 text-center py-3 border-t border-delphi-border/50">
            <div>
              <p className="text-lg font-mono text-delphi-text-primary">{mockStats.executions}</p>
              <p className="text-2xs text-delphi-text-muted">Executions</p>
            </div>
            <div>
              <p className="text-lg font-mono text-delphi-success">{mockStats.successRate.toFixed(0)}%</p>
              <p className="text-2xs text-delphi-text-muted">Success</p>
            </div>
            <div>
              <p className="text-lg font-mono text-delphi-text-primary">${mockStats.avgCost}</p>
              <p className="text-2xs text-delphi-text-muted">Avg Cost</p>
            </div>
          </div>

          {/* Sparkline */}
          <div className="mt-2 flex justify-center">
            <Sparkline data={mockStats.sparkline} color={chartColors.blue} width={200} height={30} />
          </div>
        </div>
      </Link>
    </motion.div>
  )
}

function AgentListRow({ agent, index }: { agent: Agent; index: number }) {
  const mockStats = {
    executions: Math.floor(Math.random() * 1000),
    successRate: 85 + Math.random() * 15,
    avgCost: (Math.random() * 2).toFixed(2),
    lastRun: '2 hours ago',
  }

  return (
    <motion.div
      initial={{ opacity: 0, x: -20 }}
      animate={{ opacity: 1, x: 0 }}
      transition={{ delay: index * 0.03 }}
    >
      <Link to={`/agents/${agent.id}`}>
        <div className="card py-3 hover:border-delphi-accent/50 transition-colors cursor-pointer">
          <div className="flex items-center gap-4">
            {/* Status indicator */}
            <div className={`w-2 h-2 rounded-full ${
              agent.status === 'ready' ? 'bg-delphi-accent' :
              agent.status === 'executing' ? 'bg-delphi-success animate-pulse' :
              agent.status === 'error' ? 'bg-delphi-error' :
              'bg-delphi-text-muted'
            }`} />

            {/* Name and Description */}
            <div className="flex-1 min-w-0">
              <h3 className="font-medium text-delphi-text-primary truncate">{agent.name}</h3>
              <p className="text-xs text-delphi-text-muted truncate">{agent.description}</p>
            </div>

            {/* Provider */}
            <span className="px-2 py-1 text-xs bg-delphi-bg-primary rounded text-delphi-text-secondary w-20 text-center">
              {agent.modelProvider}
            </span>

            {/* Status */}
            <div className="w-24">
              <StatusBadge status={agent.status} />
            </div>

            {/* Stats */}
            <div className="hidden md:flex items-center gap-6 text-sm text-delphi-text-secondary">
              <div className="w-20 text-center">
                <span className="font-mono">{mockStats.executions}</span>
                <span className="text-delphi-text-muted ml-1 text-xs">runs</span>
              </div>
              <div className="w-16 text-center">
                <span className="font-mono text-delphi-success">{mockStats.successRate.toFixed(0)}%</span>
              </div>
              <div className="w-16 text-center font-mono">${mockStats.avgCost}</div>
              <div className="w-24 text-xs text-delphi-text-muted">{mockStats.lastRun}</div>
            </div>

            {/* Actions */}
            <button className="px-3 py-1.5 text-xs text-delphi-accent border border-delphi-accent rounded
                             hover:bg-delphi-accent hover:text-white transition-colors">
              Run
            </button>
          </div>
        </div>
      </Link>
    </motion.div>
  )
}
