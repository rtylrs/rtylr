import { test, expect, type Route } from './fixtures';
import { captureForPR } from './support/screenshot';

/**
 * Screenshot matrix for the Auth showcase.
 *
 * The rtylr#1386 review asked for runtime evidence across languages, widths and
 * themes. The showcase is `hidden lg:flex`, so at 375/768 it must not render at
 * all; the captures record 1024 and 1440 where it does, in English and Arabic,
 * light and dark. Tagged `@screenshots`, so the normal e2e run ignores the
 * artefacts: `pnpm -C clients/web/rtylr test:screenshots -- --grep auth-showcase`.
 */
const AUTH_URL = process.env.AUTH_URL ?? 'http://localhost:5176';

const WIDTHS = [375, 768, 1024, 1440];
const SHOWCASE_MIN_WIDTH = 1024;
const LANGUAGES = ['en', 'ar'] as const;
const THEMES = ['light', 'dark'] as const;

const TAGLINE = {
  en: 'Rtylr suite, Everything you need to run your operation, unified in a single workspace',
  ar: 'منظومة Rtylr، كل ما تحتاجه لتشغيل أعمالك، موحّد في مساحة عمل واحدة',
};

test.describe('auth-showcase responsive matrix @screenshots', () => {
  test.beforeEach(async ({ page }) => {
    // The sign-in route reads runtime config and hub features on load. Answer
    // them so no request is in flight while the capture settles.
    await page.route(/localhost:9999\//, (route: Route) =>
      route.fulfill({ status: 200, contentType: 'application/json', body: '{}' })
    );
  });

  for (const language of LANGUAGES) {
    for (const theme of THEMES) {
      test(`${language} ${theme}`, async ({ context, page }, testInfo) => {
        // Every width is set explicitly below, so the 390px project would only
        // repeat the same captures.
        test.skip(testInfo.project.name === 'chromium-390', 'widths are set explicitly');

        await context.addCookies([
          {
            name: 'language',
            value: language,
            domain: 'localhost',
            path: '/',
            httpOnly: false,
            sameSite: 'Lax',
            secure: false,
          },
        ]);
        await page.emulateMedia({ colorScheme: theme });
        await page.goto(AUTH_URL);
        await expect(page.locator('input[type="password"]').first()).toBeVisible({
          timeout: 10_000,
        });

        for (const width of WIDTHS) {
          await page.setViewportSize({ width, height: 900 });

          const showcase = page.locator('aside');
          if (width >= SHOWCASE_MIN_WIDTH) {
            await expect(showcase).toBeVisible();
            await expect(showcase.getByText(TAGLINE[language], { exact: false })).toBeVisible();
          } else {
            await expect(showcase).toBeHidden();
          }

          // The seven product pills are gone at every width.
          await expect(page.getByText('Insights', { exact: true })).toHaveCount(0);
          await expect(page.getByText('POS', { exact: true })).toHaveCount(0);

          // Measured, not eyeballed: a screenshot cannot show sideways scroll.
          const overflows = await page.evaluate(
            () => document.documentElement.scrollWidth > document.documentElement.clientWidth
          );
          expect(overflows, `${language}/${theme} overflows horizontally at ${width}px`).toBe(
            false
          );

          await captureForPR(page, `auth-showcase-${language}-${theme}-${width}`);
        }

        await expect(page.locator('html')).toHaveAttribute(
          'dir',
          language === 'ar' ? 'rtl' : 'ltr'
        );
      });
    }
  }
});
