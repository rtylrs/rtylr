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
        // Override primary so text-primary / bg-primary use theme CSS vars
        primary: {
          DEFAULT: 'var(--hub-text)',
          foreground: 'var(--hub-text)',
        },
        // Override admin-bg so bg-admin-bg follows the theme
        admin: {
          DEFAULT: '#262626',
          bg: 'var(--hub-bg)',
          border: 'var(--hub-border)',
        },
      },
    },
  },
  content: ['./index.html', './src/**/*.{ts,tsx}', '../../../../packages/ui/src/**/*.{ts,tsx}'],
};

export default config;
