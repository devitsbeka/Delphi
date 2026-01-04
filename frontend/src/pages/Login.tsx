import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useAuthStore } from '../stores/auth'

export function Login() {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const { login, loading, error, clearError } = useAuthStore()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    try {
      await login(email, password)
      navigate('/')
    } catch {
      // Error is handled in store
    }
  }

  // Demo mode quick login
  const handleDemoLogin = async () => {
    try {
      await login('demo@delphi.dev', 'demo')
      navigate('/')
    } catch {
      // Error is handled in store
    }
  }

  return (
    <div className="min-h-screen bg-delphi-bg-primary flex items-center justify-center p-4">
      {/* Background effects */}
      <div className="absolute inset-0 bg-grid opacity-20" />
      <div className="absolute inset-0 bg-radial-overlay" />

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        className="relative w-full max-w-md"
      >
        {/* Logo */}
        <div className="text-center mb-8">
          <span className="text-5xl">🔮</span>
          <h1 className="text-3xl font-bold text-gradient-accent mt-4">Delphi</h1>
          <p className="text-delphi-text-muted mt-2">AI Agent Command Center</p>
        </div>

        {/* Login Form */}
        <div className="card">
          <h2 className="text-xl font-semibold text-delphi-text-primary mb-6">Welcome back</h2>

          {error && (
            <div className="mb-4 p-3 bg-delphi-error/10 border border-delphi-error/30 rounded-lg text-delphi-error text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="label">Email</label>
              <input
                type="email"
                value={email}
                onChange={(e) => {
                  setEmail(e.target.value)
                  clearError()
                }}
                placeholder="you@example.com"
                className="input-primary w-full"
                required
              />
            </div>

            <div>
              <label className="label">Password</label>
              <input
                type="password"
                value={password}
                onChange={(e) => {
                  setPassword(e.target.value)
                  clearError()
                }}
                placeholder="••••••••"
                className="input-primary w-full"
                required
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              className="btn-primary w-full disabled:opacity-50"
            >
              {loading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>

          {/* Demo Mode Button */}
          <div className="mt-4 pt-4 border-t border-delphi-border">
            <button
              onClick={handleDemoLogin}
              disabled={loading}
              className="w-full py-2.5 text-sm text-delphi-accent border border-delphi-accent/50 rounded-lg
                       hover:bg-delphi-accent/10 transition-colors disabled:opacity-50"
            >
              🎮 Try Demo Mode
            </button>
            <p className="text-xs text-delphi-text-muted text-center mt-2">
              Explore the UI with sample data
            </p>
          </div>

          <p className="text-center text-sm text-delphi-text-muted mt-6">
            Don't have an account?{' '}
            <Link to="/register" className="text-delphi-accent hover:underline">
              Sign up
            </Link>
          </p>
        </div>

        {/* Footer */}
        <p className="text-center text-xs text-delphi-text-muted mt-8">
          By signing in, you agree to our Terms of Service and Privacy Policy
        </p>
      </motion.div>
    </div>
  )
}
