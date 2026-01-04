import { motion } from 'framer-motion'
import { Link } from 'react-router-dom'
import type { Agent } from '../types'
import clsx from 'clsx'
import { 
  PlayIcon, 
  PauseIcon, 
  StopIcon,
  CpuChipIcon,
} from '@heroicons/react/24/outline'

interface AgentCardProps {
  agent: Agent
  onLaunch?: (id: string) => void
  onPause?: (id: string) => void
  onTerminate?: (id: string) => void
}

const purposeColors: Record<string, string> = {
  coding: 'tag-blue',
  content: 'tag-purple',
  devops: 'tag-green',
  analysis: 'tag-yellow',
  support: 'tag-blue',
  custom: 'tag-purple',
}

const providerLabels: Record<string, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  google: 'Google',
  ollama: 'Ollama',
  custom: 'Custom',
}

export default function AgentCard({ agent, onLaunch, onPause, onTerminate }: AgentCardProps) {
  const canLaunch = ['configured', 'paused', 'terminated'].includes(agent.status)
  const canPause = ['ready', 'executing'].includes(agent.status)
  const canTerminate = agent.status !== 'terminated' && agent.status !== 'configured'

  return (
    <motion.div
      className="glass-panel p-4 hover:border-delphi-border-light transition-colors"
      whileHover={{ scale: 1.01 }}
      transition={{ type: 'spring', stiffness: 400, damping: 30 }}
    >
      <div className="flex items-start justify-between gap-4">
        {/* Icon and info */}
        <div className="flex items-start gap-3 flex-1 min-w-0">
          <div className="w-10 h-10 rounded-lg bg-delphi-bg-tertiary flex items-center justify-center shrink-0">
            <CpuChipIcon className="w-5 h-5 text-delphi-accent-blue" />
          </div>
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-2 mb-1">
              <Link 
                to={`/agents/${agent.id}`}
                className="text-sm font-medium text-delphi-text-primary hover:text-delphi-accent-blue transition-colors truncate"
              >
                {agent.name}
              </Link>
              <div className={clsx('status-dot', agent.status)} />
            </div>
            <p className="text-2xs text-delphi-text-muted line-clamp-1">
              {agent.description || 'No description'}
            </p>
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center gap-1 shrink-0">
          {canLaunch && (
            <button
              onClick={() => onLaunch?.(agent.id)}
              className="p-1.5 rounded hover:bg-delphi-accent-green/20 text-delphi-accent-green transition-colors"
              title="Launch"
            >
              <PlayIcon className="w-4 h-4" />
            </button>
          )}
          {canPause && (
            <button
              onClick={() => onPause?.(agent.id)}
              className="p-1.5 rounded hover:bg-delphi-accent-yellow/20 text-delphi-accent-yellow transition-colors"
              title="Pause"
            >
              <PauseIcon className="w-4 h-4" />
            </button>
          )}
          {canTerminate && (
            <button
              onClick={() => onTerminate?.(agent.id)}
              className="p-1.5 rounded hover:bg-delphi-accent-red/20 text-delphi-accent-red transition-colors"
              title="Terminate"
            >
              <StopIcon className="w-4 h-4" />
            </button>
          )}
        </div>
      </div>

      {/* Meta info */}
      <div className="mt-3 flex items-center gap-2 flex-wrap">
        <span className={clsx('tag', purposeColors[agent.purpose] || 'tag-blue')}>
          {agent.purpose}
        </span>
        <span className="text-2xs text-delphi-text-muted">
          {providerLabels[agent.model_provider]} · {agent.model}
        </span>
      </div>

      {/* Status bar */}
      <div className="mt-3 pt-3 border-t border-delphi-border/50 flex items-center justify-between text-2xs">
        <span className="text-delphi-text-muted">Status</span>
        <span className={clsx(
          'uppercase tracking-wider font-medium',
          agent.status === 'ready' && 'text-delphi-accent-blue',
          agent.status === 'executing' && 'text-delphi-accent-green',
          agent.status === 'briefing' && 'text-delphi-accent-yellow',
          agent.status === 'paused' && 'text-delphi-accent-yellow',
          agent.status === 'error' && 'text-delphi-accent-red',
          (agent.status === 'configured' || agent.status === 'terminated') && 'text-delphi-text-muted',
        )}>
          {agent.status}
        </span>
      </div>
    </motion.div>
  )
}

