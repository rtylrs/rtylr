import { expect } from '../fixtures';
import type { Locator } from '@playwright/test';
import {
  contrastRatio,
  deriveActionFill,
  hexToRgb,
} from '../../../packages/utils/src/appearance/colorMath';

/** The app's solid action: its accent's button fill, with text that reads on it. */
export async function expectAppAction(locator: Locator, accent: string): Promise<void> {
  await expect(locator).toBeVisible();
  const fill = deriveActionFill(accent);
  const rgb = hexToRgb(fill);
  await expect(locator).toHaveCSS('background-color', `rgb(${rgb.r}, ${rgb.g}, ${rgb.b})`);
  const foreground = await locator.evaluate((el) => getComputedStyle(el).color);
  const channels = foreground.match(/\d+/g)!.slice(0, 3).map(Number);
  const hex = '#' + channels.map((channel) => channel.toString(16).padStart(2, '0')).join('');
  expect(contrastRatio(hex, fill)).toBeGreaterThanOrEqual(4.5);
}
