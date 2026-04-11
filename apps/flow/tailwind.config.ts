import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        accent: '#64C5B9',
        'flow-bg': 'var(--flow-bg)',
        'flow-surface': 'var(--flow-surface)',
        'flow-card': 'var(--flow-card)',
        'flow-border': 'var(--flow-border)',
        'flow-text': 'var(--flow-text)',
        'flow-text-muted': 'var(--flow-text-muted)',
        'flow-text-dim': 'var(--flow-text-dim)',
        'flow-hover': 'var(--flow-hover)',
        'flow-active': 'var(--flow-active)',
      },
    },
  },
  plugins: [],
};

export default config;
