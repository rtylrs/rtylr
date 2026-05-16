import { describe, expect, it } from 'vitest';
import { clampPanelWidth } from './clampWidth';

describe('clampPanelWidth', () => {
  it('clamps below the minimum', () => {
    expect(clampPanelWidth(100, 1600)).toBe(320);
  });
  it('clamps above the window-relative maximum', () => {
    expect(clampPanelWidth(2000, 1000)).toBe(600);
  });
  it('passes through reasonable widths', () => {
    expect(clampPanelWidth(420, 1600)).toBe(420);
  });
});
