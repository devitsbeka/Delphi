import { useState, useMemo } from 'react'
import { motion } from 'framer-motion'
import {
  ActivityChart,
  CostBreakdownChart,
  chartColors,
} from '../components/Charts'

// Generate mock data
const generateTimeSeriesData = (days: number) => {
  const data = []
  for (let i = days - 1; i >= 0; i--) {
    const date = new Date()
    date.setDate(date.getDate() - i)
    data.push({
      date: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      executions: Math.floor(Math.random() * 500) + 100,
      cost: Math.random() * 30 + 5,
    })
  }
  return data
}

const mockProviderCosts = [
  { name: 'OpenAI', value: 245.80, tokens: 1245000, executions: 342 },
  { name: 'Anthropic', value: 189.50, tokens: 890000, executions: 256 },
  { name: 'Google', value: 67.20, tokens: 456000, executions: 178 },
  { name: 'Ollama', value: 0, tokens: 234000, executions: 89 },
]

const mockAgentCosts = [
  { name: 'Code Review Oracle', cost: 89.50, executions: 156, avgCost: 0.57, trend: 12 },
  { name: 'Bug Fixer', cost: 78.30, executions: 134, avgCost: 0.58, trend: -5 },
  { name: 'Content Writer', cost: 65.20, executions: 89, avgCost: 0.73, trend: 8 },
  { name: 'DevOps Agent', cost: 45.10, executions: 78, avgCost: 0.58, trend: 3 },
  { name: 'Data Analyst', cost: 34.80, executions: 67, avgCost: 0.52, trend: -2 },
  { name: 'Customer Support', cost: 28.60, executions: 45, avgCost: 0.64, trend: 15 },
]

const mockBusinessCosts = [
  { name: 'Mobile Game Studio', cost: 189.50, percentage: 37.8 },
  { name: 'Crash Game Dev', cost: 145.20, percentage: 28.9 },
  { name: 'SaaS Startups', cost: 167.80, percentage: 33.4 },
]

type TimeRange = '7d' | '14d' | '30d' | '90d'

export default function Costs() {
  const [timeRange, setTimeRange] = useState<TimeRange>('30d')

  const timeSeriesData = useMemo(() => {
    const days = timeRange === '7d' ? 7 : timeRange === '14d' ? 14 : timeRange === '30d' ? 30 : 90
    return generateTimeSeriesData(days)
  }, [timeRange])

  const totalCost = mockProviderCosts.reduce((sum, p) => sum + p.value, 0)
  const totalExecutions = mockProviderCosts.reduce((sum, p) => sum + p.executions, 0)
  const totalTokens = mockProviderCosts.reduce((sum, p) => sum + p.tokens, 0)

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-delphi-text-primary">Cost Analytics</h1>
          <p className="text-sm text-delphi-text-muted">
            Monitor and optimize your AI spending
          </p>
        </div>
        <div className="flex gap-2">
          {(['7d', '14d', '30d', '90d'] as TimeRange[]).map((range) => (
            <button
              key={range}
              onClick={() => setTimeRange(range)}
              className={`px-4 py-2 text-sm rounded-lg transition-colors ${
                timeRange === range
                  ? 'bg-delphi-accent text-white'
                  : 'bg-delphi-bg-elevated text-delphi-text-muted hover:text-delphi-text-primary border border-delphi-border'
              }`}
            >
              {range}
            </button>
          ))}
        </div>
      </div>

      {/* Summary Cards */}
      <div className="grid grid-cols-4 gap-4">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="card"
        >
          <p className="text-xs text-delphi-text-muted uppercase tracking-wide">Total Spend</p>
          <p className="text-3xl font-bold font-mono text-delphi-text-primary mt-1">${totalCost.toFixed(2)}</p>
          <p className="text-xs text-delphi-success mt-1">↓ 12% from last period</p>
        </motion.div>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="card"
        >
          <p className="text-xs text-delphi-text-muted uppercase tracking-wide">Avg Cost/Execution</p>
          <p className="text-3xl font-bold font-mono text-delphi-text-primary mt-1">
            ${(totalCost / totalExecutions).toFixed(2)}
          </p>
          <p className="text-xs text-delphi-success mt-1">↓ 8% from last period</p>
        </motion.div>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="card"
        >
          <p className="text-xs text-delphi-text-muted uppercase tracking-wide">Total Executions</p>
          <p className="text-3xl font-bold font-mono text-delphi-text-primary mt-1">
            {totalExecutions.toLocaleString()}
          </p>
          <p className="text-xs text-delphi-accent mt-1">↑ 23% from last period</p>
        </motion.div>
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="card"
        >
          <p className="text-xs text-delphi-text-muted uppercase tracking-wide">Total Tokens</p>
          <p className="text-3xl font-bold font-mono text-delphi-text-primary mt-1">
            {(totalTokens / 1000000).toFixed(2)}M
          </p>
          <p className="text-xs text-delphi-accent mt-1">↑ 18% from last period</p>
        </motion.div>
      </div>

      {/* Charts Row */}
      <div className="grid grid-cols-3 gap-6">
        {/* Cost Over Time */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="col-span-2 card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Cost & Executions Over Time</h3>
          <ActivityChart data={timeSeriesData} height={280} />
        </motion.div>

        {/* Provider Breakdown */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
          className="card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">By Provider</h3>
          <CostBreakdownChart data={mockProviderCosts.map(p => ({ name: p.name, value: p.value }))} height={200} />
          <div className="mt-4 space-y-2">
            {mockProviderCosts.map((provider) => (
              <div key={provider.name} className="flex justify-between text-sm">
                <span className="text-delphi-text-muted">{provider.name}</span>
                <span className="font-mono text-delphi-text-primary">${provider.value.toFixed(2)}</span>
              </div>
            ))}
          </div>
        </motion.div>
      </div>

      {/* Detailed Tables */}
      <div className="grid grid-cols-2 gap-6">
        {/* Cost by Agent */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6 }}
          className="card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Cost by Oracle</h3>
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="text-left text-delphi-text-muted border-b border-delphi-border">
                  <th className="pb-3">Oracle</th>
                  <th className="pb-3 text-right">Executions</th>
                  <th className="pb-3 text-right">Total Cost</th>
                  <th className="pb-3 text-right">Avg Cost</th>
                  <th className="pb-3 text-right">Trend</th>
                </tr>
              </thead>
              <tbody>
                {mockAgentCosts.map((agent) => (
                  <tr key={agent.name} className="border-b border-delphi-border/50">
                    <td className="py-3 text-delphi-text-primary">{agent.name}</td>
                    <td className="py-3 text-right font-mono text-delphi-text-secondary">{agent.executions}</td>
                    <td className="py-3 text-right font-mono text-delphi-text-primary">${agent.cost.toFixed(2)}</td>
                    <td className="py-3 text-right font-mono text-delphi-text-secondary">${agent.avgCost.toFixed(2)}</td>
                    <td className={`py-3 text-right font-mono ${agent.trend > 0 ? 'text-delphi-error' : 'text-delphi-success'}`}>
                      {agent.trend > 0 ? '↑' : '↓'} {Math.abs(agent.trend)}%
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </motion.div>

        {/* Cost by Business */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.7 }}
          className="card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Cost by Business</h3>
          <div className="space-y-4">
            {mockBusinessCosts.map((business) => (
              <div key={business.name}>
                <div className="flex justify-between text-sm mb-1">
                  <span className="text-delphi-text-primary">{business.name}</span>
                  <span className="font-mono text-delphi-text-primary">${business.cost.toFixed(2)}</span>
                </div>
                <div className="h-2 bg-delphi-bg-primary rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-delphi-accent to-delphi-success"
                    style={{ width: `${business.percentage}%` }}
                  />
                </div>
                <p className="text-xs text-delphi-text-muted mt-1">{business.percentage.toFixed(1)}% of total</p>
              </div>
            ))}
          </div>

          <div className="mt-6 pt-4 border-t border-delphi-border">
            <h4 className="text-sm font-semibold text-delphi-text-primary mb-3">Budget Status</h4>
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-delphi-text-muted">Monthly Budget</span>
              <span className="font-mono text-delphi-text-primary">$500.00</span>
            </div>
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-delphi-text-muted">Spent</span>
              <span className="font-mono text-delphi-text-primary">${totalCost.toFixed(2)}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-sm text-delphi-text-muted">Remaining</span>
              <span className="font-mono text-delphi-success">${(500 - totalCost).toFixed(2)}</span>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Token Usage by Provider */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.8 }}
        className="card"
      >
        <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Token Usage by Provider</h3>
        <div className="grid grid-cols-4 gap-6">
          {mockProviderCosts.map((provider) => (
            <div key={provider.name} className="text-center">
              <div className="relative w-24 h-24 mx-auto mb-3">
                <svg className="transform -rotate-90 w-24 h-24">
                  <circle
                    cx="48"
                    cy="48"
                    r="40"
                    stroke="#2a2a3a"
                    strokeWidth="8"
                    fill="none"
                  />
                  <circle
                    cx="48"
                    cy="48"
                    r="40"
                    stroke={chartColors.blue}
                    strokeWidth="8"
                    fill="none"
                    strokeDasharray={`${(provider.tokens / totalTokens) * 251} 251`}
                    strokeLinecap="round"
                  />
                </svg>
                <div className="absolute inset-0 flex items-center justify-center">
                  <span className="text-lg font-mono text-delphi-text-primary">
                    {((provider.tokens / totalTokens) * 100).toFixed(0)}%
                  </span>
                </div>
              </div>
              <p className="font-medium text-delphi-text-primary">{provider.name}</p>
              <p className="text-sm text-delphi-text-muted">{(provider.tokens / 1000).toFixed(0)}K tokens</p>
            </div>
          ))}
        </div>
      </motion.div>
    </div>
  )
}
