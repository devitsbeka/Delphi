import { useEffect, useState, useMemo } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { Link } from 'react-router-dom'
import useAgentsStore from '../stores/agents'
import StatusBadge from '../components/StatusBadge'
import type { Agent, AgentStatus, ModelProvider, AgentPurpose } from '../types'

const PURPOSES: (AgentPurpose | 'all')[] = ['all', 'coding', 'content', 'devops', 'analysis', 'support', 'custom']
const STATUSES: (AgentStatus | 'all')[] = ['all', 'ready', 'executing', 'briefing', 'paused', 'error', 'configured']
const PROVIDERS: (ModelProvider | 'all')[] = ['all', 'openai', 'anthropic', 'google', 'ollama']

export default function Agents() {
  const { agents, fetchAgents, loading, launchAgent, pauseAgent, terminateAgent } = useAgentsStore()
  const [search, setSearch] = useState('')
  const [purposeFilter, setPurposeFilter] = useState<AgentPurpose | 'all'>('all')
  const [statusFilter, setStatusFilter] = useState<AgentStatus | 'all'>('all')
  const [providerFilter, setProviderFilter] = useState<ModelProvider | 'all'>('all')
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid')

  useEffect(() => {
    fetchAgents()
  }, [fetchAgents])

  const filteredAgents = useMemo(() => {
    return agents.filter((agent: Agent) => {
      const matchesSearch = 
        agent.name.toLowerCase().includes(search.toLowerCase()) ||
        (agent.description?.toLowerCase().includes(search.toLowerCase()) ?? false)
      const matchesPurpose = purposeFilter === 'all' || agent.purpose === purposeFilter
      const matchesStatus = statusFilter === 'all' || agent.status === statusFilter
      const matchesProvider = providerFilter === 'all' || agent.model_provider === providerFilter
      return matchesSearch && matchesPurpose && matchesStatus && matchesProvider
    })
  }, [agents, search, purposeFilter, statusFilter, providerFilter])

  const statusCounts = useMemo(() => {
    const counts: Record<string, number> = { all: agents.length }
    agents.forEach((agent: Agent) => {
      counts[agent.status] = (counts[agent.status] || 0) + 1
    })
    return counts
  }, [agents])

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="p-6 space-y-6"
    >
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-delphi-text-primary">Oracles</h1>
          <p className="text-sm text-delphi-text-muted mt-1">
            {agents.length} agents configured • {statusCounts['ready'] || 0} ready
          </p>
        </div>
        <Link to="/agents/new" className="btn-primary">
          + New Oracle
        </Link>
      </div>

      {/* Filters */}
      <div className="flex flex-wrap gap-3">
        <input
          type="text"
          placeholder="Search oracles..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="input-primary flex-1 min-w-[200px]"
        />
        <select
          value={purposeFilter}
          onChange={(e) => setPurposeFilter(e.target.value as AgentPurpose | 'all')}
          className="input-primary"
        >
          {PURPOSES.map((p) => (
            <option key={p} value={p}>{p === 'all' ? 'All Purposes' : p}</option>
          ))}
        </select>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value as AgentStatus | 'all')}
          className="input-primary"
        >
          {STATUSES.map((s) => (
            <option key={s} value={s}>{s === 'all' ? 'All Statuses' : s}</option>
          ))}
        </select>
        <select
          value={providerFilter}
          onChange={(e) => setProviderFilter(e.target.value as ModelProvider | 'all')}
          className="input-primary"
        >
          {PROVIDERS.map((p) => (
            <option key={p} value={p}>{p === 'all' ? 'All Providers' : p}</option>
          ))}
        </select>
        <div className="flex border border-delphi-border rounded-lg overflow-hidden">
          <button
            onClick={() => setViewMode('grid')}
            className={`px-3 py-2 text-sm ${viewMode === 'grid' ? 'bg-delphi-accent text-white' : 'bg-delphi-bg-secondary text-delphi-text-muted'}`}
          >
            Grid
          </button>
          <button
            onClick={() => setViewMode('list')}
            className={`px-3 py-2 text-sm ${viewMode === 'list' ? 'bg-delphi-accent text-white' : 'bg-delphi-bg-secondary text-delphi-text-muted'}`}
          >
            List
          </button>
        </div>
      </div>

      {/* Loading State */}
      {loading && (
        <div className="text-center py-12">
          <div className="w-8 h-8 border-2 border-delphi-accent border-t-transparent rounded-full animate-spin mx-auto" />
          <p className="text-delphi-text-muted mt-4">Loading oracles...</p>
        </div>
      )}

      {/* Agents Grid/List */}
      {!loading && (
        <AnimatePresence mode="popLayout">
          <motion.div
            layout
            className={viewMode === 'grid' 
              ? 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4'
              : 'space-y-3'
            }
          >
            {filteredAgents.map((agent: Agent, index: number) => (
              <motion.div
                key={agent.id}
                layout
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, scale: 0.9 }}
                transition={{ delay: index * 0.05 }}
                className="card hover:border-delphi-accent/50 transition-colors group"
              >
                <div className="flex items-start justify-between">
                  <div className="flex-1 min-w-0">
                    <Link to={`/agents/${agent.id}`} className="block">
                      <h3 className="font-semibold text-delphi-text-primary group-hover:text-delphi-accent transition-colors truncate">
                        {agent.name}
                      </h3>
                    </Link>
                    <p className="text-sm text-delphi-text-muted mt-1 line-clamp-2">
                      {agent.description || 'No description'}
                    </p>
                  </div>
                  <StatusBadge status={agent.status} />
                </div>

                <div className="mt-4 flex items-center gap-2 flex-wrap">
                  <span className="px-2 py-0.5 text-xs rounded bg-delphi-bg-tertiary text-delphi-text-secondary">
                    {agent.purpose}
                  </span>
                  <span className="px-2 py-0.5 text-xs rounded bg-delphi-bg-tertiary text-delphi-text-secondary">
                    {agent.model_provider}
                  </span>
                  <span className="px-2 py-0.5 text-xs rounded bg-delphi-bg-tertiary text-delphi-text-secondary">
                    {agent.model}
                  </span>
                </div>

                <div className="mt-4 pt-4 border-t border-delphi-border flex items-center justify-between">
                  <div className="flex gap-2">
                    {(agent.status === 'configured' || agent.status === 'paused') && (
                      <button
                        onClick={() => launchAgent(agent.id)}
                        className="px-3 py-1 text-xs bg-green-500/20 text-green-400 rounded hover:bg-green-500/30 transition-colors"
                      >
                        Launch
                      </button>
                    )}
                    {(agent.status === 'ready' || agent.status === 'executing') && (
                      <button
                        onClick={() => pauseAgent(agent.id)}
                        className="px-3 py-1 text-xs bg-yellow-500/20 text-yellow-400 rounded hover:bg-yellow-500/30 transition-colors"
                      >
                        Pause
                      </button>
                    )}
                    {agent.status !== 'terminated' && agent.status !== 'configured' && (
                      <button
                        onClick={() => terminateAgent(agent.id)}
                        className="px-3 py-1 text-xs bg-red-500/20 text-red-400 rounded hover:bg-red-500/30 transition-colors"
                      >
                        Terminate
                      </button>
                    )}
                  </div>
                  <Link
                    to={`/execute?agent=${agent.id}`}
                    className="px-3 py-1 text-xs bg-delphi-accent/20 text-delphi-accent rounded hover:bg-delphi-accent/30 transition-colors"
                  >
                    Execute
                  </Link>
                </div>
              </motion.div>
            ))}
          </motion.div>
        </AnimatePresence>
      )}

      {/* Empty State */}
      {!loading && filteredAgents.length === 0 && (
        <div className="text-center py-12">
          <span className="text-4xl">🔮</span>
          <p className="text-delphi-text-muted mt-4">No oracles found</p>
          <Link to="/agents/new" className="btn-primary mt-4 inline-block">
            Create your first Oracle
          </Link>
        </div>
      )}
    </motion.div>
  )
}
