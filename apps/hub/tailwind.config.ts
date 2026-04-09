import type { Config } from 'tailwindcss';
import sharedConfig from '@voxire/tailwind-config';

const config: Config = {
  ...sharedConfig,
  content: [
    './index.html',
    './src/**/*.{ts,tsx}',
    '../../packages/ui/src/**/*.{ts,tsx}',
  ],
};

export default config;
