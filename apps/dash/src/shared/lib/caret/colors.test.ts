import { describe, expect, it } from 'vitest';
import { paletteColor } from './colors';

describe('paletteColor', () => {
  it('returns a defined palette entry', () => {
    const c = paletteColor('hello');
    expect(c.bg).toMatch(/rgba/);
    expect(c.fg.startsWith('#')).toBe(true);
  });
  it('is deterministic for the same seed', () => {
    expect(paletteColor('a')).toEqual(paletteColor('a'));
  });
});
