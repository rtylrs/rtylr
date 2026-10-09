import { expectAppAction } from './support/app-appearance';
import { test, expect } from './fixtures';
import { PORTAL_URL, stubPortalSession } from './portal-session';
import {
  contrastRatio,
  compositeOver,
  hexToRgb,
} from '../../packages/utils/src/appearance/colorMath';

const apps = {
  insights: '#6fa3db',
  flow: '#ef4444',
  crm: '#8b5cf6',
  erp: '#64c5b9',
  hr: '#10b981',
  finance: '#f59e0b',
};
for (const theme of ['light', 'dark'] as const) {
  test(`app navigation retains distinct accessible colors in ${theme} and returns Home to neutral`, async ({
    page,
    context,
  }) => {
    await page.addInitScript(
      (theme) =>
        localStorage.setItem('portal-theme', JSON.stringify({ state: { theme }, version: 0 })),
      theme
    );
    await stubPortalSession(page, context, { isOwner: true, appIds: Object.keys(apps) });
    // Business data is irrelevant to this navigation check; provide empty API lists.
    await page.route(/localhost:9999\/(?:hr|crm|erp|finance|flow)\//, (route) =>
      route.fulfill({
        json: {
          config: null,
          steps: [],
          projects: [],
          tasks: [],
          workspaces: [],
          labels: [],
          notifications: [],
          employees: [],
          transactions: [],
          accounts: [],
          contacts: [],
          items: [],
          total: 0,
        },
      })
    );
    for (const [app, accent] of Object.entries(apps)) {
      await page.goto(`${PORTAL_URL}/${app}`);
      const shell = page.locator('.portal-shell');
      await expect(shell).toBeVisible();
      const style = await shell.evaluate((el) => {
        const css = getComputedStyle(el);
        // Insights and Flow use their own headers; the other apps have a sidebar link.
        const active = el.querySelector('.portal-nav-link.is-active');
        return {
          accent: css.getPropertyValue('--accent').trim(),
          fg: css.getPropertyValue('--accent-on-muted').trim(),
          focus: css.getPropertyValue('--accent-focus').trim(),
          activeColor: active ? getComputedStyle(active).color : null,
          theme: document.documentElement.dataset.theme,
        };
      });
      expect(style.accent).toBe(accent);
      expect(style.theme).toBe(theme);
      for (const ground of theme === 'light'
        ? ['#ffffff', '#f8f8f8']
        : ['#0a0a0a', '#111111', '#161616']) {
        expect(contrastRatio(style.fg, compositeOver(accent, 0.12, ground))).toBeGreaterThanOrEqual(
          4.5
        );
        expect(contrastRatio(style.focus, ground)).toBeGreaterThanOrEqual(3);
      }
      if (style.activeColor) {
        const fg = hexToRgb(style.fg);
        expect(style.activeColor).toBe(`rgb(${fg.r}, ${fg.g}, ${fg.b})`);
      }
    }
    await page.goto(`${PORTAL_URL}/hr/setup`);
    await expectAppAction(page.getByRole('button', { name: 'Next', exact: true }), apps.hr);

    await page.goto(`${PORTAL_URL}/flow/my-issues`);
    const newTask = page
      .getByRole('button', { name: 'New task', exact: true })
      .filter({ hasText: 'New task' });
    await expectAppAction(newTask, apps.flow);
    await newTask.click();
    const dialog = page.getByRole('dialog', { name: 'New task', exact: true });
    await expect(dialog).toBeVisible();
    expect(await dialog.evaluate((el) => Boolean(el.closest('.portal-shell')))).toBe(false);
    await dialog.getByRole('textbox', { name: 'Task name', exact: true }).fill('Review app colors');
    await expectAppAction(
      dialog.getByRole('button', { name: 'Create task', exact: true }),
      apps.flow
    );
    await page.screenshot({
      path: `/tmp/rtylr-flow-identity-${theme}-${page.viewportSize()!.width}.png`,
    });
    await page.keyboard.press('Escape');
    await expect(dialog).not.toBeVisible();
    await expect(newTask).toBeFocused();

    await page
      .getByRole('navigation', { name: 'Apps', exact: true })
      .getByRole('link', { name: 'Home', exact: true })
      .click();
    await expect(page.getByRole('heading', { name: 'Welcome back, Sam' })).toBeVisible();
    await expect(page.locator('.portal-shell')).not.toHaveAttribute('data-app');
  });
}
