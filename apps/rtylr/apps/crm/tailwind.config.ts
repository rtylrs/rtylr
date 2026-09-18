import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./index.html', './src/**/*.{ts,tsx}'],
  theme: {
    extend: {
      colors: {
        accent: '#3b82f6',
        'accent-muted': 'rgba(59,130,246,0.12)',
        'accent-border': 'rgba(59,130,246,0.3)',
      },
    },
  },
  plugins: [],
};

export default config;
