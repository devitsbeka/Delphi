import { motion } from 'framer-motion'
import {
  LineChart,
  Line,
  AreaChart,
  Area,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts'

// =============================================================================
// Chart Colors
// =============================================================================

export const chartColors = {
  blue: '#4a9eff',
  cyan: '#00d4ff',
  green: '#00d68f',
  yellow: '#ffaa00',
  orange: '#ff6b35',
  red: '#ff4757',
  purple: '#a855f7',
  gray: '#606070',
}

const colorArray = [
  chartColors.blue,
  chartColors.cyan,
  chartColors.green,
  chartColors.yellow,
  chartColors.purple,
  chartColors.orange,
]

// =============================================================================
// Custom Tooltip
// =============================================================================

interface CustomTooltipProps {
  active?: boolean
  payload?: Array<{ name: string; value: number; color: string }>
  label?: string
}

const CustomTooltip = ({ active, payload, label }: CustomTooltipProps) => {
  if (!active || !payload) return null

  return (
    <div className="bg-delphi-bg-elevated border border-delphi-border rounded-lg p-3 shadow-lg">
      <p className="text-xs text-delphi-text-muted mb-2">{label}</p>
      {payload.map((entry, index) => (
        <div key={index} className="flex items-center gap-2 text-sm">
          <span
            className="w-2 h-2 rounded-full"
            style={{ backgroundColor: entry.color }}
          />
          <span className="text-delphi-text-secondary">{entry.name}:</span>
          <span className="text-delphi-text-primary font-mono">
            {typeof entry.value === 'number' ? entry.value.toLocaleString() : entry.value}
          </span>
        </div>
      ))}
    </div>
  )
}

// =============================================================================
// Activity Chart (Line/Area)
// =============================================================================

interface ActivityChartProps {
  data: Array<{ date: string; executions: number; cost: number }>
  height?: number
}

export function ActivityChart({ data, height = 200 }: ActivityChartProps) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
    >
      <ResponsiveContainer width="100%" height={height}>
        <AreaChart data={data} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
          <defs>
            <linearGradient id="colorExecutions" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={chartColors.blue} stopOpacity={0.3} />
              <stop offset="95%" stopColor={chartColors.blue} stopOpacity={0} />
            </linearGradient>
            <linearGradient id="colorCost" x1="0" y1="0" x2="0" y2="1">
              <stop offset="5%" stopColor={chartColors.green} stopOpacity={0.3} />
              <stop offset="95%" stopColor={chartColors.green} stopOpacity={0} />
            </linearGradient>
          </defs>
          <CartesianGrid strokeDasharray="3 3" stroke="#2a2a3a" />
          <XAxis
            dataKey="date"
            stroke="#606070"
            fontSize={10}
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            stroke="#606070"
            fontSize={10}
            tickLine={false}
            axisLine={false}
            yAxisId="left"
          />
          <YAxis
            stroke="#606070"
            fontSize={10}
            tickLine={false}
            axisLine={false}
            yAxisId="right"
            orientation="right"
            tickFormatter={(value) => `$${value}`}
          />
          <Tooltip content={<CustomTooltip />} />
          <Area
            type="monotone"
            dataKey="executions"
            stroke={chartColors.blue}
            fillOpacity={1}
            fill="url(#colorExecutions)"
            yAxisId="left"
            name="Executions"
          />
          <Area
            type="monotone"
            dataKey="cost"
            stroke={chartColors.green}
            fillOpacity={1}
            fill="url(#colorCost)"
            yAxisId="right"
            name="Cost"
          />
        </AreaChart>
      </ResponsiveContainer>
    </motion.div>
  )
}

// =============================================================================
// Cost Breakdown Chart (Pie)
// =============================================================================

interface CostBreakdownProps {
  data: Array<{ name: string; value: number }>
  height?: number
}

export function CostBreakdownChart({ data, height = 200 }: CostBreakdownProps) {
  return (
    <motion.div
      initial={{ opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ duration: 0.5 }}
    >
      <ResponsiveContainer width="100%" height={height}>
        <PieChart>
          <Pie
            data={data}
            cx="50%"
            cy="50%"
            innerRadius={50}
            outerRadius={70}
            paddingAngle={2}
            dataKey="value"
          >
            {data.map((_, index) => (
              <Cell
                key={`cell-${index}`}
                fill={colorArray[index % colorArray.length]}
              />
            ))}
          </Pie>
          <Tooltip content={<CustomTooltip />} />
          <Legend
            layout="vertical"
            align="right"
            verticalAlign="middle"
            iconType="circle"
            iconSize={8}
            formatter={(value) => (
              <span className="text-xs text-delphi-text-secondary">{value}</span>
            )}
          />
        </PieChart>
      </ResponsiveContainer>
    </motion.div>
  )
}

// =============================================================================
// Agent Status Chart (Bar)
// =============================================================================

interface AgentStatusChartProps {
  data: Array<{ status: string; count: number }>
  height?: number
}

const statusColors: Record<string, string> = {
  ready: chartColors.blue,
  executing: chartColors.green,
  briefing: chartColors.yellow,
  paused: chartColors.orange,
  error: chartColors.red,
  configured: chartColors.gray,
  terminated: chartColors.gray,
}

export function AgentStatusChart({ data, height = 150 }: AgentStatusChartProps) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
    >
      <ResponsiveContainer width="100%" height={height}>
        <BarChart data={data} layout="vertical" margin={{ left: 60, right: 20 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#2a2a3a" horizontal={false} />
          <XAxis type="number" stroke="#606070" fontSize={10} tickLine={false} />
          <YAxis
            type="category"
            dataKey="status"
            stroke="#606070"
            fontSize={10}
            tickLine={false}
            axisLine={false}
          />
          <Tooltip content={<CustomTooltip />} />
          <Bar dataKey="count" name="Agents" radius={[0, 4, 4, 0]}>
            {data.map((entry, index) => (
              <Cell
                key={`cell-${index}`}
                fill={statusColors[entry.status.toLowerCase()] || chartColors.gray}
              />
            ))}
          </Bar>
        </BarChart>
      </ResponsiveContainer>
    </motion.div>
  )
}

// =============================================================================
// Token Usage Chart (Line)
// =============================================================================

interface TokenUsageChartProps {
  data: Array<{ date: string; input: number; output: number }>
  height?: number
}

export function TokenUsageChart({ data, height = 200 }: TokenUsageChartProps) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
    >
      <ResponsiveContainer width="100%" height={height}>
        <LineChart data={data} margin={{ top: 10, right: 10, left: 0, bottom: 0 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#2a2a3a" />
          <XAxis
            dataKey="date"
            stroke="#606070"
            fontSize={10}
            tickLine={false}
            axisLine={false}
          />
          <YAxis
            stroke="#606070"
            fontSize={10}
            tickLine={false}
            axisLine={false}
            tickFormatter={(value) => `${(value / 1000).toFixed(0)}K`}
          />
          <Tooltip content={<CustomTooltip />} />
          <Legend
            iconType="line"
            iconSize={12}
            formatter={(value) => (
              <span className="text-xs text-delphi-text-secondary">{value}</span>
            )}
          />
          <Line
            type="monotone"
            dataKey="input"
            stroke={chartColors.cyan}
            strokeWidth={2}
            dot={false}
            name="Input Tokens"
          />
          <Line
            type="monotone"
            dataKey="output"
            stroke={chartColors.purple}
            strokeWidth={2}
            dot={false}
            name="Output Tokens"
          />
        </LineChart>
      </ResponsiveContainer>
    </motion.div>
  )
}

// =============================================================================
// Heatmap (Simple Grid)
// =============================================================================

interface HeatmapProps {
  data: Array<{ day: string; hour: number; value: number }>
  maxValue?: number
}

export function Heatmap({ data, maxValue }: HeatmapProps) {
  const hours = Array.from({ length: 24 }, (_, i) => i)
  const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  
  const max = maxValue || Math.max(...data.map((d) => d.value), 1)

  const getValue = (day: string, hour: number) => {
    const item = data.find((d) => d.day === day && d.hour === hour)
    return item?.value || 0
  }

  const getColor = (value: number) => {
    const intensity = value / max
    if (intensity === 0) return '#1a1a24'
    if (intensity < 0.25) return 'rgba(74, 158, 255, 0.2)'
    if (intensity < 0.5) return 'rgba(74, 158, 255, 0.4)'
    if (intensity < 0.75) return 'rgba(74, 158, 255, 0.6)'
    return 'rgba(74, 158, 255, 0.9)'
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      transition={{ duration: 0.5 }}
      className="overflow-x-auto"
    >
      <div className="flex gap-1">
        <div className="flex flex-col gap-1 pr-2">
          <div className="h-4" />
          {days.map((day) => (
            <div
              key={day}
              className="h-4 text-2xs text-delphi-text-muted flex items-center"
            >
              {day}
            </div>
          ))}
        </div>
        <div className="flex flex-col gap-1">
          <div className="flex gap-1">
            {hours.map((hour) => (
              <div
                key={hour}
                className="w-4 text-2xs text-delphi-text-muted text-center"
              >
                {hour % 6 === 0 ? hour : ''}
              </div>
            ))}
          </div>
          {days.map((day) => (
            <div key={day} className="flex gap-1">
              {hours.map((hour) => {
                const value = getValue(day, hour)
                return (
                  <div
                    key={`${day}-${hour}`}
                    className="w-4 h-4 rounded-sm transition-colors"
                    style={{ backgroundColor: getColor(value) }}
                    title={`${day} ${hour}:00 - ${value} executions`}
                  />
                )
              })}
            </div>
          ))}
        </div>
      </div>
    </motion.div>
  )
}

// =============================================================================
// Sparkline
// =============================================================================

interface SparklineProps {
  data: number[]
  color?: string
  width?: number
  height?: number
}

export function Sparkline({
  data,
  color = chartColors.blue,
  width = 100,
  height = 30,
}: SparklineProps) {
  const chartData = data.map((value, index) => ({ index, value }))

  return (
    <ResponsiveContainer width={width} height={height}>
      <LineChart data={chartData}>
        <Line
          type="monotone"
          dataKey="value"
          stroke={color}
          strokeWidth={1.5}
          dot={false}
        />
      </LineChart>
    </ResponsiveContainer>
  )
}

