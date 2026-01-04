/** @type {import('tailwindcss').Config} */
export default {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        delphi: {
          'bg-primary': '#0a0a0f',
          'bg-secondary': '#12121a',
          'bg-elevated': '#1a1a24',
          'bg-hover': '#22222e',
          'border': '#2a2a3a',
          'border-hover': '#3a3a4a',
          'text-primary': '#e8e8ec',
          'text-secondary': '#a8a8b0',
          'text-muted': '#606070',
          'accent': '#4a9eff',
          'accent-hover': '#6ab0ff',
          'success': '#00d68f',
          'warning': '#ffaa00',
          'error': '#ff4757',
        },
      },
      fontFamily: {
        sans: ['Inter', 'SF Pro Display', '-apple-system', 'BlinkMacSystemFont', 'sans-serif'],
        mono: ['JetBrains Mono', 'Fira Code', 'SF Mono', 'Consolas', 'monospace'],
      },
      fontSize: {
        '2xs': ['0.625rem', { lineHeight: '0.875rem' }],
      },
      animation: {
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'spin-slow': 'spin 3s linear infinite',
      },
      boxShadow: {
        'glow': '0 0 20px rgba(74, 158, 255, 0.3)',
        'glow-lg': '0 0 40px rgba(74, 158, 255, 0.4)',
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-conic': 'conic-gradient(from 180deg at 50% 50%, var(--tw-gradient-stops))',
      },
    },
  },
  plugins: [],
}
