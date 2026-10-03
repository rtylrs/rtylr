import { test, expect } from './fixtures';
import { PORTAL_URL, stubPortalSession } from './portal-session';

const API = 'http://localhost:9999';

test('General has one company name and HR inherits edits across navigation and reload', async ({
  page,
  context,
}) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['hr'] });
  let companyName = 'Northwind';
  let completed = false;
  const hrWrites: Record<string, unknown>[] = [];
  await page.route(`${API}/hub/config/v1/config`, async (route) => {
    if (route.request().method() === 'PUT') {
      const body = route.request().postDataJSON();
      companyName = body.site_name;
      await route.fulfill({ json: { config: { site_name: companyName }, message: 'Saved' } });
    } else
      await route.fulfill({
        json: { business: { id: '1', name: companyName }, maintenance: false },
      });
  });
  await page.route(`${API}/hr/config`, async (route) => {
    if (route.request().method() === 'PUT') hrWrites.push(route.request().postDataJSON());
    await route.fulfill({
      json: {
        config: {
          id: '1',
          company_name: 'Stale HR name',
          default_currency: 'EUR',
          default_timezone: 'Europe/Berlin',
          directory_enabled: false,
        },
      },
    });
  });
  await page.route(`${API}/hr/setup/status`, (route) =>
    route.fulfill({
      json: {
        is_complete: false,
        steps: completed ? [{ step_key: 'company_profile', completed_at: '1790990000' }] : [],
      },
    })
  );
  await page.route(`${API}/hr/setup/steps/**`, (route) => {
    completed = true;
    return route.fulfill({ json: {} });
  });
  // Prime HR before editing General so this also exercises cached setup data.
  await page.goto(`${PORTAL_URL}/hr/setup`);
  await expect(page.getByLabel('Company name', { exact: true })).toHaveValue('Northwind');
  await page
    .getByRole('navigation', { name: 'Apps', exact: true })
    .getByRole('link', { name: 'Settings', exact: true })
    .click();
  if (page.viewportSize()!.width < 900)
    await page.getByRole('button', { name: 'Toggle navigation menu', exact: true }).click();
  await page
    .locator('.portal-app-navigation:visible')
    .getByRole('link', { name: 'General', exact: true })
    .click();
  const name = page.getByLabel('Company name', { exact: true });
  await expect(name).toHaveCount(1);
  await expect(page.getByLabel('Store name', { exact: true })).toHaveCount(0);
  await name.fill('Northwind Renamed');
  await page.getByRole('button', { name: 'Save changes', exact: true }).click();
  await expect.poll(() => companyName).toBe('Northwind Renamed');
  await page
    .getByRole('navigation', { name: 'Apps', exact: true })
    .getByRole('link', { name: 'HR', exact: true })
    .click();
  await expect(page.getByLabel('Company name', { exact: true })).toHaveValue('Northwind Renamed');
  await expect(page.getByLabel('Company name', { exact: true })).toHaveAttribute('readonly');
  await page.reload();
  await expect(page.getByLabel('Company name', { exact: true })).toHaveValue('Northwind Renamed');
  await expect(page.getByLabel('Timezone', { exact: true })).toHaveValue('Europe/Berlin');
  await expect(page.getByLabel('Currency', { exact: true })).toHaveValue('EUR');
  await page.getByRole('button', { name: 'Next', exact: true }).click();
  await expect
    .poll(() => hrWrites)
    .toEqual([
      {
        company_name: 'Northwind Renamed',
        default_currency: 'EUR',
        default_timezone: 'Europe/Berlin',
      },
    ]);
});

test('HR inherited name and guidance remain usable in Arabic RTL', async ({ page, context }) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['hr'] });
  await context.addCookies([{ name: 'language', value: 'ar', url: PORTAL_URL }]);
  await page.route(`${API}/hr/config`, (route) =>
    route.fulfill({
      json: {
        config: {
          id: '1',
          company_name: 'Stale name',
          default_currency: 'EUR',
          default_timezone: 'Europe/Berlin',
        },
      },
    })
  );
  await page.route(`${API}/hr/setup/status`, (route) =>
    route.fulfill({ json: { steps: [], is_complete: false } })
  );
  await page.goto(`${PORTAL_URL}/hr/setup`);
  const name = page.getByLabel('اسم الشركة', { exact: true });
  await expect(name).toHaveValue('Northwind');
  await expect(name).toHaveAttribute('readonly');
  await expect(page.getByText('يُدار اسم الشركة في الإعدادات العامة.')).toBeVisible();
  await expect(page.locator('html')).toHaveAttribute('dir', 'rtl');
  const bounds = await name.boundingBox();
  expect(bounds!.x).toBeGreaterThanOrEqual(0);
  expect(bounds!.x + bounds!.width).toBeLessThanOrEqual(page.viewportSize()!.width);
  if (page.viewportSize()!.width > 900)
    await page.screenshot({ path: '/tmp/rtylr-hr-shared-identity-rtl.png' });
});
