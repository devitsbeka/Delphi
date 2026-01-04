import { useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { motion } from 'framer-motion'
import { useAuthStore } from '../stores/auth'

export function Register() {
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const { register, loading, error, clearError } = useAuthStore()
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    
    if (password !== confirmPassword) {
      return
    }

    try {
      await register(email, password, name)
      navigate('/')
    } catch {
      // Error is handled in store
    }
  }

  const passwordsMatch = password === confirmPassword || confirmPassword === ''

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
          <p className="text-delphi-text-muted mt-2">Create your command center</p>
        </div>

        {/* Register Form */}
        <div className="card">
          <h2 className="text-xl font-semibold text-delphi-text-primary mb-6">Create account</h2>

          {error && (
            <div className="mb-4 p-3 bg-delphi-error/10 border border-delphi-error/30 rounded-lg text-delphi-error text-sm">
              {error}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="label">Name</label>
              <input
                type="text"
                value={name}
                onChange={(e) => {
                  setName(e.target.value)
                  clearError()
                }}
                placeholder="John Doe"
                className="input-primary w-full"
                required
              />
            </div>

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
                minLength={8}
                required
              />
            </div>

            <div>
              <label className="label">Confirm Password</label>
              <input
                type="password"
                value={confirmPassword}
                onChange={(e) => {
                  setConfirmPassword(e.target.value)
                  clearError()
                }}
                placeholder="••••••••"
                className={`input-primary w-full ${!passwordsMatch ? 'border-delphi-error focus:border-delphi-error' : ''}`}
                required
              />
              {!passwordsMatch && (
                <p className="text-xs text-delphi-error mt-1">Passwords don't match</p>
              )}
            </div>

            <button
              type="submit"
              disabled={loading || !passwordsMatch}
              className="btn-primary w-full disabled:opacity-50"
            >
              {loading ? 'Creating account...' : 'Create Account'}
            </button>
          </form>

          <p className="text-center text-sm text-delphi-text-muted mt-6">
            Already have an account?{' '}
            <Link to="/login" className="text-delphi-accent hover:underline">
              Sign in
            </Link>
          </p>
        </div>

        {/* Footer */}
        <p className="text-center text-xs text-delphi-text-muted mt-8">
          By creating an account, you agree to our Terms of Service and Privacy Policy
        </p>
      </motion.div>
    </div>
  )
}
