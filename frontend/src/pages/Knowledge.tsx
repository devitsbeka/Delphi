import { useState } from 'react'
import { motion } from 'framer-motion'

interface KnowledgeBase {
  id: string
  name: string
  description: string
  type: 'repository' | 'document' | 'api' | 'custom'
  documentCount: number
  vectorCount: number
  lastUpdated: string
  size: string
  businessId: string
  businessName: string
}

interface Document {
  id: string
  name: string
  type: string
  chunks: number
  addedAt: string
  source: string
}

const mockKnowledgeBases: KnowledgeBase[] = [
  {
    id: '1', name: 'Mobile Games Codebase', description: 'Code and documentation from all mobile game repositories',
    type: 'repository', documentCount: 1245, vectorCount: 45000, lastUpdated: '5 minutes ago', size: '156 MB',
    businessId: '1', businessName: 'Mobile Game Studio'
  },
  {
    id: '2', name: 'Game Design Documents', description: 'GDDs, specs, and design documentation',
    type: 'document', documentCount: 89, vectorCount: 12000, lastUpdated: '1 hour ago', size: '24 MB',
    businessId: '1', businessName: 'Mobile Game Studio'
  },
  {
    id: '3', name: 'Crash Game Backend', description: 'Backend services and API documentation',
    type: 'repository', documentCount: 234, vectorCount: 18000, lastUpdated: '30 minutes ago', size: '45 MB',
    businessId: '2', businessName: 'Crash Game Dev'
  },
  {
    id: '4', name: 'Company Policies', description: 'HR policies, guidelines, and procedures',
    type: 'document', documentCount: 45, vectorCount: 3000, lastUpdated: '2 days ago', size: '8 MB',
    businessId: '3', businessName: 'SaaS Startups'
  },
]

const mockDocuments: Document[] = [
  { id: '1', name: 'README.md', type: 'markdown', chunks: 12, addedAt: '2 hours ago', source: 'puzzle-blast' },
  { id: '2', name: 'game_manager.ts', type: 'typescript', chunks: 45, addedAt: '2 hours ago', source: 'puzzle-blast' },
  { id: '3', name: 'API Documentation.pdf', type: 'pdf', chunks: 156, addedAt: '1 day ago', source: 'upload' },
  { id: '4', name: 'Design Spec v2.docx', type: 'docx', chunks: 89, addedAt: '3 days ago', source: 'upload' },
  { id: '5', name: 'config.go', type: 'go', chunks: 23, addedAt: '5 hours ago', source: 'crash-royale' },
]

const typeIcons: Record<string, string> = {
  repository: '📁',
  document: '📄',
  api: '🔌',
  custom: '⚙️',
}

export default function Knowledge() {
  const [selectedKB, setSelectedKB] = useState<KnowledgeBase | null>(null)
  const [search, setSearch] = useState('')
  const [uploadModalOpen, setUploadModalOpen] = useState(false)

  const filteredKBs = mockKnowledgeBases.filter(kb =>
    kb.name.toLowerCase().includes(search.toLowerCase()) ||
    kb.description.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-delphi-text-primary">Knowledge Base</h1>
          <p className="text-sm text-delphi-text-muted">
            Central repository of context for your AI oracles
          </p>
        </div>
        <div className="flex gap-3">
          <button
            onClick={() => setUploadModalOpen(true)}
            className="px-4 py-2 text-sm text-delphi-accent border border-delphi-accent rounded-lg hover:bg-delphi-accent hover:text-white transition-colors"
          >
            Upload Documents
          </button>
          <button className="btn-primary">
            + Create Knowledge Base
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        {[
          { label: 'Knowledge Bases', value: mockKnowledgeBases.length },
          { label: 'Total Documents', value: mockKnowledgeBases.reduce((s, k) => s + k.documentCount, 0).toLocaleString() },
          { label: 'Vector Embeddings', value: (mockKnowledgeBases.reduce((s, k) => s + k.vectorCount, 0) / 1000).toFixed(0) + 'K' },
          { label: 'Total Size', value: '233 MB' },
        ].map((stat) => (
          <div key={stat.label} className="card py-4">
            <p className="text-xs text-delphi-text-muted uppercase tracking-wide">{stat.label}</p>
            <p className="text-2xl font-bold font-mono text-delphi-text-primary">{stat.value}</p>
          </div>
        ))}
      </div>

      {/* Search */}
      <div className="card">
        <input
          type="text"
          placeholder="Search knowledge bases..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="input-primary w-full"
        />
      </div>

      {/* Knowledge Base Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
        {filteredKBs.map((kb, index) => (
          <motion.div
            key={kb.id}
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: index * 0.05 }}
            onClick={() => setSelectedKB(kb)}
            className="card hover:border-delphi-accent/50 transition-colors cursor-pointer"
          >
            <div className="flex items-start gap-4">
              <div className="text-3xl">{typeIcons[kb.type]}</div>
              <div className="flex-1">
                <div className="flex items-center justify-between mb-1">
                  <h3 className="font-semibold text-delphi-text-primary">{kb.name}</h3>
                  <span className="px-2 py-0.5 text-xs bg-delphi-bg-primary rounded border border-delphi-border capitalize">
                    {kb.type}
                  </span>
                </div>
                <p className="text-sm text-delphi-text-secondary mb-3">{kb.description}</p>
                <p className="text-xs text-delphi-text-muted mb-3">{kb.businessName}</p>
                
                <div className="grid grid-cols-4 gap-2 text-center">
                  <div>
                    <p className="font-mono text-delphi-text-primary">{kb.documentCount}</p>
                    <p className="text-2xs text-delphi-text-muted">Documents</p>
                  </div>
                  <div>
                    <p className="font-mono text-delphi-text-primary">{(kb.vectorCount / 1000).toFixed(0)}K</p>
                    <p className="text-2xs text-delphi-text-muted">Vectors</p>
                  </div>
                  <div>
                    <p className="font-mono text-delphi-text-primary">{kb.size}</p>
                    <p className="text-2xs text-delphi-text-muted">Size</p>
                  </div>
                  <div>
                    <p className="font-mono text-delphi-text-primary">{kb.lastUpdated}</p>
                    <p className="text-2xs text-delphi-text-muted">Updated</p>
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        ))}
      </div>

      {/* Query Interface */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ delay: 0.3 }}
        className="card"
      >
        <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Query Knowledge Base</h3>
        <div className="flex gap-4">
          <input
            type="text"
            placeholder="Ask a question about your codebase or documents..."
            className="input-primary flex-1"
          />
          <button className="btn-primary">Search</button>
        </div>
        <p className="text-xs text-delphi-text-muted mt-2">
          Uses semantic search to find relevant context from all connected knowledge bases
        </p>
      </motion.div>

      {/* Knowledge Base Detail Modal */}
      {selectedKB && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="bg-delphi-bg-elevated border border-delphi-border rounded-xl max-w-3xl w-full max-h-[80vh] overflow-y-auto"
          >
            <div className="p-6 border-b border-delphi-border">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <span className="text-3xl">{typeIcons[selectedKB.type]}</span>
                  <div>
                    <h2 className="text-xl font-bold text-delphi-text-primary">{selectedKB.name}</h2>
                    <p className="text-sm text-delphi-text-muted">{selectedKB.description}</p>
                  </div>
                </div>
                <button
                  onClick={() => setSelectedKB(null)}
                  className="text-delphi-text-muted hover:text-delphi-text-primary"
                >
                  ✕
                </button>
              </div>
            </div>

            <div className="p-6 space-y-6">
              {/* Stats */}
              <div className="grid grid-cols-4 gap-4">
                <div className="p-4 bg-delphi-bg-primary rounded-lg text-center">
                  <p className="text-2xl font-mono text-delphi-text-primary">{selectedKB.documentCount}</p>
                  <p className="text-xs text-delphi-text-muted">Documents</p>
                </div>
                <div className="p-4 bg-delphi-bg-primary rounded-lg text-center">
                  <p className="text-2xl font-mono text-delphi-text-primary">{(selectedKB.vectorCount / 1000).toFixed(0)}K</p>
                  <p className="text-xs text-delphi-text-muted">Vectors</p>
                </div>
                <div className="p-4 bg-delphi-bg-primary rounded-lg text-center">
                  <p className="text-2xl font-mono text-delphi-text-primary">{selectedKB.size}</p>
                  <p className="text-xs text-delphi-text-muted">Size</p>
                </div>
                <div className="p-4 bg-delphi-bg-primary rounded-lg text-center">
                  <p className="text-2xl font-mono text-delphi-text-primary">{selectedKB.lastUpdated}</p>
                  <p className="text-xs text-delphi-text-muted">Last Update</p>
                </div>
              </div>

              {/* Documents */}
              <div>
                <h3 className="text-sm font-semibold text-delphi-text-primary mb-3">Recent Documents</h3>
                <div className="space-y-2">
                  {mockDocuments.map((doc) => (
                    <div
                      key={doc.id}
                      className="flex items-center justify-between p-3 bg-delphi-bg-primary rounded-lg border border-delphi-border"
                    >
                      <div className="flex items-center gap-3">
                        <span className="w-8 h-8 flex items-center justify-center bg-delphi-bg-elevated rounded text-sm">
                          {doc.type === 'markdown' ? '📝' :
                           doc.type === 'typescript' ? '📘' :
                           doc.type === 'go' ? '🔵' :
                           doc.type === 'pdf' ? '📕' : '📄'}
                        </span>
                        <div>
                          <p className="text-sm text-delphi-text-primary">{doc.name}</p>
                          <p className="text-xs text-delphi-text-muted">{doc.source} • {doc.chunks} chunks</p>
                        </div>
                      </div>
                      <span className="text-xs text-delphi-text-muted">{doc.addedAt}</span>
                    </div>
                  ))}
                </div>
              </div>

              {/* Actions */}
              <div className="flex gap-3 pt-4 border-t border-delphi-border">
                <button className="btn-primary">Sync Now</button>
                <button className="px-4 py-2 text-sm text-delphi-text-secondary border border-delphi-border rounded-lg hover:border-delphi-accent transition-colors">
                  Rebuild Index
                </button>
                <button className="px-4 py-2 text-sm text-delphi-error border border-delphi-error/50 rounded-lg hover:bg-delphi-error/10 transition-colors ml-auto">
                  Delete
                </button>
              </div>
            </div>
          </motion.div>
        </div>
      )}

      {/* Upload Modal */}
      {uploadModalOpen && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="bg-delphi-bg-elevated border border-delphi-border rounded-xl max-w-lg w-full"
          >
            <div className="p-6 border-b border-delphi-border">
              <div className="flex items-center justify-between">
                <h2 className="text-xl font-bold text-delphi-text-primary">Upload Documents</h2>
                <button
                  onClick={() => setUploadModalOpen(false)}
                  className="text-delphi-text-muted hover:text-delphi-text-primary"
                >
                  ✕
                </button>
              </div>
            </div>

            <div className="p-6 space-y-4">
              <div className="border-2 border-dashed border-delphi-border rounded-lg p-8 text-center hover:border-delphi-accent transition-colors cursor-pointer">
                <p className="text-3xl mb-2">📤</p>
                <p className="text-delphi-text-primary mb-1">Drop files here or click to upload</p>
                <p className="text-xs text-delphi-text-muted">Supports PDF, DOCX, MD, TXT, and code files</p>
              </div>

              <div>
                <label className="block text-sm text-delphi-text-secondary mb-2">Target Knowledge Base</label>
                <select className="input-primary w-full">
                  {mockKnowledgeBases.map(kb => (
                    <option key={kb.id} value={kb.id}>{kb.name}</option>
                  ))}
                </select>
              </div>

              <div className="flex gap-3 pt-4">
                <button className="btn-primary flex-1">Upload</button>
                <button
                  onClick={() => setUploadModalOpen(false)}
                  className="px-4 py-2 text-sm text-delphi-text-secondary border border-delphi-border rounded-lg hover:border-delphi-accent transition-colors"
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
