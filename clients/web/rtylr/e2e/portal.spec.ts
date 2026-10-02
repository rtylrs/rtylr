import { test, expect } from './fixtures';
import { PORTAL_URL, stubPortalSession } from './portal-session';

test.describe('Portal Home', () => {
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

  test('an owner sees management shortcuts without public setup products in the launcher', async ({
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
    await expect(page.getByRole('link', { name: /Open Menu|Open Recruit/ })).toHaveCount(0);
  });
});
