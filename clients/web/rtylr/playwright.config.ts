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
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  // Real dev servers are flakier than jsdom - two retries covers the
  // common "port warmed up mid-test" race without masking real bugs.
  retries: process.env.CI ? 2 : 0,
  // All specs share the same dev servers, so a single worker is required for
  // deterministic release-gate runs. Parallel workers tear down a server
  // while another spec is still using its port.
  workers: 1,
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
      use: {
        ...devices['Desktop Chrome'],
        // Opt-in escape hatch for a browser cache that does not match the
        // Playwright version this repo pins. Unset, this is a no-op and
        // Playwright resolves its own download as usual.
        //
        // It exists because `npx playwright install chromium` can appear to
        // succeed while delivering nothing: it left a 448 kB stub directory for
        // the expected build and never exited, so every launch failed with
        // "Executable doesn't exist" and no obvious cause. Pointing at a working
        // build that IS on disk is the difference between a runnable suite and
        // an unrunnable one, and hard-coding a path would not survive CI.
        //
        //   PLAYWRIGHT_CHROMIUM_PATH="$HOME/Library/Caches/ms-playwright/\
        //   chromium-1223/chrome-mac-arm64/Google Chrome for Testing.app/\
        //   Contents/MacOS/Google Chrome for Testing" pnpm test:e2e
        ...(process.env.PLAYWRIGHT_CHROMIUM_PATH
          ? { launchOptions: { executablePath: process.env.PLAYWRIGHT_CHROMIUM_PATH } }
          : {}),
      },
    },
    {
      name: 'chromium-390',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 390, height: 844 },
        ...(process.env.PLAYWRIGHT_CHROMIUM_PATH
          ? { launchOptions: { executablePath: process.env.PLAYWRIGHT_CHROMIUM_PATH } }
          : {}),
      },
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
      env: { VITE_API_URL: 'http://localhost:9999/' },
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/finance dev',
      url: 'http://localhost:5182',
      env: { VITE_API_URL: 'https://api.voxire.com/' },
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/crm dev',
      url: 'http://localhost:5181',
      env: { VITE_API_URL: 'http://localhost:9999/' },
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/auth dev',
      url: 'http://localhost:5176',
      env: { VITE_API_URL: 'http://localhost:9999/' },
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/portal dev',
      url: 'http://localhost:5175',
      env: { VITE_API_URL: 'http://localhost:9999/' },
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
      env: { VITE_API_URL: 'http://localhost:9999/' },
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
    {
      command: 'pnpm --filter @rtylr/flow dev',
      url: 'http://localhost:5183',
      // Without this the app falls back to the production API, so every
      // `page.route(/localhost:9999/)` stub in the flow specs matched nothing
      // and the specs asserted against whatever prod returned to an
      // unauthenticated request. Pointing the dev server at the stub host is
      // what makes those stubs mean anything.
      env: { VITE_API_URL: 'http://localhost:9999/' },
      reuseExistingServer: !process.env.CI,
      timeout: 120_000,
      stdout: 'pipe',
      stderr: 'pipe',
    },
  ],
});
