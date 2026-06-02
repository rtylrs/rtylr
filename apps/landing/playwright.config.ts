import { defineConfig, devices } from '@playwright/test';

/**
 * E2E for the landing site. The site ships as a static export (`output:
 * 'export'`) served by Cloudflare Pages, so we test exactly that: Playwright
 * runs `next build` then serves the static `dist/landing` output on port 4200.
 * (`next dev` is unusable here - the out-of-tree `distDir` breaks dev-mode
 * module resolution, and dev wouldn't exercise the exported artifact anyway.)
 * Specs hit the local server only; the fixture blocks any call to a local API.
 */
export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: process.env.CI ? 'github' : 'html',
  timeout: 60_000,

  use: {
    // 127.0.0.1, not localhost: `serve` binds IPv4 only (below), and on
    // Linux CI `localhost` can resolve to IPv6 first -> connection refused.
    baseURL: 'http://127.0.0.1:4200',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],

  webServer: {
    command: 'pnpm build && pnpm exec serve dist/landing -l tcp://127.0.0.1:4200 --no-port-switching',
    url: 'http://127.0.0.1:4200',
    reuseExistingServer: !process.env.CI,
    timeout: 240_000,
    stdout: 'pipe',
    stderr: 'pipe',
  },
});
