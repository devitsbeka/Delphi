import { useState } from 'react'
import { motion } from 'framer-motion'

interface APIKeyConfig {
  id: string
  provider: 'openai' | 'anthropic' | 'google' | 'ollama'
  label: string
  key: string
  isSet: boolean
  lastUsed?: string
  usageThisMonth?: number
}

const mockAPIKeys: APIKeyConfig[] = [
  { id: '1', provider: 'openai', label: 'OpenAI', key: '', isSet: true, lastUsed: '2 hours ago', usageThisMonth: 245.80 },
  { id: '2', provider: 'anthropic', label: 'Anthropic', key: '', isSet: true, lastUsed: '1 day ago', usageThisMonth: 189.50 },
  { id: '3', provider: 'google', label: 'Google AI', key: '', isSet: false },
  { id: '4', provider: 'ollama', label: 'Ollama (Local)', key: '', isSet: true, lastUsed: '3 hours ago', usageThisMonth: 0 },
]

interface NotificationSetting {
  id: string
  label: string
  description: string
  email: boolean
  slack: boolean
  discord: boolean
}

const mockNotifications: NotificationSetting[] = [
  { id: '1', label: 'Execution Complete', description: 'When an oracle finishes a task', email: false, slack: true, discord: true },
  { id: '2', label: 'Execution Failed', description: 'When an oracle encounters an error', email: true, slack: true, discord: true },
  { id: '3', label: 'Budget Alert', description: 'When spending reaches 80% of budget', email: true, slack: true, discord: false },
  { id: '4', label: 'PR Created', description: 'When an oracle creates a pull request', email: false, slack: true, discord: false },
  { id: '5', label: 'Weekly Digest', description: 'Summary of weekly activity', email: true, slack: false, discord: false },
]

export function Settings() {
  const [activeTab, setActiveTab] = useState<'api-keys' | 'notifications' | 'security' | 'billing'>('api-keys')
  const [apiKeys, setApiKeys] = useState(mockAPIKeys)
  const [notifications, setNotifications] = useState(mockNotifications)
  const [editingKey, setEditingKey] = useState<string | null>(null)
  const [keyValue, setKeyValue] = useState('')

  const tabs = [
    { id: 'api-keys', label: 'API Keys' },
    { id: 'notifications', label: 'Notifications' },
    { id: 'security', label: 'Security' },
    { id: 'billing', label: 'Billing' },
  ] as const

  const handleSaveKey = (id: string) => {
    setApiKeys(prev => prev.map(key => 
      key.id === id ? { ...key, key: keyValue, isSet: keyValue.length > 0 } : key
    ))
    setEditingKey(null)
    setKeyValue('')
  }

  const handleDeleteKey = (id: string) => {
    setApiKeys(prev => prev.map(key => 
      key.id === id ? { ...key, key: '', isSet: false } : key
    ))
  }

  const toggleNotification = (id: string, channel: 'email' | 'slack' | 'discord') => {
    setNotifications(prev => prev.map(notif => 
      notif.id === id ? { ...notif, [channel]: !notif[channel] } : notif
    ))
  }

  return (
    <div className="p-6 max-w-5xl mx-auto">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-delphi-text-primary">Settings</h1>
        <p className="text-sm text-delphi-text-muted">
          Configure your Delphi platform
        </p>
      </div>

      {/* Tabs */}
      <div className="flex gap-1 mb-6 bg-delphi-bg-elevated p-1 rounded-lg border border-delphi-border w-fit">
        {tabs.map((tab) => (
          <button
            key={tab.id}
            onClick={() => setActiveTab(tab.id)}
            className={`px-4 py-2 text-sm rounded-md transition-colors ${
              activeTab === tab.id
                ? 'bg-delphi-accent text-white'
                : 'text-delphi-text-muted hover:text-delphi-text-primary'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <motion.div
        key={activeTab}
        initial={{ opacity: 0, y: 10 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.2 }}
      >
        {/* API Keys Tab */}
        {activeTab === 'api-keys' && (
          <div className="space-y-6">
            <div className="card">
              <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">AI Provider API Keys</h3>
              <p className="text-sm text-delphi-text-muted mb-6">
                Connect your AI provider accounts to power your oracles. Keys are encrypted and stored securely.
              </p>

              <div className="space-y-4">
                {apiKeys.map((apiKey) => (
                  <div
                    key={apiKey.id}
                    className="p-4 bg-delphi-bg-primary rounded-lg border border-delphi-border"
                  >
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-4">
                        <div className={`w-10 h-10 rounded-lg flex items-center justify-center text-white font-bold ${
                          apiKey.provider === 'openai' ? 'bg-emerald-600' :
                          apiKey.provider === 'anthropic' ? 'bg-orange-600' :
                          apiKey.provider === 'google' ? 'bg-blue-600' :
                          'bg-gray-600'
                        }`}>
                          {apiKey.label[0]}
                        </div>
                        <div>
                          <p className="font-medium text-delphi-text-primary">{apiKey.label}</p>
                          <p className="text-xs text-delphi-text-muted">
                            {apiKey.isSet ? (
                              <>
                                <span className="text-delphi-success">● Connected</span>
                                {apiKey.lastUsed && <span> • Last used {apiKey.lastUsed}</span>}
                                {apiKey.usageThisMonth !== undefined && (
                                  <span> • ${apiKey.usageThisMonth.toFixed(2)} this month</span>
                                )}
                              </>
                            ) : (
                              <span className="text-delphi-text-muted">● Not configured</span>
                            )}
                          </p>
                        </div>
                      </div>

                      <div className="flex items-center gap-2">
                        {editingKey === apiKey.id ? (
                          <>
                            <input
                              type="password"
                              value={keyValue}
                              onChange={(e) => setKeyValue(e.target.value)}
                              placeholder={`Enter ${apiKey.label} API key`}
                              className="input-primary w-64"
                              autoFocus
                            />
                            <button
                              onClick={() => handleSaveKey(apiKey.id)}
                              className="btn-primary text-sm"
                            >
                              Save
                            </button>
                            <button
                              onClick={() => { setEditingKey(null); setKeyValue(''); }}
                              className="px-3 py-2 text-sm text-delphi-text-muted hover:text-delphi-text-primary"
                            >
                              Cancel
                            </button>
                          </>
                        ) : (
                          <>
                            <button
                              onClick={() => setEditingKey(apiKey.id)}
                              className="px-3 py-2 text-sm text-delphi-accent border border-delphi-accent rounded
                                       hover:bg-delphi-accent hover:text-white transition-colors"
                            >
                              {apiKey.isSet ? 'Update' : 'Configure'}
                            </button>
                            {apiKey.isSet && (
                              <button
                                onClick={() => handleDeleteKey(apiKey.id)}
                                className="px-3 py-2 text-sm text-delphi-error hover:bg-delphi-error/10 rounded transition-colors"
                              >
                                Remove
                              </button>
                            )}
                          </>
                        )}
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>

            {/* Ollama Configuration */}
            <div className="card">
              <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Local LLM Configuration</h3>
              <p className="text-sm text-delphi-text-muted mb-6">
                Configure Ollama or other local LLM endpoints for on-premise inference.
              </p>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm text-delphi-text-secondary mb-2">Ollama Endpoint</label>
                  <input
                    type="text"
                    defaultValue="http://localhost:11434"
                    className="input-primary w-full"
                  />
                </div>
                <div>
                  <label className="block text-sm text-delphi-text-secondary mb-2">Default Model</label>
                  <input
                    type="text"
                    defaultValue="llama2"
                    className="input-primary w-full"
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Notifications Tab */}
        {activeTab === 'notifications' && (
          <div className="space-y-6">
            <div className="card">
              <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Notification Preferences</h3>
              <p className="text-sm text-delphi-text-muted mb-6">
                Choose how you want to be notified about oracle activity.
              </p>

              <div className="space-y-4">
                {/* Header */}
                <div className="flex items-center gap-4 pb-2 border-b border-delphi-border">
                  <div className="flex-1" />
                  <div className="w-20 text-center text-xs text-delphi-text-muted">Email</div>
                  <div className="w-20 text-center text-xs text-delphi-text-muted">Slack</div>
                  <div className="w-20 text-center text-xs text-delphi-text-muted">Discord</div>
                </div>

                {notifications.map((notif) => (
                  <div key={notif.id} className="flex items-center gap-4">
                    <div className="flex-1">
                      <p className="font-medium text-delphi-text-primary">{notif.label}</p>
                      <p className="text-xs text-delphi-text-muted">{notif.description}</p>
                    </div>
                    {(['email', 'slack', 'discord'] as const).map((channel) => (
                      <div key={channel} className="w-20 flex justify-center">
                        <button
                          onClick={() => toggleNotification(notif.id, channel)}
                          className={`w-10 h-6 rounded-full transition-colors relative ${
                            notif[channel] ? 'bg-delphi-accent' : 'bg-delphi-bg-primary border border-delphi-border'
                          }`}
                        >
                          <div
                            className={`absolute top-1 w-4 h-4 rounded-full bg-white transition-transform ${
                              notif[channel] ? 'left-5' : 'left-1'
                            }`}
                          />
                        </button>
                      </div>
                    ))}
                  </div>
                ))}
              </div>
            </div>

            {/* Webhook Configuration */}
            <div className="card">
              <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Webhook Configuration</h3>
              <div className="grid gap-4">
                <div>
                  <label className="block text-sm text-delphi-text-secondary mb-2">Slack Webhook URL</label>
                  <input
                    type="text"
                    placeholder="https://hooks.slack.com/services/..."
                    className="input-primary w-full"
                  />
                </div>
                <div>
                  <label className="block text-sm text-delphi-text-secondary mb-2">Discord Webhook URL</label>
                  <input
                    type="text"
                    placeholder="https://discord.com/api/webhooks/..."
                    className="input-primary w-full"
                  />
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Security Tab */}
        {activeTab === 'security' && (
          <div className="space-y-6">
            <div className="card">
              <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Two-Factor Authentication</h3>
              <p className="text-sm text-delphi-text-muted mb-4">
                Add an extra layer of security to your account.
              </p>
              <button className="btn-primary">Enable 2FA</button>
            </div>

            <div className="card">
              <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Active Sessions</h3>
              <div className="space-y-3">
                {[
                  { device: 'Chrome on macOS', location: 'San Francisco, CA', current: true, lastActive: 'Now' },
                  { device: 'Safari on iPhone', location: 'San Francisco, CA', current: false, lastActive: '2 hours ago' },
                ].map((session, i) => (
                  <div key={i} className="flex items-center justify-between p-3 bg-delphi-bg-primary rounded-lg border border-delphi-border">
                    <div>
                      <p className="font-medium text-delphi-text-primary">
                        {session.device}
                        {session.current && <span className="ml-2 text-xs text-delphi-success">(current)</span>}
                      </p>
                      <p className="text-xs text-delphi-text-muted">{session.location} • {session.lastActive}</p>
                    </div>
                    {!session.current && (
                      <button className="text-sm text-delphi-error hover:underline">Revoke</button>
                    )}
                  </div>
                ))}
              </div>
            </div>

            <div className="card">
              <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Audit Log</h3>
              <p className="text-sm text-delphi-text-muted mb-4">
                Recent security-related events for your account.
              </p>
              <div className="space-y-2 text-sm font-mono">
                {[
                  { action: 'API key updated', provider: 'OpenAI', time: '2 hours ago' },
                  { action: 'Login from new device', provider: 'Safari on iPhone', time: '1 day ago' },
                  { action: 'Password changed', provider: null, time: '3 days ago' },
                ].map((event, i) => (
                  <div key={i} className="flex items-center gap-4 p-2 bg-delphi-bg-primary rounded">
                    <span className="text-delphi-text-muted">{event.time}</span>
                    <span className="text-delphi-text-primary">{event.action}</span>
                    {event.provider && <span className="text-delphi-text-secondary">({event.provider})</span>}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Billing Tab */}
        {activeTab === 'billing' && (
          <div className="space-y-6">
            <div className="card">
              <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Current Plan</h3>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-2xl font-bold text-delphi-accent">Pro</p>
                  <p className="text-sm text-delphi-text-muted">$49/month • Renews on Feb 1, 2026</p>
                </div>
                <button className="px-4 py-2 text-sm text-delphi-accent border border-delphi-accent rounded hover:bg-delphi-accent hover:text-white transition-colors">
                  Manage Subscription
                </button>
              </div>
            </div>

            <div className="card">
              <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Usage This Month</h3>
              <div className="grid grid-cols-3 gap-6">
                <div>
                  <p className="text-sm text-delphi-text-muted">Executions</p>
                  <p className="text-2xl font-mono text-delphi-text-primary">847 <span className="text-sm text-delphi-text-muted">/ 1000</span></p>
                  <div className="mt-2 h-2 bg-delphi-bg-primary rounded-full overflow-hidden">
                    <div className="h-full bg-delphi-accent" style={{ width: '84.7%' }} />
                  </div>
                </div>
                <div>
                  <p className="text-sm text-delphi-text-muted">Tokens</p>
                  <p className="text-2xl font-mono text-delphi-text-primary">654K <span className="text-sm text-delphi-text-muted">/ 1M</span></p>
                  <div className="mt-2 h-2 bg-delphi-bg-primary rounded-full overflow-hidden">
                    <div className="h-full bg-delphi-accent" style={{ width: '65.4%' }} />
                  </div>
                </div>
                <div>
                  <p className="text-sm text-delphi-text-muted">Storage</p>
                  <p className="text-2xl font-mono text-delphi-text-primary">23 GB <span className="text-sm text-delphi-text-muted">/ 50 GB</span></p>
                  <div className="mt-2 h-2 bg-delphi-bg-primary rounded-full overflow-hidden">
                    <div className="h-full bg-delphi-accent" style={{ width: '46%' }} />
                  </div>
                </div>
              </div>
            </div>

            <div className="card">
              <h3 className="text-lg font-semibold text-delphi-text-primary mb-4">Invoices</h3>
              <div className="space-y-2">
                {[
                  { date: 'Jan 1, 2026', amount: 49.00, status: 'Paid' },
                  { date: 'Dec 1, 2025', amount: 49.00, status: 'Paid' },
                  { date: 'Nov 1, 2025', amount: 49.00, status: 'Paid' },
                ].map((invoice, i) => (
                  <div key={i} className="flex items-center justify-between p-3 bg-delphi-bg-primary rounded-lg border border-delphi-border">
                    <div className="flex items-center gap-4">
                      <span className="text-delphi-text-primary">{invoice.date}</span>
                      <span className="text-delphi-success text-xs">{invoice.status}</span>
                    </div>
                    <div className="flex items-center gap-4">
                      <span className="font-mono text-delphi-text-primary">${invoice.amount.toFixed(2)}</span>
                      <button className="text-sm text-delphi-accent hover:underline">Download</button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}
      </motion.div>
    </div>
  )
}
