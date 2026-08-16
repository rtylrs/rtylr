import { defineConfig } from 'vitest/config';
import path from 'path';

export default defineConfig({
  resolve: {
    alias: {
      react: path.resolve(__dirname, '../../node_modules/react'),
      'react-dom': path.resolve(__dirname, '../../node_modules/react-dom'),
    },
  },
  test: {
    globals: true,
    environment: 'happy-dom',
    setupFiles: [path.resolve(__dirname, '../../test/setup.ts')],
    include: ['src/**/*.test.{ts,tsx}'],
    server: {
      deps: {
        inline: [/react/, /@tanstack\/react-query/],
      },
    },
  },
});
