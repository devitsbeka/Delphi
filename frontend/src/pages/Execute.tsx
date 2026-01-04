import { useState, useEffect, useRef } from 'react'
import { motion } from 'framer-motion'
import { PlayIcon, SparklesIcon, ClipboardDocumentIcon, CheckIcon } from '@heroicons/react/24/outline'
import useAgentsStore, { Execution } from '../stores/agents'
import { toast } from 'react-hot-toast'
import clsx from 'clsx'

export default function Execute() {
  const { agents, fetchAgents, executeAgent, executions, fetchExecutions, executing } = useAgentsStore()
  const [selectedAgentId, setSelectedAgentId] = useState<string>('')
  const [prompt, setPrompt] = useState('')
  const [currentExecution, setCurrentExecution] = useState<Execution | null>(null)
  const [copied, setCopied] = useState(false)
  const responseRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    fetchAgents()
    fetchExecutions()
  }, [fetchAgents, fetchExecutions])

  useEffect(() => {
    if (agents.length > 0 && !selectedAgentId) {
      setSelectedAgentId(agents[0].id)
    }
  }, [agents, selectedAgentId])

  const selectedAgent = agents.find((agent) => agent.id === selectedAgentId)

  const handleExecute = async () => {
    if (!selectedAgentId || !prompt.trim()) {
      toast.error('Please select an Oracle and provide a prompt.')
      return
    }

    try {
      const execution = await executeAgent(selectedAgentId, prompt)
      setCurrentExecution(execution)
      
      // Scroll to response
      setTimeout(() => {
        responseRef.current?.scrollIntoView({ behavior: 'smooth' })
      }, 100)
    } catch (error) {
      console.error('Execution failed:', error)
    }
  }

  const handleCopy = async () => {
    if (currentExecution?.response) {
      await navigator.clipboard.writeText(currentExecution.response)
      setCopied(true)
      toast.success('Copied to clipboard!')
      setTimeout(() => setCopied(false), 2000)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && (e.metaKey || e.ctrlKey)) {
      handleExecute()
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
      className="p-6 space-y-6 max-w-6xl mx-auto"
    >
      <div className="flex items-center gap-3">
        <SparklesIcon className="w-8 h-8 text-delphi-accent-blue" />
        <div>
          <h1 className="text-3xl font-bold text-delphi-text-primary">Execute Oracle</h1>
          <p className="text-delphi-text-muted">Send prompts to your AI agents and get real responses</p>
        </div>
      </div>

      {/* Agent Selection and Prompt */}
      <div className="glass-panel p-6 space-y-4">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="md:col-span-1">
            <label htmlFor="agent-select" className="label">Select Oracle</label>
            <select
              id="agent-select"
              className="input-primary w-full"
              value={selectedAgentId}
              onChange={(e) => setSelectedAgentId(e.target.value)}
            >
              {agents.map((agent) => (
                <option key={agent.id} value={agent.id}>
                  {agent.name} ({agent.model_provider})
                </option>
              ))}
            </select>
          </div>
          <div className="md:col-span-3">
            {selectedAgent && (
              <div className="p-3 rounded-lg bg-delphi-bg-tertiary border border-delphi-border">
                <div className="flex items-center gap-2 mb-1">
                  <span className="text-xs font-medium text-delphi-accent-blue uppercase tracking-wider">
                    {selectedAgent.purpose}
                  </span>
                  <span className="text-xs text-delphi-text-muted">•</span>
                  <span className="text-xs text-delphi-text-muted">
                    {selectedAgent.model_provider} / {selectedAgent.model}
                  </span>
                </div>
                <p className="text-sm text-delphi-text-secondary line-clamp-2">
                  {selectedAgent.description}
                </p>
              </div>
            )}
          </div>
        </div>

        <div>
          <label htmlFor="prompt-input" className="label">Your Prompt</label>
          <textarea
            id="prompt-input"
            className="input-primary w-full min-h-[150px] font-mono text-sm"
            placeholder="Enter your task or question for the Oracle...

Example prompts:
• Write a Go function that validates email addresses with proper error handling
• Create a marketing campaign for a new mobile game launch
• Analyze the financial implications of switching from AWS to Fly.io
• Design a CI/CD pipeline for a React + Go monorepo"
            value={prompt}
            onChange={(e) => setPrompt(e.target.value)}
            onKeyDown={handleKeyDown}
          />
          <p className="text-xs text-delphi-text-muted mt-1">
            Press ⌘+Enter to execute
          </p>
        </div>

        <button
          onClick={handleExecute}
          disabled={executing || !selectedAgentId || !prompt.trim()}
          className={clsx(
            'btn-primary w-full flex items-center justify-center gap-2 py-3',
            executing && 'opacity-75 cursor-wait'
          )}
        >
          {executing ? (
            <>
              <div className="w-5 h-5 border-2 border-white border-t-transparent rounded-full animate-spin" />
              Processing with {selectedAgent?.model_provider}...
            </>
          ) : (
            <>
              <PlayIcon className="w-5 h-5" />
              Execute Task
            </>
          )}
        </button>
      </div>

      {/* Response Section */}
      {currentExecution && (
        <motion.div
          ref={responseRef}
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="glass-panel p-6"
        >
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <SparklesIcon className="w-6 h-6 text-delphi-accent-green" />
              <div>
                <h2 className="text-lg font-semibold text-delphi-text-primary">Response</h2>
                <p className="text-xs text-delphi-text-muted">
                  {currentExecution.agent_name} • {currentExecution.provider}/{currentExecution.model}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <span className={clsx(
                'px-2 py-1 rounded text-xs font-medium',
                currentExecution.status === 'completed' && 'bg-green-500/20 text-green-400',
                currentExecution.status === 'failed' && 'bg-red-500/20 text-red-400',
                currentExecution.status === 'running' && 'bg-yellow-500/20 text-yellow-400'
              )}>
                {currentExecution.status}
              </span>
              <button
                onClick={handleCopy}
                className="p-2 rounded hover:bg-delphi-bg-tertiary transition-colors"
                title="Copy response"
              >
                {copied ? (
                  <CheckIcon className="w-5 h-5 text-green-400" />
                ) : (
                  <ClipboardDocumentIcon className="w-5 h-5 text-delphi-text-muted" />
                )}
              </button>
            </div>
          </div>

          {currentExecution.status === 'failed' ? (
            <div className="p-4 bg-red-500/10 border border-red-500/30 rounded-lg">
              <p className="text-red-400">{currentExecution.error_message}</p>
            </div>
          ) : (
            <div className="prose prose-invert max-w-none">
              <pre className="whitespace-pre-wrap text-sm text-delphi-text-primary bg-delphi-bg-tertiary p-4 rounded-lg overflow-x-auto">
                {currentExecution.response}
              </pre>
            </div>
          )}

          <div className="mt-4 pt-4 border-t border-delphi-border flex items-center justify-between text-xs text-delphi-text-muted">
            <div className="flex items-center gap-4">
              <span>Tokens: ~{currentExecution.tokens_used?.toLocaleString()}</span>
              <span>Cost: ~${currentExecution.cost_usd?.toFixed(4)}</span>
            </div>
            <span>
              {new Date(currentExecution.end_time).toLocaleString()}
            </span>
          </div>
        </motion.div>
      )}

      {/* Execution History */}
      {executions.length > 0 && (
        <div className="glass-panel p-6">
          <h2 className="text-lg font-semibold text-delphi-text-primary mb-4">Recent Executions</h2>
          <div className="space-y-3 max-h-96 overflow-y-auto custom-scrollbar">
            {executions.slice(0, 10).map((exec) => (
              <motion.div
                key={exec.id}
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className={clsx(
                  'p-4 rounded-lg border cursor-pointer transition-colors',
                  currentExecution?.id === exec.id 
                    ? 'bg-delphi-accent-blue/10 border-delphi-accent-blue' 
                    : 'bg-delphi-bg-tertiary border-delphi-border hover:border-delphi-border-light'
                )}
                onClick={() => setCurrentExecution(exec)}
              >
                <div className="flex items-start justify-between gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2 mb-1">
                      <span className="font-medium text-delphi-text-primary">{exec.agent_name}</span>
                      <span className={clsx(
                        'px-1.5 py-0.5 rounded text-2xs font-medium',
                        exec.status === 'completed' && 'bg-green-500/20 text-green-400',
                        exec.status === 'failed' && 'bg-red-500/20 text-red-400'
                      )}>
                        {exec.status}
                      </span>
                    </div>
                    <p className="text-sm text-delphi-text-muted line-clamp-1">{exec.prompt}</p>
                  </div>
                  <span className="text-xs text-delphi-text-muted whitespace-nowrap">
                    {new Date(exec.start_time).toLocaleTimeString()}
                  </span>
                </div>
              </motion.div>
            ))}
          </div>
        </div>
      )}
    </motion.div>
  )
}
