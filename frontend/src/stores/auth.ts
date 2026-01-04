import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import { apiClient } from '../services/api'
import type { User } from '../types'

interface AuthState {
  user: User | null
  token: string | null
  isAuthenticated: boolean
  loading: boolean
  error: string | null
  
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, name: string) => Promise<void>
  logout: () => void
  checkAuth: () => void
  clearError: () => void
}

// Check if we're in demo mode
const isDemoMode = typeof window !== 'undefined' && 
  (import.meta.env.VITE_DEMO_MODE === 'true' || !import.meta.env.VITE_API_URL)

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      user: null,
      token: null,
      isAuthenticated: false,
      loading: false,
      error: null,

      login: async (email: string, password: string) => {
        set({ loading: true, error: null })
        try {
          const response = await apiClient.post('/auth/login', { email, password })
          const { token, user } = response.data

          localStorage.setItem('token', token)
          set({ user, token, isAuthenticated: true, loading: false })
        } catch (error: unknown) {
          const err = error as { response?: { data?: { message?: string } } }
          set({
            error: err.response?.data?.message || 'Login failed',
            loading: false,
          })
          throw error
        }
      },

      register: async (email: string, password: string, name: string) => {
        set({ loading: true, error: null })
        try {
          const response = await apiClient.post('/auth/register', { email, password, name })
          const { token, user } = response.data

          localStorage.setItem('token', token)
          set({ user, token, isAuthenticated: true, loading: false })
        } catch (error: unknown) {
          const err = error as { response?: { data?: { message?: string } } }
          set({
            error: err.response?.data?.message || 'Registration failed',
            loading: false,
          })
          throw error
        }
      },

      logout: () => {
        localStorage.removeItem('token')
        set({ user: null, token: null, isAuthenticated: false })
      },

      checkAuth: () => {
        const token = localStorage.getItem('token')
        if (token) {
          // In demo mode, just set authenticated
          if (isDemoMode) {
            set({ 
              isAuthenticated: true, 
              token,
              user: {
                id: 'demo-user-id',
                email: 'demo@delphi.dev',
                name: 'Demo User',
              }
            })
          } else {
            set({ isAuthenticated: true, token })
          }
        }
      },

      clearError: () => set({ error: null }),
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ token: state.token }),
    }
  )
)
