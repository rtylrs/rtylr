import { describe, expect, it } from 'vitest';
import { diffById } from './diff';

describe('diffById', () => {
  it('returns added/removed/unchanged sets', () => {
    const prev = [{ id: 1 }, { id: 2 }];
    const next = [{ id: 2 }, { id: 3 }];
    const out = diffById(prev, next, (x) => x.id);
    expect(out.added).toEqual([{ id: 3 }]);
    expect(out.removed).toEqual([{ id: 1 }]);
    expect(out.unchanged).toEqual([{ id: 2 }]);
  });
});
