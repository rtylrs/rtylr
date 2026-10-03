import { expect } from '../fixtures';
import type { Locator } from '@playwright/test';
import { contrastRatio, hexToRgb } from '../../../packages/utils/src/appearance/colorMath';

export async function expectAppAction(locator: Locator, accent: string): Promise<void> {
  await expect(locator).toBeVisible();
  const rgb = hexToRgb(accent);
  await expect(locator).toHaveCSS('background-color', `rgb(${rgb.r}, ${rgb.g}, ${rgb.b})`);
  const foreground = await locator.evaluate((el) => getComputedStyle(el).color);
  const channels = foreground.match(/\d+/g)!.slice(0, 3).map(Number);
  const hex = '#' + channels.map((channel) => channel.toString(16).padStart(2, '0')).join('');
  expect(contrastRatio(hex, accent)).toBeGreaterThanOrEqual(4.5);
}
