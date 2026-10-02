import { test, expect } from './fixtures';
import { PORTAL_URL, stubPortalSession } from './portal-session';

test.describe('Portal Home', () => {
  test('the command launcher stays centered and visible on desktop and mobile', async ({
    page,
    context,
  }) => {
    await stubPortalSession(page, context, { isOwner: true });
    await page.goto(PORTAL_URL);
    await expect(page.getByRole('heading', { name: 'Welcome back, Sam' })).toBeVisible();
    await page.getByRole('button', { name: 'Open command palette' }).click();
    const dialog = page.getByRole('dialog', { name: 'Search Portal' });
    await expect(dialog).toBeVisible();
    const bounds = await dialog.boundingBox();
    const viewport = page.viewportSize();
    expect(bounds).not.toBeNull();
    expect(viewport).not.toBeNull();
    if (!bounds || !viewport) throw new Error('Missing launcher bounds');
    expect(Math.abs(bounds.x + bounds.width / 2 - viewport.width / 2)).toBeLessThan(2);
    expect(Math.abs(bounds.y + bounds.height / 2 - viewport.height / 2)).toBeLessThan(2);
    expect(bounds.y).toBeGreaterThanOrEqual(0);
    expect(bounds.y + bounds.height).toBeLessThanOrEqual(viewport.height);
    await expect(page.getByRole('combobox')).toBeFocused();
    await page.getByRole('combobox').fill('reports');
    await expect(
      page
        .getByRole('group', { name: 'Insights', exact: true })
        .getByRole('option', { name: 'Reports', exact: true })
    ).toBeVisible();
    await page.getByRole('combobox').press('Escape');
    await expect(dialog).toHaveCount(0);
    await expect(page.getByRole('button', { name: 'Open command palette' })).toBeFocused();
  });

  test('a one-app member stays on Home and sees only assigned access', async ({
    page,
    context,
  }) => {
    await stubPortalSession(page, context);
    await page.goto(PORTAL_URL);
    await expect(page.getByRole('heading', { name: 'Welcome back, Sam' })).toBeVisible();
    await expect(page).toHaveURL(`${PORTAL_URL}/`);
    await expect(page.getByRole('link', { name: 'Open Insights', exact: true })).toHaveAttribute(
      'href',
      '/insights'
    );
    await expect(page.getByRole('link', { name: /Open Flow/ })).toHaveCount(0);
    await expect(page.getByRole('link', { name: 'Team', exact: true })).toHaveCount(0);
    await expect(page.getByRole('link', { name: 'Groups', exact: true })).toHaveCount(0);
    await page.goto(`${PORTAL_URL}/team`);
    await expect(page.getByRole('alert')).toContainText('You don’t have access to this page');
  });

  test('an owner sees management shortcuts and Recruit administration in Home', async ({
    page,
    context,
  }) => {
    await stubPortalSession(page, context, {
      isOwner: true,
      appIds: ['insights', 'menu', 'recruit'],
    });
    await page.goto(PORTAL_URL);
    await expect(page.getByRole('heading', { name: 'Welcome back, Sam' })).toBeVisible();
    await expect(page.getByRole('link', { name: /Members and access/ })).toBeVisible();
    await expect(page.getByRole('link', { name: /Shared permissions/ })).toBeVisible();
    await expect(page.getByRole('link', { name: /Open Menu/ })).toHaveCount(0);
    await expect(page.getByRole('link', { name: /Open Recruit/ })).toHaveAttribute(
      'href',
      '/recruit'
    );
  });
});

test('Home opens every remaining business app inside Portal', async ({ page, context }) => {
  await stubPortalSession(page, context, {
    isOwner: true,
    appIds: ['finance', 'hr', 'pos', 'recruit'],
  });
  await page.goto(PORTAL_URL);
  const homeApps = page.locator('.portal-home-app-grid');
  for (const id of ['finance', 'hr', 'pos', 'recruit']) {
    const link = homeApps.locator(`a[href="/${id}"]`);
    await expect(link).toBeVisible();
    await expect(link).not.toHaveAttribute('target', '_blank');
  }
});
