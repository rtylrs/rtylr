import { describe, expect, it } from 'vitest';
import { sumBy } from './sumBy';

describe('sumBy', () => {
  it('sums via selector', () => {
    expect(sumBy([1, 2, 3], (n) => n)).toBe(6);
  });
  it('handles empty arrays', () => {
    expect(sumBy([], (n: number) => n)).toBe(0);
  });
});
