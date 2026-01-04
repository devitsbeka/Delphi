import { Link, useLocation, Outlet, useNavigate } from 'react-router-dom'
import { motion, AnimatePresence } from 'framer-motion'
import { useState } from 'react'
import useAuthStore from '../stores/auth'

const navigation = [
  { name: 'Dashboard', path: '/', icon: '◆' },
  { name: 'Oracles', path: '/agents', icon: '◎' },
  { name: 'Execute', path: '/execute', icon: '▶' },
  { name: 'Knowledge', path: '/knowledge', icon: '◈' },
  { name: 'Repositories', path: '/repositories', icon: '⬡' },
  { name: 'Businesses', path: '/businesses', icon: '◇' },
  { name: 'Costs', path: '/costs', icon: '◐' },
  { name: 'Settings', path: '/settings', icon: '⚙' },
]

export default function Layout() {
  const location = useLocation()
  const navigate = useNavigate()
  const { user, logout } = useAuthStore()
  const [sidebarCollapsed, setSidebarCollapsed] = useState(false)
  const [showUserMenu, setShowUserMenu] = useState(false)

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  return (
    <div className="flex h-screen bg-delphi-bg-primary">
      {/* Sidebar */}
      <motion.aside
        initial={false}
        animate={{ width: sidebarCollapsed ? 64 : 240 }}
        className="bg-delphi-bg-secondary border-r border-delphi-border flex flex-col"
      >
        {/* Logo */}
        <div className="h-16 flex items-center px-4 border-b border-delphi-border">
          <Link to="/" className="flex items-center gap-3">
            <span className="text-2xl">🔮</span>
            {!sidebarCollapsed && (
              <motion.span
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="text-xl font-bold text-gradient-accent"
              >
                Delphi
              </motion.span>
            )}
          </Link>
        </div>

        {/* Navigation */}
        <nav className="flex-1 py-4 overflow-y-auto">
          <ul className="space-y-1 px-2">
            {navigation.map((item) => {
              const isActive = location.pathname === item.path ||
                (item.path !== '/' && location.pathname.startsWith(item.path))

              return (
                <li key={item.path}>
                  <Link
                    to={item.path}
                    className={`flex items-center gap-3 px-3 py-2.5 rounded-lg transition-all duration-200 group
                      ${isActive
                        ? 'bg-delphi-accent/10 text-delphi-accent border-l-2 border-delphi-accent'
                        : 'text-delphi-text-muted hover:bg-delphi-bg-hover hover:text-delphi-text-primary'
                      }`}
                  >
                    <span className={`text-lg ${isActive ? 'text-delphi-accent' : 'text-delphi-text-muted group-hover:text-delphi-text-primary'}`}>
                      {item.icon}
                    </span>
                    {!sidebarCollapsed && (
                      <motion.span
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        className="text-sm font-medium"
                      >
                        {item.name}
                      </motion.span>
                    )}
                  </Link>
                </li>
              )
            })}
          </ul>
        </nav>

        {/* Collapse Toggle */}
        <div className="p-2 border-t border-delphi-border">
          <button
            onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
            className="w-full flex items-center justify-center py-2 rounded-lg text-delphi-text-muted hover:bg-delphi-bg-hover hover:text-delphi-text-primary transition-colors"
          >
            <motion.span
              animate={{ rotate: sidebarCollapsed ? 180 : 0 }}
              className="text-sm"
            >
              ◀
            </motion.span>
          </button>
        </div>
      </motion.aside>

      {/* Main Content */}
      <div className="flex-1 flex flex-col overflow-hidden">
        {/* Header */}
        <header className="h-16 bg-delphi-bg-secondary border-b border-delphi-border flex items-center justify-between px-6">
          {/* Search */}
          <div className="flex items-center gap-4">
            <div className="relative">
              <input
                type="text"
                placeholder="Search... (⌘K)"
                className="w-64 px-4 py-2 pl-10 text-sm bg-delphi-bg-primary border border-delphi-border rounded-lg
                         text-delphi-text-primary placeholder:text-delphi-text-muted
                         focus:outline-none focus:border-delphi-accent focus:ring-1 focus:ring-delphi-accent/30"
              />
              <span className="absolute left-3 top-1/2 -translate-y-1/2 text-delphi-text-muted">
                🔍
              </span>
            </div>
          </div>

          {/* Right side */}
          <div className="flex items-center gap-4">
            {/* Notifications */}
            <button className="relative p-2 text-delphi-text-muted hover:text-delphi-text-primary transition-colors">
              <span className="text-lg">🔔</span>
              <span className="absolute top-1 right-1 w-2 h-2 bg-delphi-accent rounded-full" />
            </button>

            {/* User Menu */}
            <div className="relative">
              <button
                onClick={() => setShowUserMenu(!showUserMenu)}
                className="flex items-center gap-2 p-2 rounded-lg hover:bg-delphi-bg-hover transition-colors"
              >
                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-delphi-accent to-purple-500 flex items-center justify-center text-white text-sm font-medium">
                  {user?.name?.[0]?.toUpperCase() || 'U'}
                </div>
                <span className="text-sm text-delphi-text-primary hidden md:block">
                  {user?.name || user?.email || 'User'}
                </span>
                <span className="text-xs text-delphi-text-muted">▼</span>
              </button>

              <AnimatePresence>
                {showUserMenu && (
                  <motion.div
                    initial={{ opacity: 0, y: -10 }}
                    animate={{ opacity: 1, y: 0 }}
                    exit={{ opacity: 0, y: -10 }}
                    className="absolute right-0 mt-2 w-48 bg-delphi-bg-elevated border border-delphi-border rounded-lg shadow-lg overflow-hidden z-50"
                  >
                    <div className="p-3 border-b border-delphi-border">
                      <p className="text-sm font-medium text-delphi-text-primary">{user?.name || 'User'}</p>
                      <p className="text-xs text-delphi-text-muted">{user?.email}</p>
                    </div>
                    <ul>
                      <li>
                        <Link
                          to="/settings"
                          className="block px-4 py-2 text-sm text-delphi-text-secondary hover:bg-delphi-bg-hover transition-colors"
                          onClick={() => setShowUserMenu(false)}
                        >
                          Settings
                        </Link>
                      </li>
                      <li>
                        <button
                          onClick={handleLogout}
                          className="w-full text-left px-4 py-2 text-sm text-delphi-error hover:bg-delphi-bg-hover transition-colors"
                        >
                          Sign Out
                        </button>
                      </li>
                    </ul>
                  </motion.div>
                )}
              </AnimatePresence>
            </div>
          </div>
        </header>

        {/* Page Content */}
        <main className="flex-1 overflow-y-auto bg-delphi-bg-primary">
          <AnimatePresence mode="wait">
            <motion.div
              key={location.pathname}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              transition={{ duration: 0.2 }}
              className="h-full"
            >
              <Outlet />
            </motion.div>
          </AnimatePresence>
        </main>
      </div>
    </div>
  )
}
