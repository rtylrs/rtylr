import { defineConfig } from 'vitest/config';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

/**
 * Where a test run is allowed to send a request: nowhere. Port 9 is discard,
 * so a connection is refused immediately and nothing stays in flight.
 */
const TEST_API_BASE = 'http://127.0.0.1:9';

export default defineConfig({
  // `test.env` reaches process.env, but @voxire/config reads
  // `import.meta.env.VITE_API_URL` and is resolved as an external dependency,
  // so the value has to be substituted at transform time to reach it.
  define: {
    'import.meta.env.VITE_API_URL': JSON.stringify(TEST_API_BASE),
  },
  resolve: {
    alias: {
      react: path.resolve(__dirname, 'node_modules/react'),
      'react-dom': path.resolve(__dirname, 'node_modules/react-dom'),
    },
  },
  test: {
    globals: true,
    environment: 'happy-dom',
    // Keep tests away from production: a port-9 API base refuses immediately,
    // so no request can remain in flight when happy-dom tears down a window.
    env: {
      VITE_API_URL: TEST_API_BASE,
    },
    environmentOptions: {
      happyDOM: {
        settings: {
          disableCSSFileLoading: true,
          disableJavaScriptFileLoading: true,
          // Report the skipped load as a success, so a stylesheet link does not
          // surface as an error the page never sees in a browser either.
          handleDisabledFileLoadingAsSuccess: true,
        },
      },
    },
    setupFiles: [path.resolve(__dirname, 'test/setup.ts')],
    exclude: ['**/node_modules/**', 'dist', '.next', 'e2e'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json-summary'],
      include: ['../../packages/*/src/**', 'apps/*/src/**'],
      exclude: ['**/*.d.ts', '**/*.test.*', '**/node_modules/**'],
    },
    server: {
      deps: {
        inline: [/react/, /@tanstack\/react-query/],
      },
    },
    // The shared packages live at the repository root, outside this workspace
    // (apps/rtylr). Vitest only loads test files that sit under the project
    // they belong to, so the shared packages get their own project rooted at
    // the repository. Inline projects inherit the options above.
    projects: [
      {
        extends: true,
        test: {
          name: 'packages',
          root: path.resolve(__dirname, '../..'),
          include: ['packages/*/src/**/*.test.{ts,tsx}'],
        },
      },
      {
        extends: true,
        test: {
          name: 'apps',
          root: __dirname,
          include: ['apps/*/src/**/*.test.{ts,tsx}', './*.test.ts'],
        },
      },
    ],
  },
});
