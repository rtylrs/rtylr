import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        accent: {
          DEFAULT: '#8b5cf6',
          muted: 'rgba(139,92,246,0.12)',
          border: 'rgba(139,92,246,0.3)',
        },
        'hr-bg': 'var(--hr-bg)',
        'hr-surface': 'var(--hr-surface)',
        'hr-card': 'var(--hr-card)',
        'hr-border': 'var(--hr-border)',
        'hr-text': 'var(--hr-text)',
        'hr-muted': 'var(--hr-text-muted)',
      },
      fontFamily: {
        sans: ['Plus Jakarta Sans', 'Inter', 'system-ui', 'sans-serif'],
      },
    },
  },
  plugins: [],
};

export default config;
