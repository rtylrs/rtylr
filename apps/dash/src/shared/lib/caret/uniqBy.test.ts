import { describe, expect, it } from 'vitest';
import { uniqBy } from './uniqBy';

describe('uniqBy', () => {
  it('keeps the first occurrence per key', () => {
    const out = uniqBy(
      [
        { id: 1, v: 'a' },
        { id: 2, v: 'b' },
        { id: 1, v: 'c' },
      ],
      (x) => x.id
    );
    expect(out).toEqual([
      { id: 1, v: 'a' },
      { id: 2, v: 'b' },
    ]);
  });
});
