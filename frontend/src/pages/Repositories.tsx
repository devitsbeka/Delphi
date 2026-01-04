import { useState } from 'react'
import { motion } from 'framer-motion'

interface Repository {
  id: string
  name: string
  fullName: string
  description: string
  language: string
  defaultBranch: string
  devBranch: string
  stagingBranch: string
  lastSync: string
  indexed: boolean
  businessId: string
  businessName: string
  stats: {
    commits: number
    prs: number
    issues: number
  }
}

const mockRepositories: Repository[] = [
  {
    id: '1', name: 'puzzle-blast', fullName: 'gamedev/puzzle-blast',
    description: 'Puzzle Blast - Match 3 mobile game', language: 'TypeScript',
    defaultBranch: 'main', devBranch: 'dev', stagingBranch: 'staging',
    lastSync: '5 minutes ago', indexed: true, businessId: '1', businessName: 'Mobile Game Studio',
    stats: { commits: 1234, prs: 45, issues: 12 }
  },
  {
    id: '2', name: 'word-master', fullName: 'gamedev/word-master',
    description: 'Word Master - Word puzzle game', language: 'TypeScript',
    defaultBranch: 'main', devBranch: 'development', stagingBranch: 'staging',
    lastSync: '1 hour ago', indexed: true, businessId: '1', businessName: 'Mobile Game Studio',
    stats: { commits: 892, prs: 32, issues: 8 }
  },
  {
    id: '3', name: 'crash-royale', fullName: 'crashgames/crash-royale',
    description: 'Crash Royale - Multiplayer crash game', language: 'Go',
    defaultBranch: 'main', devBranch: 'dev', stagingBranch: 'staging',
    lastSync: '30 minutes ago', indexed: true, businessId: '2', businessName: 'Crash Game Dev',
    stats: { commits: 567, prs: 23, issues: 5 }
  },
  {
    id: '4', name: 'delphi-api', fullName: 'saas/delphi-api',
    description: 'Delphi Platform API', language: 'Go',
    defaultBranch: 'main', devBranch: 'develop', stagingBranch: 'staging',
    lastSync: '2 hours ago', indexed: false, businessId: '3', businessName: 'SaaS Startups',
    stats: { commits: 234, prs: 12, issues: 3 }
  },
]

const languageColors: Record<string, string> = {
  TypeScript: '#3178c6',
  JavaScript: '#f7df1e',
  Go: '#00add8',
  Python: '#3572A5',
  Rust: '#dea584',
}

export default function Repositories() {
  const [search, setSearch] = useState('')
  const [businessFilter, setBusinessFilter] = useState<string>('all')
  const [selectedRepo, setSelectedRepo] = useState<Repository | null>(null)

  const businesses = [...new Set(mockRepositories.map(r => r.businessName))]

  const filteredRepos = mockRepositories.filter(repo => {
    const matchesSearch = 
      repo.name.toLowerCase().includes(search.toLowerCase()) ||
      repo.description.toLowerCase().includes(search.toLowerCase())
    const matchesBusiness = businessFilter === 'all' || repo.businessName === businessFilter
    return matchesSearch && matchesBusiness
  })

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-delphi-text-primary">Repositories</h1>
          <p className="text-sm text-delphi-text-muted">
            Connected GitHub repositories and their sync status
          </p>
        </div>
        <button className="btn-primary">
          + Connect Repository
        </button>
      </div>

      {/* Filters */}
      <div className="card">
        <div className="flex items-center gap-4">
          <input
            type="text"
            placeholder="Search repositories..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="input-primary flex-1"
          />
          <select
            value={businessFilter}
            onChange={(e) => setBusinessFilter(e.target.value)}
            className="input-primary"
          >
            <option value="all">All Businesses</option>
            {businesses.map(b => (
              <option key={b} value={b}>{b}</option>
            ))}
          </select>
        </div>
      </div>

      {/* Repository Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {filteredRepos.map((repo, index) => (
          <motion.div
            key={repo.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.05 }}
            onClick={() => setSelectedRepo(repo)}
            className="card hover:border-delphi-accent/50 transition-colors cursor-pointer"
          >
            <div className="flex items-start justify-between mb-3">
              <div>
                <div className="flex items-center gap-2">
                  <h3 className="font-semibold text-delphi-text-primary">{repo.name}</h3>
                  <span
                    className="w-3 h-3 rounded-full"
                    style={{ backgroundColor: languageColors[repo.language] || '#888' }}
                    title={repo.language}
                  />
                </div>
                <p className="text-xs text-delphi-text-muted mt-0.5">{repo.fullName}</p>
              </div>
              <div className="flex items-center gap-2">
                {repo.indexed ? (
                  <span className="px-2 py-1 text-xs bg-delphi-success/10 text-delphi-success rounded">Indexed</span>
                ) : (
                  <span className="px-2 py-1 text-xs bg-delphi-warning/10 text-delphi-warning rounded">Pending</span>
                )}
              </div>
            </div>

            <p className="text-sm text-delphi-text-secondary mb-4">{repo.description}</p>

            <div className="flex items-center gap-4 text-xs text-delphi-text-muted mb-3">
              <span>{repo.businessName}</span>
              <span>•</span>
              <span>Synced {repo.lastSync}</span>
            </div>

            {/* Branch Strategy */}
            <div className="flex items-center gap-2 mb-4">
              {[
                { label: 'main', branch: repo.defaultBranch },
                { label: 'staging', branch: repo.stagingBranch },
                { label: 'dev', branch: repo.devBranch },
              ].map((b) => (
                <span
                  key={b.label}
                  className="px-2 py-1 text-xs bg-delphi-bg-primary rounded border border-delphi-border"
                >
                  {b.branch}
                </span>
              ))}
            </div>

            {/* Stats */}
            <div className="flex items-center gap-6 pt-3 border-t border-delphi-border/50 text-sm">
              <div className="flex items-center gap-1">
                <span className="text-delphi-text-muted">Commits:</span>
                <span className="font-mono text-delphi-text-primary">{repo.stats.commits}</span>
              </div>
              <div className="flex items-center gap-1">
                <span className="text-delphi-text-muted">PRs:</span>
                <span className="font-mono text-delphi-text-primary">{repo.stats.prs}</span>
              </div>
              <div className="flex items-center gap-1">
                <span className="text-delphi-text-muted">Issues:</span>
                <span className="font-mono text-delphi-text-primary">{repo.stats.issues}</span>
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Repository Detail Modal */}
      {selectedRepo && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="bg-delphi-bg-elevated border border-delphi-border rounded-xl max-w-2xl w-full max-h-[80vh] overflow-y-auto"
          >
            <div className="p-6 border-b border-delphi-border">
              <div className="flex items-center justify-between">
                <div>
                  <h2 className="text-xl font-bold text-delphi-text-primary">{selectedRepo.name}</h2>
                  <p className="text-sm text-delphi-text-muted">{selectedRepo.fullName}</p>
                </div>
                <button
                  onClick={() => setSelectedRepo(null)}
                  className="text-delphi-text-muted hover:text-delphi-text-primary"
                >
                  ✕
                </button>
              </div>
            </div>

            <div className="p-6 space-y-6">
              {/* Description */}
              <div>
                <h3 className="text-sm font-semibold text-delphi-text-primary mb-2">Description</h3>
                <p className="text-sm text-delphi-text-secondary">{selectedRepo.description}</p>
              </div>

              {/* Branch Strategy */}
              <div>
                <h3 className="text-sm font-semibold text-delphi-text-primary mb-2">Branch Strategy</h3>
                <div className="flex items-center gap-2">
                  <div className="flex-1 p-3 bg-delphi-bg-primary rounded-lg border border-delphi-border">
                    <p className="text-xs text-delphi-text-muted mb-1">Main Branch</p>
                    <p className="font-mono text-delphi-text-primary">{selectedRepo.defaultBranch}</p>
                  </div>
                  <span className="text-delphi-text-muted">←</span>
                  <div className="flex-1 p-3 bg-delphi-bg-primary rounded-lg border border-delphi-border">
                    <p className="text-xs text-delphi-text-muted mb-1">Staging Branch</p>
                    <p className="font-mono text-delphi-text-primary">{selectedRepo.stagingBranch}</p>
                  </div>
                  <span className="text-delphi-text-muted">←</span>
                  <div className="flex-1 p-3 bg-delphi-bg-primary rounded-lg border border-delphi-border">
                    <p className="text-xs text-delphi-text-muted mb-1">Dev Branch</p>
                    <p className="font-mono text-delphi-text-primary">{selectedRepo.devBranch}</p>
                  </div>
                </div>
                <p className="text-xs text-delphi-text-muted mt-2">
                  Oracles commit to <code>{selectedRepo.devBranch}</code> → You review and merge to <code>{selectedRepo.stagingBranch}</code> → Deploy to production from <code>{selectedRepo.defaultBranch}</code>
                </p>
              </div>

              {/* Knowledge Base Status */}
              <div>
                <h3 className="text-sm font-semibold text-delphi-text-primary mb-2">Knowledge Base</h3>
                <div className="p-4 bg-delphi-bg-primary rounded-lg border border-delphi-border">
                  <div className="flex items-center justify-between mb-3">
                    <span className="text-sm text-delphi-text-secondary">Indexing Status</span>
                    {selectedRepo.indexed ? (
                      <span className="text-delphi-success text-sm">✓ Fully Indexed</span>
                    ) : (
                      <span className="text-delphi-warning text-sm">◐ Indexing in progress</span>
                    )}
                  </div>
                  <div className="grid grid-cols-3 gap-4 text-center">
                    <div>
                      <p className="text-lg font-mono text-delphi-text-primary">156</p>
                      <p className="text-xs text-delphi-text-muted">Files indexed</p>
                    </div>
                    <div>
                      <p className="text-lg font-mono text-delphi-text-primary">12.5K</p>
                      <p className="text-xs text-delphi-text-muted">Code chunks</p>
                    </div>
                    <div>
                      <p className="text-lg font-mono text-delphi-text-primary">4.2 MB</p>
                      <p className="text-xs text-delphi-text-muted">Vector size</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Recent Activity */}
              <div>
                <h3 className="text-sm font-semibold text-delphi-text-primary mb-2">Recent Oracle Activity</h3>
                <div className="space-y-2">
                  {[
                    { agent: 'Code Review Oracle', action: 'Reviewed PR #45', time: '2 hours ago' },
                    { agent: 'Bug Fixer', action: 'Created PR #44', time: '5 hours ago' },
                    { agent: 'Code Review Oracle', action: 'Reviewed PR #43', time: '1 day ago' },
                  ].map((activity, i) => (
                    <div key={i} className="flex items-center justify-between p-2 bg-delphi-bg-primary rounded">
                      <div>
                        <span className="text-sm text-delphi-accent">{activity.agent}</span>
                        <span className="text-sm text-delphi-text-secondary ml-2">{activity.action}</span>
                      </div>
                      <span className="text-xs text-delphi-text-muted">{activity.time}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Actions */}
              <div className="flex gap-3 pt-4 border-t border-delphi-border">
                <button className="btn-primary">Sync Now</button>
                <button className="px-4 py-2 text-sm text-delphi-text-secondary border border-delphi-border rounded-lg hover:border-delphi-accent transition-colors">
                  Reindex
                </button>
                <button className="px-4 py-2 text-sm text-delphi-text-secondary border border-delphi-border rounded-lg hover:border-delphi-accent transition-colors">
                  View on GitHub
                </button>
              </div>
            </div>
          </motion.div>
        </div>
      )}
    </div>
  )
}
