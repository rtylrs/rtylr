import type { Config } from 'tailwindcss';
import sharedConfig from '@voxire/tailwind-config';

const config: Config = {
  ...sharedConfig,
  theme: {
    ...sharedConfig.theme,
    extend: {
      ...sharedConfig.theme?.extend,
      colors: {
        ...sharedConfig.theme?.extend?.colors,
        primary: {
          DEFAULT: 'var(--hub-text)',
          foreground: 'var(--hub-text)',
        },
        admin: {
          DEFAULT: '#262626',
          bg: 'var(--hub-bg)',
          border: 'var(--hub-border)',
        },
        insights: {
          DEFAULT: '#6366f1',
          muted: 'rgba(99,102,241,0.12)',
          border: 'rgba(99,102,241,0.3)',
        },
      },
    },
  },
  content: ['./index.html', './src/**/*.{ts,tsx}', '../../../../packages/ui/src/**/*.{ts,tsx}'],
};

export default config;
