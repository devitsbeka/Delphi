import { motion } from 'framer-motion'

interface StatusBadgeProps {
  status: string
  size?: 'sm' | 'md'
}

const statusConfig: Record<string, { label: string; color: string; bgColor: string }> = {
  // Agent statuses
  configured: { label: 'Configured', color: 'text-delphi-text-muted', bgColor: 'bg-delphi-text-muted/10' },
  ready: { label: 'Ready', color: 'text-delphi-accent', bgColor: 'bg-delphi-accent/10' },
  briefing: { label: 'Briefing', color: 'text-delphi-warning', bgColor: 'bg-delphi-warning/10' },
  executing: { label: 'Executing', color: 'text-delphi-success', bgColor: 'bg-delphi-success/10' },
  paused: { label: 'Paused', color: 'text-orange-400', bgColor: 'bg-orange-400/10' },
  error: { label: 'Error', color: 'text-delphi-error', bgColor: 'bg-delphi-error/10' },
  terminated: { label: 'Terminated', color: 'text-delphi-text-muted', bgColor: 'bg-delphi-text-muted/10' },
  
  // Execution statuses
  pending: { label: 'Pending', color: 'text-delphi-text-muted', bgColor: 'bg-delphi-text-muted/10' },
  running: { label: 'Running', color: 'text-delphi-success', bgColor: 'bg-delphi-success/10' },
  completed: { label: 'Completed', color: 'text-delphi-success', bgColor: 'bg-delphi-success/10' },
  failed: { label: 'Failed', color: 'text-delphi-error', bgColor: 'bg-delphi-error/10' },
  
  // General statuses
  active: { label: 'Active', color: 'text-delphi-success', bgColor: 'bg-delphi-success/10' },
  inactive: { label: 'Inactive', color: 'text-delphi-text-muted', bgColor: 'bg-delphi-text-muted/10' },
  online: { label: 'Online', color: 'text-delphi-success', bgColor: 'bg-delphi-success/10' },
  offline: { label: 'Offline', color: 'text-delphi-text-muted', bgColor: 'bg-delphi-text-muted/10' },
}

export function StatusBadge({ status, size = 'sm' }: StatusBadgeProps) {
  const config = statusConfig[status.toLowerCase()] || {
    label: status,
    color: 'text-delphi-text-muted',
    bgColor: 'bg-delphi-text-muted/10',
  }

  const isAnimated = ['executing', 'running', 'briefing'].includes(status.toLowerCase())

  return (
    <span
      className={`inline-flex items-center gap-1.5 rounded-full font-medium ${config.color} ${config.bgColor} ${
        size === 'sm' ? 'px-2 py-0.5 text-xs' : 'px-3 py-1 text-sm'
      }`}
    >
      {isAnimated ? (
        <motion.span
          className="w-1.5 h-1.5 rounded-full bg-current"
          animate={{ opacity: [1, 0.3, 1] }}
          transition={{ duration: 1.5, repeat: Infinity }}
        />
      ) : (
        <span className="w-1.5 h-1.5 rounded-full bg-current" />
      )}
      {config.label}
    </span>
  )
}

export default StatusBadge
