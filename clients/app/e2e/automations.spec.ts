import { test, expect, type Page } from './fixtures';
import { PORTAL_URL, stubPortalSession } from './portal-session';
import { captureForPR } from './support/screenshot';

interface WireAutomation {
  id: number;
  name: string;
  trigger_type: string;
  trigger_config: Record<string, unknown> | string;
  actions: Array<Record<string, unknown>>;
  is_active: boolean;
  run_count: number;
  created_at: string;
}
const existing = (): WireAutomation => ({
  id: 5,
  name: 'Order follow-up',
  trigger_type: 'order_created',
  trigger_config: { retained: 'trigger-config' },
  actions: [
    {
      type: 'send_email',
      subject: 'Thanks',
      body: 'Thanks for your order.',
      retained: 'action-config',
    },
  ],
  is_active: true,
  run_count: 4,
  created_at: '2026-10-01T10:00:00Z',
});

async function stubAutomations(page: Page, items: WireAutomation[] = []) {
  const state = { items, writes: [] as Record<string, unknown>[], fail: false };
  await page.route(/localhost:9999\/crm\/automations(?:\/\d+)?(?:\?.*)?$/, async (route) => {
    const method = route.request().method();
    const id = Number(new URL(route.request().url()).pathname.split('/').at(-1));
    if (state.fail) return route.fulfill({ status: 500, json: { error: 'fixture failure' } });
    if (method === 'GET')
      return route.fulfill({
        json: id
          ? { automation: state.items.find((item) => item.id === id) }
          : { automations: state.items },
      });
    if (method === 'DELETE') {
      state.items = state.items.filter((item) => item.id !== id);
      return route.fulfill({ json: {} });
    }
    const body = route.request().postDataJSON();
    state.writes.push(body);
    const item: WireAutomation = { ...existing(), ...body, id: id || 8 };
    state.items = [...state.items.filter((entry) => entry.id !== item.id), item];
    return route.fulfill({ json: { automation: item } });
  });
  return state;
}

test('standalone workflow creation, editing, run controls and deletion preserve configuration', async ({
  page,
  context,
}) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['crm', 'flow'] });
  const state = await stubAutomations(page);
  await page.goto(`${PORTAL_URL}/automations`);
  await expect(page.getByRole('heading', { name: 'Automations', exact: true })).toBeVisible();
  await expect(page.locator('.portal-rail [role="separator"]')).toHaveCount(1);
  await expect(
    page.locator('.portal-rail').getByRole('link', { name: 'Automations' })
  ).toHaveAttribute('aria-current', 'page');
  await page.getByRole('button', { name: 'New automation', exact: true }).click();
  await expect(page).toHaveURL(/\/automations\/new$/);
  await page.getByLabel('Name', { exact: true }).fill('Order thanks');
  await expect(page.getByRole('button', { name: 'Save workflow' })).toBeDisabled();
  await page.getByRole('button', { name: 'Send email', exact: true }).click();
  await page.getByLabel('Email subject').fill('Thank you');
  await page.getByLabel('Email message').fill('Thanks for your order.');
  await captureForPR(page, `automations-builder-${test.info().project.name}`);
  await page.getByRole('button', { name: 'Save workflow' }).click();
  await expect(page).toHaveURL(/\/automations$/);
  await expect(page.getByRole('button', { name: /Order thanks/ })).toBeVisible();
  expect(state.writes[0]).toMatchObject({
    trigger_type: 'order_created',
    is_active: false,
    actions: [{ type: 'send_email', subject: 'Thank you', body: 'Thanks for your order.' }],
  });
  await page.getByRole('button', { name: 'Resume automation' }).click();
  await expect(page.getByRole('button', { name: 'Pause automation' })).toBeVisible();
  expect(state.items[0].actions[0].subject).toBe('Thank you');
  await page.getByRole('button', { name: /Order thanks/ }).click();
  await page.getByLabel('Name', { exact: true }).fill('Updated workflow');
  await expect(page.getByLabel('Email subject')).toHaveValue('Thank you');
  await page.getByRole('button', { name: 'Save workflow' }).click();
  await expect(page.getByRole('button', { name: /Updated workflow/ })).toBeVisible();
  expect(state.items[0].is_active).toBe(true);
  await captureForPR(page, `automations-list-${test.info().project.name}`);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await page.getByRole('button', { name: 'Pause automation' }).click();
  await expect(page.getByRole('button', { name: 'Resume automation' })).toBeVisible();
  await page.getByRole('button', { name: 'Delete automation', exact: true }).click();
  await page.getByRole('dialog').getByRole('button', { name: 'Delete', exact: true }).click();
  await expect(page.getByRole('heading', { name: 'No automations yet' })).toBeVisible();
});

test('legacy links redirect and keep trigger/action details through save and pause', async ({
  page,
  context,
}) => {
  await stubPortalSession(page, context, {
    appIds: ['crm'],
    permissions: ['crm:automations:manage'],
  });
  const entry = existing();
  entry.trigger_config = JSON.stringify(entry.trigger_config);
  const state = await stubAutomations(page, [entry]);
  await page.goto(`${PORTAL_URL}/crm/automations/new?id=5`);
  await expect(page).toHaveURL(/\/automations\/new\?id=5$/);
  await expect(page.getByLabel('Email subject')).toHaveValue('Thanks');
  await page.getByLabel('Name', { exact: true }).fill('Renamed');
  await page.getByRole('button', { name: 'Save workflow' }).click();
  await expect(page).toHaveURL(/\/automations$/);
  expect(state.writes[0]).toMatchObject({
    trigger_config: { retained: 'trigger-config' },
    actions: [{ retained: 'action-config', subject: 'Thanks' }],
  });
  await page.getByRole('button', { name: 'Pause automation' }).click();
  await expect(page.getByRole('button', { name: 'Resume automation' })).toBeVisible();
  expect(state.items[0].actions[0].retained).toBe('action-config');
  await page.goto(`${PORTAL_URL}/crm`);
  await expect(page).toHaveURL(/\/automations$/);
});

test('search, filters and retry distinguish errors from an empty list', async ({
  page,
  context,
}) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['crm'] });
  const state = await stubAutomations(page, [existing()]);
  state.fail = true;
  await page.goto(`${PORTAL_URL}/automations`);
  await expect(
    page.getByRole('alert').filter({ has: page.getByRole('button', { name: 'Refresh' }) })
  ).toBeVisible();
  await expect(page.getByRole('heading', { name: 'No automations yet' })).toHaveCount(0);
  state.fail = false;
  await page.getByRole('button', { name: 'Refresh', exact: true }).click();
  await expect(page.getByRole('button', { name: /Order follow-up/ })).toBeVisible();
  await page.getByLabel('Status', { exact: true }).selectOption('paused');
  await expect(page.getByText('No matching automations')).toBeVisible();
  await page.getByLabel('Status', { exact: true }).selectOption('all');
  await page.getByRole('searchbox').fill('missing');
  await expect(page.getByText('No matching automations')).toBeVisible();
});

test('a Flow-only member cannot access or request workflows', async ({ page, context }) => {
  await stubPortalSession(page, context, { appIds: ['flow'], permissions: ['flow:tasks:view'] });
  const state = await stubAutomations(page);
  await page.goto(`${PORTAL_URL}/automations/new`);
  await expect(page.locator('.portal-access-denied')).toBeVisible();
  await expect(page.locator('.portal-rail').getByRole('link', { name: 'Automations' })).toHaveCount(
    0
  );
  expect(state.writes).toHaveLength(0);
});

test('Arabic layout keeps the standalone tab and setup fields usable', async ({
  page,
  context,
}) => {
  await context.addCookies([{ name: 'language', value: 'ar', url: PORTAL_URL }]);
  await stubPortalSession(page, context, { isOwner: true, appIds: ['crm'] });
  await stubAutomations(page, [existing()]);
  await page.goto(`${PORTAL_URL}/automations/new?id=5`);
  await expect(page.getByLabel('موضوع البريد الإلكتروني')).toHaveValue('Thanks');
  await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
  await captureForPR(page, `automations-ar-${test.info().project.name}`);
});
