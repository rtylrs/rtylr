import { flowAuthFields } from './support/flow-auth';
import fs from 'node:fs';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
import { test, expect, type Route } from './fixtures';

/**
 * Flow Auth and Guard Specs
 * Verifies ErrorBoundary -> AppFlagGuard -> AppGuard hierarchy, auth token checks,
 * and preview route flag enforcement.
 */

const FLOW_URL = process.env.FLOW_URL ?? 'http://localhost:5197/flow';
const API_HOST = 'http://localhost:9999';

async function primeAuthCookie(context: import('@playwright/test').BrowserContext): Promise<void> {
  const fakeJwt = 'e2e-test-token';
  await context.addCookies([
    {
      name: 'token',
      value: fakeJwt,
      domain: 'localhost',
      path: '/',
      httpOnly: false,
      sameSite: 'Lax',
      secure: false,
    },
  ]);
}

test.describe('Flow - Auth & Guard Verification', () => {
  // Regression test for FLOW-BUG-001: onion's committed database default must
  // not point at a remote host.
  //
  // This used to read `process.env.UNOVERRIDDEN_DB_HOST ?? '127.0.0.1'` and
  // assert that same '127.0.0.1', so it supplied its own answer and could not
  // fail whatever onion actually shipped (#1026). It now reads onion's source.
  //
  // What a test living in rtylr can police is onion's *committed* default. A
  // developer's own `.env` is gitignored and is their business, so it is
  // deliberately not asserted on here; onion has a runtime guard,
  // checkProductionSafety, that refuses a production host unless the run
  // declares that intent, which is the right place for that check.
  test('FLOW-BUG-001: onion ships a local database default, not a remote one', async () => {
    const onionDir =
      process.env.ONION_DIR ??
      path.resolve(path.dirname(fileURLToPath(import.meta.url)), '../../../services/onion');
    const envExample = path.join(onionDir, '.env.example');
    const mysqlSource = path.join(onionDir, 'internal/platform/database/dbclient/mysql.go');

    // Skip rather than pass when there is nothing to inspect, so a green run on
    // a machine without the onion checkout is not mistaken for evidence.
    test.skip(
      !fs.existsSync(envExample) || !fs.existsSync(mysqlSource),
      `onion checkout not found at ${onionDir}; set ONION_DIR to inspect it`
    );

    const isLocalHost = (value: string): boolean =>
      /^(localhost|127\.0\.0\.1|0\.0\.0\.0|::1|host\.docker\.internal|db|mysql)$/.test(
        value.trim().replace(/^["']|["']$/g, '')
      );

    const exampleHost = fs.readFileSync(envExample, 'utf8').match(/^\s*DB_HOST\s*=\s*(.*)$/m)?.[1];
    expect(exampleHost, '.env.example must declare DB_HOST').toBeDefined();
    expect(isLocalHost(exampleHost as string), `.env.example DB_HOST is not local`).toBe(true);

    // The compiled-in fallback used when DB_HOST is unset entirely.
    const codeDefault = fs
      .readFileSync(mysqlSource, 'utf8')
      .match(/DefaultDBHost\s*=\s*"([^"]*)"/)?.[1];
    expect(codeDefault, 'DefaultDBHost must exist in mysql.go').toBeDefined();
    expect(isLocalHost(codeDefault as string), 'DefaultDBHost is not local').toBe(true);
  });

  // Regression test for the Flow catch-all: TanStack Router v1 uses '$' as its
  // splat token, and the route was declared with '*', so it could never match
  // and an unknown URL stayed put on a bare "Not Found" (#1038).
  test('an unknown Flow URL shows a recovery link', async ({ page, context }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: '1001',
          email: 'owner@test.local',
          full_name: 'Test Owner',
          business_id: '1',
          ...flowAuthFields('owner'),
        }),
      })
    );

    await page.goto(`${FLOW_URL}/this-route-does-not-exist`);
    await page.waitForLoadState('networkidle');

    await expect(page.getByRole('heading', { name: "That page isn't here" })).toBeVisible();
    await expect(page.getByRole('link', { name: 'Go to my tasks' })).toHaveAttribute(
      'href',
      '/flow/my-issues'
    );
  });

  test('business admins can edit tasks regardless of collaborator role', async ({
    page,
    context,
  }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: '1001',
          email: 'owner@test.local',
          full_name: 'Test Owner',
          business_id: '1',
          ...flowAuthFields('admin'),
        }),
      })
    );

    await page.route(
      /localhost:9999\/flow\/project\/v2\/projects\/members(\?.*)?$/,
      (route: Route) =>
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            members: [{ id: '1', project_id: '1', user_id: '1001', role: 'viewer' }],
          }),
        })
    );

    await page.route(/localhost:9999\/flow\/task\/v2\/tasks\/101$/, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          task: {
            id: '101',
            project_id: '1',
            title: 'Implement SSO Auth Flow',
            description: '',
            status: 'todo',
            priority: 'no_priority',
            assignee_id: '1001',
            assignee_name: 'Test Owner',
            assignee_initials: 'TO',
            archived: false,
          },
          subtasks: [],
          comments: [],
          activity: [],
        }),
      })
    );

    await page.goto(`${FLOW_URL}/task/101`);
    await page.waitForLoadState('networkidle');

    const titleInput = page.locator('input[value*="SSO"]').first();
    await expect(titleInput).toBeEnabled();
  });

  test('AppGuard redirects unauthenticated requests when token cookie is missing', async ({
    page,
  }) => {
    await page.goto(`${FLOW_URL}/my-issues`);
    await page.waitForLoadState('networkidle');

    const url = page.url();
    expect(url).toMatch(/5176|\/auth|login/);
  });

  test('AppFlagGuard redirects preview routes when preview flag is off', async ({
    page,
    context,
  }) => {
    await primeAuthCookie(context);

    await page.route(`${API_HOST}/hub/user/v1/users/me`, (route: Route) =>
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          id: '1001',
          email: 'owner@test.local',
          full_name: 'Test Owner',
          business_id: '1',
          ...flowAuthFields('owner'),
        }),
      })
    );

    await page.goto(`${FLOW_URL}/timeline`);
    await page.waitForLoadState('networkidle');

    // AppFlagGuard renders its fallback in place rather than navigating, so the
    // URL is not the evidence. This used to pass on a workspace redirect that
    // happened to fire alongside it, and that redirect is gone.
    await expect(page.getByText('Preview route unavailable', { exact: true })).toBeVisible({
      timeout: 10_000,
    });
    expect(page.url()).toContain('/timeline');
  });
});
