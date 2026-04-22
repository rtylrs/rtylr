import { defineConfig } from 'vitest/config';

export default defineConfig({
  test: {
    globals: true,
    // Switched from jsdom 29 → happy-dom because jsdom 29 depends on
    // an undici export path (`wrap-handler.js`) that was removed in
    // recent undici majors. happy-dom is faster, lighter, and has no
    // undici dep, so it sidesteps the conflict.
    environment: 'happy-dom',
    setupFiles: ['./test/setup.ts'],
    include: ['packages/*/src/**/*.test.{ts,tsx}', 'apps/*/src/**/*.test.{ts,tsx}'],
    exclude: ['**/node_modules/**', 'dist', '.next', 'e2e'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json-summary'],
      include: ['packages/*/src/**', 'apps/*/src/**'],
      exclude: ['**/*.d.ts', '**/*.test.*', '**/node_modules/**'],
    },
  },
});
