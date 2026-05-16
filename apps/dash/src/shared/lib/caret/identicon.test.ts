import { describe, expect, it } from 'vitest';
import { identicon } from './identicon';

describe('identicon', () => {
  it('returns an SVG string', () => {
    const svg = identicon('seed');
    expect(svg.startsWith('<svg')).toBe(true);
    expect(svg.endsWith('</svg>')).toBe(true);
  });
  it('is deterministic for the same seed', () => {
    expect(identicon('a')).toBe(identicon('a'));
  });
});
