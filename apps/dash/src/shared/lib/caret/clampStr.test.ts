import { describe, expect, it } from 'vitest';
import { clampString } from './clampStr';

describe('clampString', () => {
  it('returns short strings unchanged', () => {
    expect(clampString('hi', 10)).toBe('hi');
  });
  it('truncates and appends ellipsis', () => {
    expect(clampString('hello world', 8)).toBe('hello w…');
  });
  it('handles tight limits', () => {
    expect(clampString('abc', 1)).toBe('…');
  });
});
