import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? 'github' : 'html',
  timeout: 30_000,

  use: {
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // Dev servers to start before E2E tests
  // Uncomment when ready to run against local dev servers
  // webServer: [
  //   { command: 'pnpm --filter @rtylr/auth dev', port: 5176, reuseExistingServer: true },
  //   { command: 'pnpm --filter dash dev', port: 5175, reuseExistingServer: true },
  // ],
});
