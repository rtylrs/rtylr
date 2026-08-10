import { defineConfig, devices } from '@playwright/test';

/**
 * End-to-end tests run against real dev servers booted by Playwright.
 * Every spec uses `page.route()` to stub API calls - we never hit the
 * real `api.voxire.com` in tests, so CI is deterministic and we don't
 * pollute prod data.
 *
 * The `webServer` array below boots each app's Vite dev server before
 * the tests run; Playwright waits for each port to accept connections.
 * `reuseExistingServer` lets local developers keep their own dev
 * server running while running e2e tests.
 *
 * To keep CI runtime manageable, we only boot the apps we actually
 * have specs for. Add more entries here when specs for other apps land.
 */
export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  // Real dev servers are flakier than jsdom - two retries covers the
  // common "port warmed up mid-test" race without masking real bugs.
  retries: process.env.CI ? 2 : 0,
  // Single worker in CI because all specs share the same dev servers
  // and port contention is worse than sequential runtime.
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? 'github' : 'html',
  // 60s ceiling is generous - a slow dev server boot can eat 20s by
  // itself, and a spec that visits 3 pages with hydration each needs
  // the headroom.
  timeout: 60_000,

  use: {
    // Default baseURL - individual specs override via `page.goto()`.
    baseURL: 'http://localhost:4200',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    // Capture videos only when a test fails and is retried - keeps
    // successful-run artifacts small.
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  webServer: [
    {
      command: 'pnpm --filter @rtylr/recruit dev',
      url: 'http://localhost:5184',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/erp dev',
      url: 'http://localhost:5179',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/finance dev',
      url: 'http://localhost:5182',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/crm dev',
      url: 'http://localhost:5181',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/auth dev',
      url: 'http://localhost:5176',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter dash dev',
      url: 'http://localhost:5175',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/menu dev',
      url: 'http://localhost:5174',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/hr dev',
      url: 'http://localhost:5180',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/flow dev',
      url: 'http://localhost:5183',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/insights dev',
      url: 'http://localhost:5187',
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
  ],
});
