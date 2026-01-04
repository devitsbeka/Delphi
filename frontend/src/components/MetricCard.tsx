import { motion } from 'framer-motion'
import { Sparkline } from './Charts'

interface MetricCardProps {
  title: string
  value: string | number
  subtitle?: string
  trend?: {
    value: number
    direction: 'up' | 'down' | 'neutral'
  }
  sparkline?: number[]
  icon?: React.ReactNode
}

export function MetricCard({ title, value, subtitle, trend, sparkline, icon }: MetricCardProps) {
  const trendColor = trend?.direction === 'up' 
    ? 'text-delphi-success' 
    : trend?.direction === 'down' 
      ? 'text-delphi-error' 
      : 'text-delphi-text-muted'

  const trendIcon = trend?.direction === 'up' 
    ? '↑' 
    : trend?.direction === 'down' 
      ? '↓' 
      : '→'

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="card"
    >
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <p className="text-xs text-delphi-text-muted uppercase tracking-wide mb-1">
            {title}
          </p>
          <p className="text-2xl font-bold font-mono text-delphi-text-primary">
            {value}
          </p>
          {subtitle && (
            <p className="text-xs text-delphi-text-muted mt-0.5">{subtitle}</p>
          )}
          {trend && (
            <p className={`text-xs mt-1 ${trendColor}`}>
              {trendIcon} {Math.abs(trend.value)}% from last period
            </p>
          )}
        </div>
        
        <div className="flex flex-col items-end gap-2">
          {icon && (
            <div className="text-2xl text-delphi-text-muted">{icon}</div>
          )}
          {sparkline && sparkline.length > 0 && (
            <Sparkline data={sparkline} width={80} height={24} />
          )}
        </div>
      </div>
    </motion.div>
  )
}

export default MetricCard
