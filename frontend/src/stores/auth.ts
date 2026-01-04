import { create } from 'zustand'
import { authAPI } from '../services/api'
import { User } from '../types'
import { toast } from 'react-hot-toast'

interface AuthState {
  token: string | null
  user: User | null
  isAuthenticated: boolean
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (name: string, email: string, password: string) => Promise<void>
  logout: () => void
}

const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem('token'),
  user: JSON.parse(localStorage.getItem('user') || 'null'),
  isAuthenticated: !!localStorage.getItem('token'),
  loading: false,

  login: async (email, password) => {
    set({ loading: true })
    try {
      const response = await authAPI.login(email, password)
      const { token, user } = response.data
      localStorage.setItem('token', token)
      localStorage.setItem('user', JSON.stringify(user))
      set({ token, user, isAuthenticated: true, loading: false })
      toast.success('Welcome to Delphi!')
    } catch (error: any) {
      set({ loading: false })
      throw error
    }
  },

  register: async (name, email, password) => {
    set({ loading: true })
    try {
      const response = await authAPI.register(name, email, password)
      const { token, user } = response.data
      localStorage.setItem('token', token)
      localStorage.setItem('user', JSON.stringify(user))
      set({ token, user, isAuthenticated: true, loading: false })
      toast.success('Account created! Welcome to Delphi!')
    } catch (error: any) {
      set({ loading: false })
      throw error
    }
  },

  logout: () => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    set({ token: null, user: null, isAuthenticated: false })
    toast('Logged out.', { icon: '👋' })
  },
}))

export default useAuthStore
