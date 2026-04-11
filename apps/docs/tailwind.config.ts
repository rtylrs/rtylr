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
        docs: {
          DEFAULT: '#4583C4',
          muted: 'rgba(69,131,196,0.08)',
          border: 'rgba(69,131,196,0.2)',
        },
      },
    },
  },
  content: [
    './index.html',
    './src/**/*.{ts,tsx}',
    '../../packages/ui/src/**/*.{ts,tsx}',
  ],
};

export default config;
