import { useState } from 'react'
import { motion } from 'framer-motion'
import { Sparkline, CostBreakdownChart, chartColors } from '../components/Charts'

interface Business {
  id: string
  name: string
  description: string
  industry: string
  status: 'active' | 'paused' | 'archived'
  agents: number
  repositories: number
  monthlyBudget: number
  monthlySpent: number
  executions: number
  createdAt: string
}

const mockBusinesses: Business[] = [
  {
    id: '1', name: 'Mobile Game Studio', description: 'Casual mobile games portfolio with 30+ titles',
    industry: 'Gaming', status: 'active', agents: 8, repositories: 12,
    monthlyBudget: 200, monthlySpent: 145.60, executions: 2456, createdAt: '2024-01-15'
  },
  {
    id: '2', name: 'Crash Game Dev', description: 'Multiplayer crash game and casino-style games',
    industry: 'Gaming', status: 'active', agents: 4, repositories: 3,
    monthlyBudget: 150, monthlySpent: 89.30, executions: 1234, createdAt: '2024-03-20'
  },
  {
    id: '3', name: 'SaaS Startups', description: 'Various B2B SaaS products and services',
    industry: 'Software', status: 'active', agents: 6, repositories: 8,
    monthlyBudget: 250, monthlySpent: 178.90, executions: 1876, createdAt: '2024-02-10'
  },
  {
    id: '4', name: 'AI Consulting', description: 'AI integration consulting services',
    industry: 'Consulting', status: 'paused', agents: 2, repositories: 2,
    monthlyBudget: 100, monthlySpent: 23.50, executions: 345, createdAt: '2024-06-01'
  },
]

const mockFinancials = {
  totalRevenue: 45678.90,
  totalExpenses: 12456.78,
  aiSpending: 502.50,
  profit: 33222.12,
}

export default function Businesses() {
  const [selectedBusiness, setSelectedBusiness] = useState<Business | null>(null)
  const [showCreateModal, setShowCreateModal] = useState(false)

  const totalBudget = mockBusinesses.reduce((sum, b) => sum + b.monthlyBudget, 0)
  const totalSpent = mockBusinesses.reduce((sum, b) => sum + b.monthlySpent, 0)
  const totalAgents = mockBusinesses.reduce((sum, b) => sum + b.agents, 0)
  const totalRepos = mockBusinesses.reduce((sum, b) => sum + b.repositories, 0)

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-delphi-text-primary">Businesses</h1>
          <p className="text-sm text-delphi-text-muted">
            Manage your business units and their AI allocations
          </p>
        </div>
        <button onClick={() => setShowCreateModal(true)} className="btn-primary">
          + Add Business
        </button>
      </div>

      {/* Overview Stats */}
      <div className="grid grid-cols-5 gap-4">
        <div className="card py-4">
          <p className="text-xs text-delphi-text-muted uppercase tracking-wide">Businesses</p>
          <p className="text-2xl font-bold font-mono text-delphi-text-primary">{mockBusinesses.length}</p>
        </div>
        <div className="card py-4">
          <p className="text-xs text-delphi-text-muted uppercase tracking-wide">Total Oracles</p>
          <p className="text-2xl font-bold font-mono text-delphi-text-primary">{totalAgents}</p>
        </div>
        <div className="card py-4">
          <p className="text-xs text-delphi-text-muted uppercase tracking-wide">Repositories</p>
          <p className="text-2xl font-bold font-mono text-delphi-text-primary">{totalRepos}</p>
        </div>
        <div className="card py-4">
          <p className="text-xs text-delphi-text-muted uppercase tracking-wide">Monthly Budget</p>
          <p className="text-2xl font-bold font-mono text-delphi-text-primary">${totalBudget}</p>
        </div>
        <div className="card py-4">
          <p className="text-xs text-delphi-text-muted uppercase tracking-wide">AI Spending (MTD)</p>
          <p className="text-2xl font-bold font-mono text-delphi-success">${totalSpent.toFixed(2)}</p>
        </div>
      </div>

      {/* Financial Overview */}
      <div className="grid grid-cols-3 gap-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="col-span-2 card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">AI Spending by Business</h3>
          <CostBreakdownChart
            data={mockBusinesses.map(b => ({ name: b.name, value: b.monthlySpent }))}
            height={200}
          />
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.1 }}
          className="card"
        >
          <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Financial Summary</h3>
          <div className="space-y-4">
            <div className="flex justify-between items-center">
              <span className="text-sm text-delphi-text-muted">Total Revenue</span>
              <span className="font-mono text-delphi-success">${mockFinancials.totalRevenue.toLocaleString()}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-delphi-text-muted">Total Expenses</span>
              <span className="font-mono text-delphi-text-primary">${mockFinancials.totalExpenses.toLocaleString()}</span>
            </div>
            <div className="flex justify-between items-center">
              <span className="text-sm text-delphi-text-muted">AI Spending</span>
              <span className="font-mono text-delphi-accent">${mockFinancials.aiSpending.toFixed(2)}</span>
            </div>
            <div className="pt-3 border-t border-delphi-border flex justify-between items-center">
              <span className="text-sm text-delphi-text-primary font-medium">Net Profit</span>
              <span className="font-mono text-delphi-success font-bold">${mockFinancials.profit.toLocaleString()}</span>
            </div>
          </div>
        </motion.div>
      </div>

      {/* Business Cards */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {mockBusinesses.map((business, index) => (
          <motion.div
            key={business.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 + index * 0.05 }}
            onClick={() => setSelectedBusiness(business)}
            className="card hover:border-delphi-accent/50 transition-colors cursor-pointer"
          >
            <div className="flex items-start justify-between mb-3">
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold text-delphi-text-primary">{business.name}</h3>
                  <span className={`px-2 py-0.5 text-xs rounded capitalize ${
                    business.status === 'active' ? 'bg-delphi-success/10 text-delphi-success' :
                    business.status === 'paused' ? 'bg-delphi-warning/10 text-delphi-warning' :
                    'bg-delphi-text-muted/10 text-delphi-text-muted'
                  }`}>
                    {business.status}
                  </span>
                </div>
                <p className="text-xs text-delphi-text-muted mt-0.5">{business.industry}</p>
              </div>
              <span className="text-xs text-delphi-text-muted">
                Since {new Date(business.createdAt).toLocaleDateString('en-US', { month: 'short', year: 'numeric' })}
              </span>
            </div>

            <p className="text-sm text-delphi-text-secondary mb-4">{business.description}</p>

            {/* Budget Progress */}
            <div className="mb-4">
              <div className="flex justify-between text-xs mb-1">
                <span className="text-delphi-text-muted">AI Budget</span>
                <span className="text-delphi-text-primary">
                  ${business.monthlySpent.toFixed(2)} / ${business.monthlyBudget}
                </span>
              </div>
              <div className="h-2 bg-delphi-bg-primary rounded-full overflow-hidden">
                <div
                  className={`h-full transition-all ${
                    (business.monthlySpent / business.monthlyBudget) > 0.9 ? 'bg-delphi-error' :
                    (business.monthlySpent / business.monthlyBudget) > 0.7 ? 'bg-delphi-warning' :
                    'bg-delphi-accent'
                  }`}
                  style={{ width: `${Math.min((business.monthlySpent / business.monthlyBudget) * 100, 100)}%` }}
                />
              </div>
            </div>

            {/* Stats */}
            <div className="grid grid-cols-4 gap-4 pt-4 border-t border-delphi-border/50">
              <div className="text-center">
                <p className="font-mono text-delphi-text-primary">{business.agents}</p>
                <p className="text-2xs text-delphi-text-muted">Oracles</p>
              </div>
              <div className="text-center">
                <p className="font-mono text-delphi-text-primary">{business.repositories}</p>
                <p className="text-2xs text-delphi-text-muted">Repos</p>
              </div>
              <div className="text-center">
                <p className="font-mono text-delphi-text-primary">{business.executions}</p>
                <p className="text-2xs text-delphi-text-muted">Executions</p>
              </div>
              <div className="text-center">
                <Sparkline
                  data={Array.from({ length: 12 }, () => Math.random() * 100)}
                  color={chartColors.blue}
                  width={60}
                  height={24}
                />
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Business Detail Modal */}
      {selectedBusiness && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="bg-delphi-bg-elevated border border-delphi-border rounded-xl max-w-2xl w-full max-h-[80vh] overflow-y-auto"
          >
            <div className="p-6 border-b border-delphi-border">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-bold text-delphi-text-primary">{selectedBusiness.name}</h2>
                  <p className="text-sm text-delphi-text-muted">{selectedBusiness.description}</p>
                </div>
                <button
                  onClick={() => setSelectedBusiness(null)}
                  className="text-delphi-text-muted hover:text-delphi-text-primary"
                >
                  ✕
                </button>
              </div>
            </div>

            <div className="p-6 space-y-6">
              {/* Stats Grid */}
              <div className="grid grid-cols-4 gap-4">
                <div className="p-4 bg-delphi-bg-primary rounded-lg text-center">
                  <p className="text-2xl font-mono text-delphi-text-primary">{selectedBusiness.agents}</p>
                  <p className="text-xs text-delphi-text-muted">Oracles</p>
                </div>
                <div className="p-4 bg-delphi-bg-primary rounded-lg text-center">
                  <p className="text-2xl font-mono text-delphi-text-primary">{selectedBusiness.repositories}</p>
                  <p className="text-xs text-delphi-text-muted">Repos</p>
                </div>
                <div className="p-4 bg-delphi-bg-primary rounded-lg text-center">
                  <p className="text-2xl font-mono text-delphi-text-primary">{selectedBusiness.executions}</p>
                  <p className="text-xs text-delphi-text-muted">Executions</p>
                </div>
                <div className="p-4 bg-delphi-bg-primary rounded-lg text-center">
                  <p className="text-2xl font-mono text-delphi-success">${selectedBusiness.monthlySpent.toFixed(2)}</p>
                  <p className="text-xs text-delphi-text-muted">Spent (MTD)</p>
                </div>
              </div>

              {/* Budget Settings */}
              <div>
                <h3 className="text-sm font-semibold text-delphi-text-primary mb-3">Budget Configuration</h3>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs text-delphi-text-muted mb-1">Monthly Budget</label>
                    <input
                      type="number"
                      defaultValue={selectedBusiness.monthlyBudget}
                      className="input-primary w-full"
                    />
                  </div>
                  <div>
                    <label className="block text-xs text-delphi-text-muted mb-1">Alert Threshold (%)</label>
                    <input
                      type="number"
                      defaultValue={80}
                      className="input-primary w-full"
                    />
                  </div>
                </div>
              </div>

              {/* Assigned Oracles */}
              <div>
                <h3 className="text-sm font-semibold text-delphi-text-primary mb-3">Assigned Oracles</h3>
                <div className="space-y-2">
                  {['Code Review Oracle', 'Bug Fixer', 'Content Writer'].slice(0, selectedBusiness.agents).map((oracle, i) => (
                    <div key={i} className="flex items-center justify-between p-3 bg-delphi-bg-primary rounded-lg border border-delphi-border">
                      <span className="text-sm text-delphi-text-primary">{oracle}</span>
                      <button className="text-xs text-delphi-error hover:underline">Remove</button>
                    </div>
                  ))}
                </div>
              </div>

              {/* Actions */}
              <div className="flex gap-3 pt-4 border-t border-delphi-border">
                <button className="btn-primary">Save Changes</button>
                <button className="px-4 py-2 text-sm text-delphi-text-secondary border border-delphi-border rounded-lg hover:border-delphi-accent transition-colors">
                  View Analytics
                </button>
                {selectedBusiness.status === 'active' ? (
                  <button className="px-4 py-2 text-sm text-delphi-warning border border-delphi-warning/50 rounded-lg hover:bg-delphi-warning/10 transition-colors ml-auto">
                    Pause
                  </button>
                ) : (
                  <button className="px-4 py-2 text-sm text-delphi-success border border-delphi-success/50 rounded-lg hover:bg-delphi-success/10 transition-colors ml-auto">
                    Activate
                  </button>
                )}
              </div>
            </div>
          </motion.div>
        </div>
      )}

      {/* Create Business Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="bg-delphi-bg-elevated border border-delphi-border rounded-xl max-w-lg w-full"
          >
            <div className="p-6 border-b border-delphi-border">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-bold text-delphi-text-primary">Add New Business</h2>
                <button
                  onClick={() => setShowCreateModal(false)}
                  className="text-delphi-text-muted hover:text-delphi-text-primary"
                >
                  ✕
                </button>
              </div>
            </div>

            <div className="p-6 space-y-4">
              <div>
                <label className="block text-sm text-delphi-text-secondary mb-2">Business Name</label>
                <input type="text" placeholder="e.g., Mobile Game Studio" className="input-primary w-full" />
              </div>
              <div>
                <label className="block text-sm text-delphi-text-secondary mb-2">Description</label>
                <textarea placeholder="Brief description of the business..." className="input-primary w-full h-20 resize-none" />
              </div>
              <div>
                <label className="block text-sm text-delphi-text-secondary mb-2">Industry</label>
                <select className="input-primary w-full">
                  <option>Gaming</option>
                  <option>Software</option>
                  <option>Consulting</option>
                  <option>E-commerce</option>
                  <option>Other</option>
                </select>
              </div>
              <div>
                <label className="block text-sm text-delphi-text-secondary mb-2">Monthly AI Budget</label>
                <input type="number" placeholder="100" className="input-primary w-full" />
              </div>

              <div className="flex gap-3 pt-4">
                <button className="btn-primary flex-1">Create Business</button>
                <button
                  onClick={() => setShowCreateModal(false)}
                  className="px-4 py-2 text-sm text-delphi-text-secondary border border-delphi-border rounded-lg"
                >
                  Cancel
                </button>
              </div>
            </div>
          </motion.div>
        </div>
      )}
    </div>
  )
}
