import { describe, it, expect } from 'vitest';
import { buildChips } from './FilterBar';
import { EMPTY_TASK_FILTERS } from './taskFilters';

const t = (key: string): string => key;

describe('recent change chips', () => {
  it('shows one chip per window that removes only itself', () => {
    const chips = buildChips(
      { ...EMPTY_TASK_FILTERS, changed: ['created_7d', 'updated_7d'] },
      [],
      t
    );
    expect(chips.map((c) => c.label)).toEqual([
      'flow.filters.changed.created7d',
      'flow.filters.changed.updated7d',
    ]);
    expect(chips[0].next.changed).toEqual(['updated_7d']);
  });
});
