import { test, expect } from './fixtures';
import { PORTAL_URL, stubPortalSession } from './portal-session';

test('settings opens as a modal over the current page and closes back to it', async ({
  page,
  context,
}) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['crm', 'hr'] });
  await page.goto(`${PORTAL_URL}/`);
  await expect(page.getByRole('heading', { name: 'Welcome back, Sam' })).toBeVisible();

  await page.getByRole('link', { name: 'Settings', exact: true }).first().click();
  const dialog = page.getByRole('dialog', { name: 'Settings' });
  const menu = dialog.getByRole('navigation', { name: 'Settings' });
  await expect(dialog).toBeVisible();
  // The page stays where it was; settings never becomes the page.
  await expect(page).toHaveURL(`${PORTAL_URL}/`);
  await expect(menu.getByRole('link', { name: 'Account', exact: true })).toBeVisible();
  await page.waitForTimeout(400);
  await page.screenshot({
    path: `/tmp/rtylr-settings-modal-overview-${page.viewportSize()!.width}.png`,
  });

  await menu.getByRole('link', { name: 'Account', exact: true }).click();
  await expect(page).toHaveURL(`${PORTAL_URL}/`);
  await expect(dialog.getByRole('heading', { level: 1 })).toBeVisible();
  await page.screenshot({
    path: `/tmp/rtylr-settings-modal-account-${page.viewportSize()!.width}.png`,
  });

  await page.keyboard.press('Escape');
  await expect(dialog).not.toBeVisible();
  await expect(page.getByRole('heading', { name: 'Welcome back, Sam' })).toBeVisible();
});

test('a settings bookmark opens the modal on that section', async ({ page, context }) => {
  await stubPortalSession(page, context, { isOwner: true, appIds: ['crm'] });
  await page.goto(`${PORTAL_URL}/settings/billing`);
  await expect(page.getByRole('dialog', { name: 'Settings' })).toBeVisible();
  await expect(page).toHaveURL(`${PORTAL_URL}/`);
});
