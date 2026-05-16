import { describe, expect, it } from 'vitest';
import { maxBy } from './maxBy';

describe('maxBy', () => {
  it('returns the max by selector', () => {
    expect(maxBy([{ n: 1 }, { n: 5 }, { n: 3 }], (x) => x.n)).toEqual({ n: 5 });
  });
  it('returns undefined for empty', () => {
    expect(maxBy([] as { n: number }[], (x) => x.n)).toBeUndefined();
  });
});
