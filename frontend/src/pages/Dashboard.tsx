import { useEffect, useState } from 'react'
import { motion } from 'framer-motion'
import { Link } from 'react-router-dom'
import {
  ActivityChart,
  CostBreakdownChart,
  AgentStatusChart,
  TokenUsageChart,
  Heatmap,
  Sparkline,
  chartColors,
} from '../components/Charts'
import StatusBadge from '../components/StatusBadge'
import MetricCard from '../components/MetricCard'
import useAgentsStore from '../stores/agents'

// Mock data for demonstration
const mockActivityData = Array.from({ length: 14 }, (_, i) => ({
  date: new Date(Date.now() - (13 - i) * 24 * 60 * 60 * 1000).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
  executions: Math.floor(Math.random() * 500) + 100,
  cost: Math.random() * 20 + 5,
}))

const mockCostBreakdown = [
  { name: 'OpenAI', value: 245.80 },
  { name: 'Anthropic', value: 189.50 },
  { name: 'Google', value: 67.20 },
  { name: 'Ollama', value: 0 },
]

const mockAgentStatus = [
  { status: 'Ready', count: 12 },
  { status: 'Executing', count: 5 },
  { status: 'Briefing', count: 2 },
  { status: 'Paused', count: 3 },
  { status: 'Error', count: 1 },
]

const mockTokenUsage = Array.from({ length: 7 }, (_, i) => ({
  date: new Date(Date.now() - (6 - i) * 24 * 60 * 60 * 1000).toLocaleDateString('en-US', { weekday: 'short' }),
  input: Math.floor(Math.random() * 50000) + 20000,
  output: Math.floor(Math.random() * 30000) + 10000,
}))

const mockHeatmapData = (() => {
  const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
  const data: Array<{ day: string; hour: number; value: number }> = []
  days.forEach((day) => {
    for (let hour = 0; hour < 24; hour++) {
      // More activity during work hours
      const isWorkHour = hour >= 9 && hour <= 18
      const isWeekday = day !== 'Sun' && day !== 'Sat'
      const baseValue = isWorkHour && isWeekday ? 50 : 10
      data.push({
        day,
        hour,
        value: Math.floor(Math.random() * baseValue) + (isWorkHour ? 20 : 0),
      })
    }
  })
  return data
})()

const mockRecentExecutions = [
  { id: '1', agent: 'Code Review Oracle', task: 'Review PR #142', status: 'completed', duration: '2m 34s', cost: '$0.24' },
  { id: '2', agent: 'Content Writer', task: 'Blog post draft', status: 'executing', duration: '5m 12s', cost: '$0.89' },
  { id: '3', agent: 'Bug Fixer', task: 'Fix login issue', status: 'completed', duration: '8m 45s', cost: '$1.23' },
  { id: '4', agent: 'DevOps Agent', task: 'Deploy staging', status: 'failed', duration: '1m 02s', cost: '$0.12' },
  { id: '5', agent: 'Data Analyst', task: 'Monthly report', status: 'briefing', duration: '0m 45s', cost: '$0.05' },
]

const mockSparklineData = [12, 15, 8, 22, 18, 25, 20, 28, 24, 30, 26, 32]

export default function Dashboard() {
  const { fetchAgents } = useAgentsStore()
  const [currentTime, setCurrentTime] = useState(new Date())

  useEffect(() => {
    fetchAgents()
    const timer = setInterval(() => setCurrentTime(new Date()), 1000)
    return () => clearInterval(timer)
  }, [fetchAgents])

  const stats = {
    totalAgents: 23,
    activeAgents: 7,
    todayExecutions: 847,
    todayCost: 45.67,
    monthlyBudget: 500,
    monthlySpent: 342.50,
  }

  return (
    <div className="space-y-6 p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-delphi-text-primary">Command Center</h1>
          <p className="text-sm text-delphi-text-muted">
            {currentTime.toLocaleDateString('en-US', { 
              weekday: 'long', 
              year: 'numeric', 
              month: 'long', 
              day: 'numeric' 
            })}
            {' • '}
            <span className="text-delphi-accent">{currentTime.toLocaleTimeString()}</span>
          </p>
        </div>
        <div className="flex items-center gap-4">
          <div className="flex items-center gap-2 px-4 py-2 bg-delphi-bg-elevated border border-delphi-border rounded-lg">
            <span className="w-2 h-2 bg-delphi-success rounded-full animate-pulse" />
            <span className="text-sm text-delphi-text-secondary">All Systems Operational</span>
          </div>
          <Link
            to="/execute"
            className="btn-primary"
          >
            Launch Oracle
          </Link>
        </div>
      </div>

      {/* Top Metrics */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <MetricCard
          title="Active Oracles"
          value={stats.activeAgents}
          subtitle={`of ${stats.totalAgents} total`}
          trend={{ value: 12, direction: 'up' }}
          sparkline={mockSparklineData}
        />
        <MetricCard
          title="Today's Executions"
          value={stats.todayExecutions}
          subtitle="across all agents"
          trend={{ value: 23, direction: 'up' }}
          sparkline={mockSparklineData.map(v => v * 30)}
        />
        <MetricCard
          title="Today's Cost"
          value={`$${stats.todayCost.toFixed(2)}`}
          subtitle="API usage"
          trend={{ value: 8, direction: 'down' }}
          sparkline={mockSparklineData.map(v => v * 1.5)}
        />
        <MetricCard
          title="Monthly Budget"
          value={`${((stats.monthlySpent / stats.monthlyBudget) * 100).toFixed(0)}%`}
          subtitle={`$${stats.monthlySpent.toFixed(2)} of $${stats.monthlyBudget}`}
          trend={{ value: 5, direction: 'neutral' }}
          sparkline={mockSparklineData.map(v => v * 10)}
        />
      </div>

      {/* Main Charts Row */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Activity Chart */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="lg:col-span-2 card"
        >
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-delphi-text-primary">Activity Overview</h3>
            <div className="flex gap-2">
              {['7D', '14D', '30D'].map((period) => (
                <button
                  key={period}
                  className="px-3 py-1 text-xs text-delphi-text-muted hover:text-delphi-text-primary
                             bg-delphi-bg-primary rounded transition-colors"
                >
                  {period}
                </button>
              ))}
            </div>
          </div>
          <ActivityChart data={mockActivityData} height={240} />
        </motion.div>

        {/* Cost Breakdown */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Cost by Provider</h3>
          <CostBreakdownChart data={mockCostBreakdown} height={200} />
          <div className="mt-4 pt-4 border-t border-delphi-border">
            <div className="flex justify-between text-sm">
              <span className="text-delphi-text-muted">Total (30d)</span>
              <span className="text-delphi-text-primary font-mono">
                ${mockCostBreakdown.reduce((sum, item) => sum + item.value, 0).toFixed(2)}
              </span>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Second Row */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Agent Status */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Oracle Status</h3>
          <AgentStatusChart data={mockAgentStatus} height={160} />
        </motion.div>

        {/* Token Usage */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.4 }}
          className="lg:col-span-2 card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Token Usage (7d)</h3>
          <TokenUsageChart data={mockTokenUsage} height={160} />
        </motion.div>

        {/* Quick Stats */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.5 }}
          className="card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Quick Stats</h3>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <span className="text-sm text-delphi-text-muted">PRs Created</span>
              <span className="text-delphi-text-primary font-mono">42</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-delphi-text-muted">Commits</span>
              <span className="text-delphi-text-primary font-mono">187</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-delphi-text-muted">KB Queries</span>
              <span className="text-delphi-text-primary font-mono">2.4K</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-delphi-text-muted">Avg Response</span>
              <span className="text-delphi-text-primary font-mono">1.2s</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-delphi-text-muted">Success Rate</span>
              <span className="text-delphi-success font-mono">98.7%</span>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Bottom Row */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Activity Heatmap */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.6 }}
          className="card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Execution Heatmap</h3>
          <Heatmap data={mockHeatmapData} />
          <div className="mt-4 flex items-center justify-between text-xs text-delphi-text-muted">
            <span>Less</span>
            <div className="flex gap-1">
              {['#1a1a24', 'rgba(74, 158, 255, 0.2)', 'rgba(74, 158, 255, 0.4)', 'rgba(74, 158, 255, 0.6)', 'rgba(74, 158, 255, 0.9)'].map((color, i) => (
                <div key={i} className="w-3 h-3 rounded-sm" style={{ backgroundColor: color }} />
              ))}
            </div>
            <span>More</span>
          </div>
        </motion.div>

        {/* Recent Executions */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.7 }}
          className="card"
        >
          <div className="flex items-center justify-between mb-4">
            <h3 className="text-lg font-semibold text-delphi-text-primary">Recent Executions</h3>
            <Link to="/executions" className="text-sm text-delphi-accent hover:underline">
              View all →
            </Link>
          </div>
          <div className="space-y-3">
            {mockRecentExecutions.map((exec) => (
              <div
                key={exec.id}
                className="flex items-center justify-between p-3 bg-delphi-bg-primary rounded-lg
                           border border-delphi-border/50 hover:border-delphi-border transition-colors"
              >
                <div className="flex items-center gap-3">
                  <StatusBadge status={exec.status} />
                  <div>
                    <p className="text-sm font-medium text-delphi-text-primary">{exec.agent}</p>
                    <p className="text-xs text-delphi-text-muted">{exec.task}</p>
                  </div>
                </div>
                <div className="text-right">
                  <p className="text-sm font-mono text-delphi-text-secondary">{exec.duration}</p>
                  <p className="text-xs text-delphi-text-muted">{exec.cost}</p>
                </div>
              </div>
            ))}
          </div>
        </motion.div>
      </div>

      {/* Businesses Overview */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.8 }}
        className="card"
      >
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-delphi-text-primary">Business Overview</h3>
          <Link to="/businesses" className="text-sm text-delphi-accent hover:underline">
            Manage →
          </Link>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          {[
            { name: 'Mobile Game Studio', agents: 8, repos: 12, executions: 342, cost: 89.50 },
            { name: 'Crash Game Dev', agents: 4, repos: 3, executions: 156, cost: 45.20 },
            { name: 'SaaS Startups', agents: 6, repos: 8, executions: 287, cost: 78.30 },
          ].map((business) => (
            <div
              key={business.name}
              className="p-4 bg-delphi-bg-primary rounded-lg border border-delphi-border/50
                         hover:border-delphi-accent/50 transition-colors cursor-pointer"
            >
              <h4 className="font-medium text-delphi-text-primary mb-3">{business.name}</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <div>
                  <p className="text-delphi-text-muted">Oracles</p>
                  <p className="text-delphi-text-primary font-mono">{business.agents}</p>
                </div>
                <div>
                  <p className="text-delphi-text-muted">Repos</p>
                  <p className="text-delphi-text-primary font-mono">{business.repos}</p>
                </div>
                <div>
                  <p className="text-delphi-text-muted">Executions</p>
                  <p className="text-delphi-text-primary font-mono">{business.executions}</p>
                </div>
                <div>
                  <p className="text-delphi-text-muted">Cost (30d)</p>
                  <p className="text-delphi-text-primary font-mono">${business.cost}</p>
                </div>
              </div>
              <div className="mt-3 pt-3 border-t border-delphi-border/50">
                <Sparkline
                  data={Array.from({ length: 12 }, () => Math.random() * 100)}
                  color={chartColors.blue}
                  width={150}
                  height={24}
                />
              </div>
            </div>
          ))}
        </div>
      </motion.div>
    </div>
  )
}
