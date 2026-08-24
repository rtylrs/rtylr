import type { Page } from '@playwright/test';
import fs from 'node:fs';
import path from 'node:path';

/**
 * Capture a screenshot for a PR body.
 *
 * The UI standards addendum makes screenshots part of the definition of done,
 * and the Flow workspace-removal tracker recorded C2 as "cannot be delivered"
 * on the belief that no image file could be produced from an agent
 * environment. That was wrong: the e2e harness already runs a real Chromium
 * against real dev servers, so `page.screenshot()` writes a real PNG. This
 * helper exists so nobody has to rediscover that.
 *
 * Two things it does that a bare `page.screenshot()` does not, both of which
 * produced misleading images before:
 *
 * 1. **Finishes running animations.** Entrance transitions stall when the
 *    browser surface is not visible, so a screenshot taken mid-transition
 *    shows half-faded content at partial opacity. The same stall made an axe
 *    contrast audit report 29 failures where 3 were real, because it measured
 *    colours mid-fade.
 * 2. **Waits for fonts.** Without it the first shot of a run can render in the
 *    fallback face, which makes a design review argue about the wrong type.
 *
 * Files land in `e2e/__screenshots__/`, which is gitignored: they are review
 * artefacts, not fixtures, and nothing asserts against them. See
 * `docs/pr-screenshots.md` for how to get them into a PR body.
 */
export async function captureForPR(page: Page, name: string): Promise<string> {
  await page.evaluate(async () => {
    document.getAnimations().forEach((animation) => animation.finish());
    await document.fonts.ready;
  });

  const dir = path.join(process.cwd(), 'e2e', '__screenshots__');
  fs.mkdirSync(dir, { recursive: true });
  const file = path.join(dir, `${name}.png`);
  await page.screenshot({ path: file });
  return file;
}

/**
 * Capture the same screen at desktop and at the 390px mobile floor.
 *
 * Mobile-first is non-negotiable here and 390px is the stated floor, so a
 * review that only ever sees the desktop shot cannot tell whether the rule was
 * followed. Returns both paths plus whether the page overflows horizontally,
 * which is the single most common mobile regression and is measured rather than
 * eyeballed off the image.
 */
export async function captureResponsivePair(
  page: Page,
  name: string
): Promise<{ desktop: string; mobile: string; overflowsAt390: boolean }> {
  await page.setViewportSize({ width: 1280, height: 900 });
  const desktop = await captureForPR(page, `${name}-desktop`);

  await page.setViewportSize({ width: 390, height: 844 });
  const mobile = await captureForPR(page, `${name}-mobile`);

  const overflowsAt390 = await page.evaluate(
    () => document.documentElement.scrollWidth > document.documentElement.clientWidth
  );

  return { desktop, mobile, overflowsAt390 };
}
