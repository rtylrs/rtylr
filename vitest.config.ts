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
    include: [
      'packages/*/src/**/*.test.{ts,tsx}',
      'apps/*/src/**/*.test.{ts,tsx}',
      // Root-level build tooling (the shared Vite plugins). Not covered by
      // `pnpm lint` or `pnpm typecheck`, which scope to apps/ and packages/,
      // so its tests are the only automated gate it has.
      './*.test.ts',
    ],
    exclude: ['**/node_modules/**', 'dist', '.next', 'e2e'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json-summary'],
      include: ['packages/*/src/**', 'apps/*/src/**'],
      exclude: ['**/*.d.ts', '**/*.test.*', '**/node_modules/**'],
    },
  },
});
