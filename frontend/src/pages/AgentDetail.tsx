import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import type { Agent, ModelProvider, AgentPurpose } from '../types'
import useAgentsStore from '../stores/agents'
import { toast } from 'react-hot-toast'

const PURPOSES: AgentPurpose[] = ['coding', 'content', 'devops', 'analysis', 'support', 'custom']
const PROVIDERS: { id: ModelProvider; name: string; models: string[] }[] = [
  { id: 'openai', name: 'OpenAI', models: ['gpt-4o', 'gpt-4-turbo', 'gpt-3.5-turbo'] },
  { id: 'anthropic', name: 'Anthropic', models: ['claude-sonnet-4-20250514', 'claude-3-opus-20240229', 'claude-3-sonnet-20240229'] },
  { id: 'google', name: 'Google AI', models: ['gemini-pro', 'gemini-pro-vision'] },
  { id: 'ollama', name: 'Ollama (Local)', models: ['llama2', 'mistral', 'codellama', 'mixtral'] },
]

const defaultSystemPrompts: Record<AgentPurpose, string> = {
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
  product: `You are a product manager. Your responsibilities:
- Define product vision and strategy
- Prioritize features based on user value
- Coordinate cross-functional teams
- Drive product success metrics`,
}

interface FormData {
  name: string
  description: string
  purpose: AgentPurpose
  model_provider: ModelProvider
  model: string
  system_prompt: string
  goal: string
}

export default function AgentDetail() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const { agents, createAgent, updateAgent, deleteAgent, fetchAgents } = useAgentsStore()
  const isNew = !id || id === 'new'

  const [formData, setFormData] = useState<FormData>({
    name: '',
    description: '',
    purpose: 'coding',
    model_provider: 'openai',
    model: 'gpt-4o',
    system_prompt: defaultSystemPrompts.coding,
    goal: '',
  })
  const [saving, setSaving] = useState(false)
  const [deleting, setDeleting] = useState(false)

  useEffect(() => {
    if (agents.length === 0) {
      fetchAgents()
    }
  }, [agents.length, fetchAgents])

  useEffect(() => {
    if (!isNew && id) {
      const agent = agents.find((a: Agent) => a.id === id)
      if (agent) {
        setFormData({
          name: agent.name,
          description: agent.description || '',
          purpose: agent.purpose,
          model_provider: agent.model_provider,
          model: agent.model,
          system_prompt: agent.system_prompt || '',
          goal: agent.goal || '',
        })
      }
    }
  }, [id, isNew, agents])

  const selectedProvider = PROVIDERS.find(p => p.id === formData.model_provider)

  const handlePurposeChange = (purpose: AgentPurpose) => {
    setFormData({
      ...formData,
      purpose,
      system_prompt: defaultSystemPrompts[purpose] || '',
    })
  }

  const handleProviderChange = (provider: ModelProvider) => {
    const providerInfo = PROVIDERS.find(p => p.id === provider)
    setFormData({
      ...formData,
      model_provider: provider,
      model: providerInfo?.models[0] || '',
    })
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.name.trim()) {
      toast.error('Please enter a name for the oracle')
      return
    }

    setSaving(true)
    try {
      if (isNew) {
        await createAgent(formData as Partial<Agent>)
        toast.success('Oracle created!')
      } else if (id) {
        await updateAgent(id, formData as Partial<Agent>)
        toast.success('Oracle updated!')
      }
      navigate('/agents')
    } catch {
      // Error handled in store
    } finally {
      setSaving(false)
    }
  }

  const handleDelete = async () => {
    if (!id || isNew) return
    if (!window.confirm('Are you sure you want to delete this oracle?')) return

    setDeleting(true)
    try {
      await deleteAgent(id)
      toast.success('Oracle deleted!')
      navigate('/agents')
    } catch {
      // Error handled in store
    } finally {
      setDeleting(false)
    }
  }

  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={{ opacity: 1 }}
      className="p-6 max-w-4xl mx-auto"
    >
      <div className="mb-6">
        <button
          onClick={() => navigate('/agents')}
          className="text-sm text-delphi-text-muted hover:text-delphi-text-primary transition-colors"
        >
          ← Back to Oracles
        </button>
        <h1 className="text-2xl font-bold text-delphi-text-primary mt-2">
          {isNew ? 'Create New Oracle' : `Edit: ${formData.name}`}
        </h1>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Basic Info */}
        <div className="card">
          <h2 className="text-lg font-semibold text-delphi-text-primary mb-4">Basic Information</h2>
          <div className="space-y-4">
            <div>
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
            <div>
              <label className="label">Description</label>
              <input
                type="text"
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                placeholder="What does this oracle do?"
                className="input-primary w-full"
              />
            </div>
            <div>
              <label className="label">Goal</label>
              <input
                type="text"
                value={formData.goal}
                onChange={(e) => setFormData({ ...formData, goal: e.target.value })}
                placeholder="e.g., Review all PRs for code quality"
                className="input-primary w-full"
              />
            </div>
            <div>
              <label className="label">Purpose</label>
              <div className="flex flex-wrap gap-2">
                {PURPOSES.map((purpose) => (
                  <button
                    key={purpose}
                    type="button"
                    onClick={() => handlePurposeChange(purpose)}
                    className={`px-3 py-1.5 rounded-lg text-sm capitalize transition-colors ${
                      formData.purpose === purpose
                        ? 'bg-delphi-accent text-white'
                        : 'bg-delphi-bg-tertiary text-delphi-text-muted hover:text-delphi-text-primary'
                    }`}
                  >
                    {purpose}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Model Configuration */}
        <div className="card">
          <h2 className="text-lg font-semibold text-delphi-text-primary mb-4">Model Configuration</h2>
          <div className="space-y-4">
            <div>
              <label className="label">Provider</label>
              <div className="grid grid-cols-2 md:grid-cols-4 gap-2">
                {PROVIDERS.map((provider) => (
                  <button
                    key={provider.id}
                    type="button"
                    onClick={() => handleProviderChange(provider.id)}
                    className={`p-3 rounded-lg text-sm transition-colors ${
                      formData.model_provider === provider.id
                        ? 'bg-delphi-accent text-white'
                        : 'bg-delphi-bg-tertiary text-delphi-text-muted hover:text-delphi-text-primary'
                    }`}
                  >
                    {provider.name}
                  </button>
                ))}
              </div>
            </div>
            <div>
              <label className="label">Model</label>
              <select
                value={formData.model}
                onChange={(e) => setFormData({ ...formData, model: e.target.value })}
                className="input-primary w-full"
              >
                {selectedProvider?.models.map((model) => (
                  <option key={model} value={model}>{model}</option>
                ))}
              </select>
            </div>
          </div>
        </div>

        {/* System Prompt */}
        <div className="card">
          <h2 className="text-lg font-semibold text-delphi-text-primary mb-4">System Prompt</h2>
          <textarea
            value={formData.system_prompt}
            onChange={(e) => setFormData({ ...formData, system_prompt: e.target.value })}
            placeholder="Instructions for the AI..."
            className="input-primary w-full min-h-[200px] font-mono text-sm"
          />
          <p className="text-xs text-delphi-text-muted mt-2">
            This prompt will be used as the system message when executing tasks.
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
                className="px-4 py-2 text-sm text-red-400 hover:text-red-300 transition-colors disabled:opacity-50"
              >
                {deleting ? 'Deleting...' : 'Delete Oracle'}
              </button>
            )}
          </div>
          <div className="flex gap-3">
            <button
              type="button"
              onClick={() => navigate('/agents')}
              className="btn-secondary"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={saving}
              className="btn-primary disabled:opacity-50"
            >
              {saving ? 'Saving...' : isNew ? 'Create Oracle' : 'Save Changes'}
            </button>
          </div>
        </div>
      </form>
    </motion.div>
  )
}
