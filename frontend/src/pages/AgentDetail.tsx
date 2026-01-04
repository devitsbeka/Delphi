import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import type { Agent, ModelProvider } from '../types'
import { useAgentStore } from '../stores/agents'

const PURPOSES = ['coding', 'content', 'devops', 'analysis', 'support', 'custom'] as const
const PROVIDERS: { id: ModelProvider; name: string; models: string[] }[] = [
  { id: 'openai', name: 'OpenAI', models: ['gpt-4-turbo', 'gpt-4', 'gpt-3.5-turbo'] },
  { id: 'anthropic', name: 'Anthropic', models: ['claude-3-opus', 'claude-3-sonnet', 'claude-3-haiku'] },
  { id: 'google', name: 'Google AI', models: ['gemini-pro', 'gemini-pro-vision'] },
  { id: 'ollama', name: 'Ollama (Local)', models: ['llama2', 'mistral', 'codellama', 'mixtral'] },
]

const defaultSystemPrompts: Record<string, string> = {
  coding: `You are an expert software engineer. Your responsibilities:
- Write clean, maintainable, and well-documented code
- Follow best practices and coding standards
- Consider edge cases and error handling
- Optimize for performance and readability`,
  content: `You are a skilled content writer. Your responsibilities:
- Create engaging and informative content
- Adapt tone and style to the target audience
- Ensure clarity and proper structure
- Include relevant calls-to-action`,
  devops: `You are a DevOps engineer. Your responsibilities:
- Manage infrastructure and deployments
- Write and maintain CI/CD pipelines
- Ensure security and reliability
- Monitor and optimize performance`,
  analysis: `You are a data analyst. Your responsibilities:
- Analyze data to extract insights
- Create clear and actionable reports
- Identify trends and patterns
- Provide data-driven recommendations`,
  support: `You are a customer support specialist. Your responsibilities:
- Resolve customer issues efficiently
- Communicate clearly and empathetically
- Escalate complex issues appropriately
- Document solutions for future reference`,
  custom: '',
}

export function AgentDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { agents, createAgent, updateAgent, deleteAgent } = useAgentStore()
  const isNew = !id || id === 'new'

  const [formData, setFormData] = useState<Partial<Agent>>({
    name: '',
    description: '',
    purpose: 'coding',
    modelProvider: 'openai',
    model: 'gpt-4-turbo',
    systemPrompt: defaultSystemPrompts.coding,
    goal: '',
  })
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    if (!isNew && id) {
      const agent = agents.find(a => a.id === id)
      if (agent) {
        setFormData(agent)
      }
    }
  }, [id, isNew, agents])

  const selectedProvider = PROVIDERS.find(p => p.id === formData.modelProvider)

  const handlePurposeChange = (purpose: typeof PURPOSES[number]) => {
    setFormData({
      ...formData,
      purpose,
      systemPrompt: defaultSystemPrompts[purpose] || '',
    })
  }

  const handleProviderChange = (provider: ModelProvider) => {
    const providerConfig = PROVIDERS.find(p => p.id === provider)
    setFormData({
      ...formData,
      modelProvider: provider,
      model: providerConfig?.models[0] || '',
    })
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSaving(true)

    try {
      if (isNew) {
        await createAgent(formData as Omit<Agent, 'id' | 'createdAt' | 'updatedAt'>)
      } else if (id) {
        await updateAgent(id, formData)
      }
      navigate('/agents')
    } catch (error) {
      console.error('Failed to save agent:', error)
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!id || isNew) return
    
    if (!confirm('Are you sure you want to delete this oracle?')) return
    
    setDeleting(true)
    try {
      await deleteAgent(id)
      navigate('/agents')
    } catch (error) {
      console.error('Failed to delete agent:', error)
    } finally {
      setDeleting(false)
    }
  }

  return (
    <div className="p-6 max-w-4xl mx-auto">
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
      >
        {/* Header */}
        <div className="mb-8">
          <button
            onClick={() => navigate('/agents')}
            className="text-sm text-delphi-text-muted hover:text-delphi-text-primary mb-4 flex items-center gap-1"
          >
            ← Back to Oracles
          </button>
          <h1 className="text-2xl font-bold text-delphi-text-primary">
            {isNew ? 'Create Oracle' : `Edit ${formData.name || 'Oracle'}`}
          </h1>
          <p className="text-sm text-delphi-text-muted mt-1">
            {isNew ? 'Configure a new AI agent for your tasks' : 'Update oracle configuration'}
          </p>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Basic Info */}
          <div className="card">
            <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Basic Information</h3>
            <div className="grid grid-cols-2 gap-4">
              <div className="col-span-2 md:col-span-1">
                <label className="label">Name</label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="e.g., Code Review Oracle"
                  className="input-primary w-full"
                  required
                />
              </div>
              <div className="col-span-2 md:col-span-1">
                <label className="label">Purpose</label>
                <select
                  value={formData.purpose}
                  onChange={(e) => handlePurposeChange(e.target.value as typeof PURPOSES[number])}
                  className="input-primary w-full"
                >
                  {PURPOSES.map((purpose) => (
                    <option key={purpose} value={purpose}>
                      {purpose.charAt(0).toUpperCase() + purpose.slice(1)}
                    </option>
                  ))}
                </select>
              </div>
              <div className="col-span-2">
                <label className="label">Description</label>
                <input
                  type="text"
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  placeholder="Brief description of what this oracle does"
                  className="input-primary w-full"
                />
              </div>
              <div className="col-span-2">
                <label className="label">Goal</label>
                <input
                  type="text"
                  value={formData.goal}
                  onChange={(e) => setFormData({ ...formData, goal: e.target.value })}
                  placeholder="e.g., Review code for quality and best practices"
                  className="input-primary w-full"
                />
              </div>
            </div>
          </div>

          {/* Model Configuration */}
          <div className="card">
            <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Model Configuration</h3>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="label">AI Provider</label>
                <select
                  value={formData.modelProvider}
                  onChange={(e) => handleProviderChange(e.target.value as ModelProvider)}
                  className="input-primary w-full"
                >
                  {PROVIDERS.map((provider) => (
                    <option key={provider.id} value={provider.id}>
                      {provider.name}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="label">Model</label>
                <select
                  value={formData.model}
                  onChange={(e) => setFormData({ ...formData, model: e.target.value })}
                  className="input-primary w-full"
                >
                  {selectedProvider?.models.map((model) => (
                    <option key={model} value={model}>
                      {model}
                    </option>
                  ))}
                </select>
              </div>
            </div>
          </div>

          {/* System Prompt */}
          <div className="card">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-delphi-text-primary">System Prompt</h3>
              <button
                type="button"
                onClick={() => setFormData({ ...formData, systemPrompt: defaultSystemPrompts[formData.purpose || 'custom'] })}
                className="text-xs text-delphi-accent hover:underline"
              >
                Reset to default
              </button>
            </div>
            <textarea
              value={formData.systemPrompt}
              onChange={(e) => setFormData({ ...formData, systemPrompt: e.target.value })}
              placeholder="Instructions for the oracle..."
              className="input-primary w-full h-48 resize-none font-mono text-sm"
            />
            <p className="text-xs text-delphi-text-muted mt-2">
              The system prompt defines the oracle's personality, capabilities, and constraints.
            </p>
          </div>

          {/* Actions */}
          <div className="flex items-center justify-between">
            <div>
              {!isNew && (
                <button
                  type="button"
                  onClick={handleDelete}
                  disabled={deleting}
                  className="px-4 py-2 text-sm text-delphi-error border border-delphi-error/50 rounded-lg hover:bg-delphi-error/10 transition-colors disabled:opacity-50"
                >
                  {deleting ? 'Deleting...' : 'Delete Oracle'}
                </button>
              )}
            </div>
            <div className="flex gap-3">
              <button
                type="button"
                onClick={() => navigate('/agents')}
                className="px-4 py-2 text-sm text-delphi-text-secondary border border-delphi-border rounded-lg hover:border-delphi-accent transition-colors"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={saving || !formData.name}
                className="btn-primary disabled:opacity-50"
              >
                {saving ? 'Saving...' : isNew ? 'Create Oracle' : 'Save Changes'}
              </button>
            </div>
          </div>
        </form>
      </motion.div>
    </div>
  )
}
