import { test, expect } from './fixtures';
import { PORTAL_URL, stubPortalSession } from './portal-session';
import {
  contrastRatio,
  compositeOver,
  hexToRgb,
} from '../../packages/utils/src/appearance/colorMath';

const apps = { flow: '#64c5b9', crm: '#3b82f6', erp: '#4583c4', hr: '#8b5cf6', finance: '#10b981' };
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
      await page.goto(`${PORTAL_URL}/settings/${app}`);
      const shell = page.locator('.portal-shell');
      await expect(shell).toBeVisible();
      const style = await shell.evaluate((el) => {
        const css = getComputedStyle(el);
        const active = el.querySelector('.portal-nav-link.is-active')!;
        return {
          accent: css.getPropertyValue('--accent').trim(),
          fg: css.getPropertyValue('--accent-on-muted').trim(),
          focus: css.getPropertyValue('--accent-focus').trim(),
          activeColor: getComputedStyle(active).color,
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
      const fg = hexToRgb(style.fg);
      expect(style.activeColor).toBe(`rgb(${fg.r}, ${fg.g}, ${fg.b})`);
    }
    await page
      .getByRole('navigation', { name: 'Apps', exact: true })
      .getByRole('link', { name: 'Home', exact: true })
      .click();
    await expect(page.getByRole('heading', { name: 'Welcome back, Sam' })).toBeVisible();
    await expect(page.locator('.portal-shell')).not.toHaveAttribute('data-app');
  });
}
