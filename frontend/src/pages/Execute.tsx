import { useState, useCallback, useEffect } from 'react'
import { motion, AnimatePresence } from 'framer-motion'
import { useAgentStore } from '../stores/agents'
import type { Agent } from '../types'

// Mock agents
const mockAgents: Agent[] = [
  { id: '1', name: 'Code Review Oracle', purpose: 'coding', status: 'ready', modelProvider: 'openai', model: 'gpt-4-turbo' },
  { id: '2', name: 'Bug Fixer', purpose: 'coding', status: 'ready', modelProvider: 'anthropic', model: 'claude-3-opus' },
  { id: '3', name: 'Content Writer', purpose: 'content', status: 'ready', modelProvider: 'openai', model: 'gpt-4-turbo' },
  { id: '4', name: 'DevOps Agent', purpose: 'devops', status: 'ready', modelProvider: 'google', model: 'gemini-pro' },
  { id: '5', name: 'Data Analyst', purpose: 'analysis', status: 'ready', modelProvider: 'openai', model: 'gpt-4-turbo' },
]

interface ExecutionLog {
  id: string
  timestamp: Date
  type: 'system' | 'agent' | 'tool' | 'error' | 'success'
  message: string
  metadata?: Record<string, unknown>
}

export function Execute() {
  const [selectedAgents, setSelectedAgents] = useState<string[]>([])
  const [task, setTask] = useState('')
  const [context, setContext] = useState('')
  const [isExecuting, setIsExecuting] = useState(false)
  const [logs, setLogs] = useState<ExecutionLog[]>([])
  const [estimatedCost, setEstimatedCost] = useState(0)

  const { agents } = useAgentStore()
  const availableAgents = agents.length > 0 ? agents : mockAgents

  // Calculate estimated cost based on selected agents and task length
  useEffect(() => {
    const baseCost = selectedAgents.length * 0.05
    const taskCost = (task.length / 1000) * 0.01
    setEstimatedCost(baseCost + taskCost)
  }, [selectedAgents, task])

  const toggleAgent = useCallback((agentId: string) => {
    setSelectedAgents(prev => 
      prev.includes(agentId) 
        ? prev.filter(id => id !== agentId)
        : [...prev, agentId]
    )
  }, [])

  const addLog = useCallback((type: ExecutionLog['type'], message: string, metadata?: Record<string, unknown>) => {
    setLogs(prev => [...prev, {
      id: crypto.randomUUID(),
      timestamp: new Date(),
      type,
      message,
      metadata,
    }])
  }, [])

  const handleExecute = useCallback(async () => {
    if (selectedAgents.length === 0 || !task.trim()) return

    setIsExecuting(true)
    setLogs([])

    // Simulate execution with mock logs
    addLog('system', 'Initializing execution environment...')
    
    await new Promise(r => setTimeout(r, 500))
    
    for (const agentId of selectedAgents) {
      const agent = availableAgents.find(a => a.id === agentId)
      if (!agent) continue

      addLog('system', `Briefing Oracle: ${agent.name}`)
      await new Promise(r => setTimeout(r, 300))

      addLog('agent', `${agent.name}: Reading task context and system prompts...`, { agent: agent.name })
      await new Promise(r => setTimeout(r, 400))

      addLog('agent', `${agent.name}: Analyzing task requirements...`, { agent: agent.name })
      await new Promise(r => setTimeout(r, 600))

      addLog('tool', `${agent.name}: Querying knowledge base...`, { agent: agent.name })
      await new Promise(r => setTimeout(r, 500))

      addLog('agent', `${agent.name}: Formulating approach...`, { agent: agent.name })
      await new Promise(r => setTimeout(r, 400))

      if (agent.purpose === 'coding') {
        addLog('tool', `${agent.name}: Reading repository files...`, { agent: agent.name })
        await new Promise(r => setTimeout(r, 300))
        addLog('tool', `${agent.name}: Creating feature branch...`, { agent: agent.name })
        await new Promise(r => setTimeout(r, 200))
      }

      addLog('success', `${agent.name}: Task completed successfully`, { agent: agent.name })
      await new Promise(r => setTimeout(r, 200))
    }

    addLog('system', 'All oracles have completed their tasks')
    addLog('success', `Execution complete. Estimated cost: $${(estimatedCost * 1.2).toFixed(2)}`)

    setIsExecuting(false)
  }, [selectedAgents, task, addLog, availableAgents, estimatedCost])

  const getLogColor = (type: ExecutionLog['type']) => {
    switch (type) {
      case 'system': return 'text-delphi-text-muted'
      case 'agent': return 'text-delphi-accent'
      case 'tool': return 'text-delphi-warning'
      case 'error': return 'text-delphi-error'
      case 'success': return 'text-delphi-success'
    }
  }

  return (
    <div className="h-full flex flex-col p-6 gap-6">
      {/* Header */}
      <div>
        <h1 className="text-2xl font-bold text-delphi-text-primary">Execute Task</h1>
        <p className="text-sm text-delphi-text-muted">
          Launch one or more oracles to work on your task
        </p>
      </div>

      <div className="flex-1 grid grid-cols-1 lg:grid-cols-3 gap-6 min-h-0">
        {/* Left Panel - Configuration */}
        <div className="lg:col-span-1 space-y-6 overflow-y-auto">
          {/* Agent Selection */}
          <div className="card">
            <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Select Oracles</h3>
            <div className="space-y-2">
              {availableAgents.filter(a => a.status === 'ready').map((agent) => (
                <button
                  key={agent.id}
                  onClick={() => toggleAgent(agent.id)}
                  disabled={isExecuting}
                  className={`w-full p-3 rounded-lg border text-left transition-colors ${
                    selectedAgents.includes(agent.id)
                      ? 'bg-delphi-accent/10 border-delphi-accent'
                      : 'bg-delphi-bg-primary border-delphi-border hover:border-delphi-accent/50'
                  } ${isExecuting ? 'opacity-50 cursor-not-allowed' : ''}`}
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <p className="font-medium text-delphi-text-primary">{agent.name}</p>
                      <p className="text-xs text-delphi-text-muted">{agent.modelProvider} • {agent.model}</p>
                    </div>
                    <div className={`w-5 h-5 rounded border-2 flex items-center justify-center ${
                      selectedAgents.includes(agent.id)
                        ? 'border-delphi-accent bg-delphi-accent'
                        : 'border-delphi-border'
                    }`}>
                      {selectedAgents.includes(agent.id) && (
                        <span className="text-white text-xs">✓</span>
                      )}
                    </div>
                  </div>
                </button>
              ))}
            </div>
          </div>

          {/* Task Input */}
          <div className="card">
            <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Task Description</h3>
            <textarea
              value={task}
              onChange={(e) => setTask(e.target.value)}
              placeholder="Describe the task you want the oracle(s) to perform..."
              disabled={isExecuting}
              className="input-primary w-full h-32 resize-none"
            />
          </div>

          {/* Context */}
          <div className="card">
            <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Additional Context</h3>
            <textarea
              value={context}
              onChange={(e) => setContext(e.target.value)}
              placeholder="Any additional context, constraints, or preferences..."
              disabled={isExecuting}
              className="input-primary w-full h-24 resize-none"
            />
          </div>

          {/* Cost Estimate & Execute */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <span className="text-delphi-text-muted">Estimated Cost</span>
              <span className="text-xl font-mono text-delphi-text-primary">
                ${estimatedCost.toFixed(2)}
              </span>
            </div>
            <button
              onClick={handleExecute}
              disabled={selectedAgents.length === 0 || !task.trim() || isExecuting}
              className="btn-primary w-full disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isExecuting ? (
                <span className="flex items-center justify-center gap-2">
                  <span className="w-4 h-4 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                  Executing...
                </span>
              ) : (
                `Launch ${selectedAgents.length} Oracle${selectedAgents.length !== 1 ? 's' : ''}`
              )}
            </button>
          </div>
        </div>

        {/* Right Panel - Execution Log */}
        <div className="lg:col-span-2 card flex flex-col min-h-0">
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-delphi-text-primary">Execution Log</h3>
            {logs.length > 0 && (
              <button
                onClick={() => setLogs([])}
                className="text-xs text-delphi-text-muted hover:text-delphi-text-primary transition-colors"
              >
                Clear
              </button>
            )}
          </div>

          <div className="flex-1 bg-delphi-bg-primary rounded-lg border border-delphi-border p-4 overflow-y-auto font-mono text-sm">
            {logs.length === 0 ? (
              <div className="h-full flex items-center justify-center text-delphi-text-muted">
                <p>Select oracles and enter a task to begin execution</p>
              </div>
            ) : (
              <AnimatePresence>
                {logs.map((log) => (
                  <motion.div
                    key={log.id}
                    initial={{ opacity: 0, y: 10 }}
                    animate={{ opacity: 1, y: 0 }}
                    className={`mb-2 ${getLogColor(log.type)}`}
                  >
                    <span className="text-delphi-text-muted mr-2">
                      [{log.timestamp.toLocaleTimeString()}]
                    </span>
                    {log.type === 'system' && <span className="text-delphi-text-muted mr-2">[SYS]</span>}
                    {log.type === 'agent' && <span className="text-delphi-accent mr-2">[AGT]</span>}
                    {log.type === 'tool' && <span className="text-delphi-warning mr-2">[TL]</span>}
                    {log.type === 'error' && <span className="text-delphi-error mr-2">[ERR]</span>}
                    {log.type === 'success' && <span className="text-delphi-success mr-2">[OK]</span>}
                    {log.message}
                  </motion.div>
                ))}
              </AnimatePresence>
            )}
          </div>

          {/* Execution Status Bar */}
          {isExecuting && (
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              className="mt-4 flex items-center gap-4"
            >
              <div className="flex-1 h-2 bg-delphi-bg-primary rounded-full overflow-hidden">
                <motion.div
                  className="h-full bg-gradient-to-r from-delphi-accent to-delphi-success"
                  initial={{ width: '0%' }}
                  animate={{ width: '100%' }}
                  transition={{ duration: 10, ease: 'linear' }}
                />
              </div>
              <span className="text-sm text-delphi-text-muted">
                Processing...
              </span>
            </motion.div>
          )}
        </div>
      </div>
    </div>
  )
}
