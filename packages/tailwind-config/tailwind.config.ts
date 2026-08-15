import type { Config } from 'tailwindcss';

const config: Omit<Config, 'content'> = {
  darkMode: ['class', '[data-theme="dark"]'],
  theme: {
    extend: {
      colors: {
        // Brand
        primary: {
          DEFAULT: '#040428',
          foreground: '#ffffff',
        },
        secondary: {
          DEFAULT: '#4485c4',
          foreground: '#ffffff',
        },
        accent: {
          DEFAULT: '#03aed2',
          // Near-black, not white: white on the accent fill measures 1.9:1 and
          // fails WCAG AA at every size. Consumers map this to
          // --color-accent-foreground in their own @theme block.
          foreground: '#0a0a0a',
        },
        // Admin / Hub
        admin: {
          DEFAULT: '#262626',
          bg: '#eff0f1',
          border: 'rgba(255,255,255,0.08)',
        },
        // POS
        pos: {
          bg: 'var(--pos-bg)',
          card: 'var(--pos-card)',
          border: 'var(--pos-border)',
          muted: 'var(--pos-muted)',
          text: 'var(--pos-text)',
          input: 'var(--pos-input)',
        },
        // Semantic
        destructive: {
          DEFAULT: '#dc2626',
          foreground: '#ffffff',
        },
        success: {
          DEFAULT: '#059669',
          foreground: '#ffffff',
        },
        warning: {
          DEFAULT: '#d97706',
          foreground: '#ffffff',
        },
        muted: {
          DEFAULT: '#6b7280',
          foreground: '#9ca3af',
        },
        border: 'rgba(0,0,0,0.1)',
        ring: '#03aed2',
      },
      fontFamily: {
        sans: ['"Plus Jakarta Sans"', 'Inter', 'system-ui', 'sans-serif'],
        mono: ['ui-monospace', 'monospace'],
      },
      borderRadius: {
        DEFAULT: '8px',
        sm: '6px',
        md: '8px',
        lg: '12px',
        xl: '16px',
        '2xl': '20px',
      },
      boxShadow: {
        card: '0 1px 4px rgba(0,0,0,0.08)',
        dropdown: '0 8px 24px rgba(0,0,0,0.14)',
        modal: '0 16px 48px rgba(0,0,0,0.18)',
        'pos-phone': '0 0 0 10px #1a1a1a, 0 24px 60px rgba(0,0,0,0.35)',
      },
      keyframes: {
        shake: {
          '0%, 100%': { transform: 'translateX(0)' },
          '20%': { transform: 'translateX(-8px)' },
          '40%': { transform: 'translateX(8px)' },
          '60%': { transform: 'translateX(-6px)' },
          '80%': { transform: 'translateX(6px)' },
        },
        'fade-in': {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
        'slide-up': {
          from: { opacity: '0', transform: 'translateY(8px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        'spin-slow': {
          to: { transform: 'rotate(360deg)' },
        },
        'dropdown-in': {
          from: { opacity: '0', transform: 'translateY(-4px) scale(0.97)' },
          to: { opacity: '1', transform: 'translateY(0) scale(1)' },
        },
      },
      animation: {
        shake: 'shake 0.4s ease',
        'fade-in': 'fade-in 0.2s ease',
        'slide-up': 'slide-up 0.2s ease',
        'spin-slow': 'spin-slow 0.8s linear infinite',
        'dropdown-in': 'dropdown-in 0.15s ease',
      },
    },
  },
  plugins: [],
};

export default config;
