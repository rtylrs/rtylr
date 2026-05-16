import { describe, expect, it } from 'vitest';
import { groupBy } from './groupBy';

describe('groupBy', () => {
  it('buckets by the key function', () => {
    const out = groupBy([{ n: 1 }, { n: 2 }, { n: 1 }], (i) => String(i.n));
    expect(out['1']).toHaveLength(2);
    expect(out['2']).toHaveLength(1);
  });
});
