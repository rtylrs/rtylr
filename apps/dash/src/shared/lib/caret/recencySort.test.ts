import { describe, expect, it } from 'vitest';
import { sortByRecency } from './recencySort';
import type { CaretChatSummary } from '@voxire/caret';

const mk = (id: string, updatedAt: string, pinned = false): CaretChatSummary => ({
  id,
  title: id,
  updatedAt,
  pinned,
});

describe('sortByRecency', () => {
  it('pinned floats above unpinned', () => {
    const out = sortByRecency([
      mk('a', '2026-05-01T00:00:00Z'),
      mk('b', '2026-04-01T00:00:00Z', true),
    ]);
    expect(out[0].id).toBe('b');
  });

  it('orders by updatedAt descending', () => {
    const out = sortByRecency([mk('a', '2026-04-01T00:00:00Z'), mk('b', '2026-05-01T00:00:00Z')]);
    expect(out.map((c) => c.id)).toEqual(['b', 'a']);
  });
});
