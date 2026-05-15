import { describe, expect, it } from 'vitest';
import { groupChatsByDate } from './groupByDate';
import type { CaretChatSummary } from '@voxire/caret';

const NOW = new Date('2026-05-16T15:00:00Z').getTime();

const mk = (id: string, hoursAgo: number): CaretChatSummary => ({
  id,
  title: id,
  updatedAt: new Date(NOW - hoursAgo * 3_600_000).toISOString(),
});

describe('groupChatsByDate', () => {
  it('drops empty buckets', () => {
    const out = groupChatsByDate([mk('a', 1)], NOW);
    expect(out.length).toBe(1);
    expect(out[0].group).toBe('today');
  });

  it('keeps bucket ordering today → older', () => {
    const out = groupChatsByDate(
      [mk('a', 1), mk('b', 26), mk('c', 24 * 5), mk('d', 24 * 20), mk('e', 24 * 100)],
      NOW
    );
    expect(out.map((g) => g.group)).toEqual(['today', 'yesterday', 'last7', 'last30', 'older']);
  });
});
