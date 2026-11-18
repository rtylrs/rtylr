import { describe, expect, it } from 'vitest';
import { EMPTY_TASK_FILTERS } from './taskFilters';
import { taskFiltersToViewFilters, viewFiltersToTaskFilters } from './viewFilters';

describe('view filters', () => {
  it('reads a view saved with the old single-value keys', () => {
    const filters = viewFiltersToTaskFilters({
      status: ['todo'],
      priority: ['high'],
      assigneeId: '12',
      projectId: '4',
      labels: ['7'],
    });
    expect(filters).toEqual({
      ...EMPTY_TASK_FILTERS,
      statuses: ['todo'],
      priorities: ['high'],
      assignees: ['12'],
      projects: ['4'],
      labels: ['7'],
    });
  });

  it('prefers the multi-value keys when a view has both', () => {
    const filters = viewFiltersToTaskFilters({ assigneeId: '12', assignees: ['12', '13'] });
    expect(filters.assignees).toEqual(['12', '13']);
  });

  it('writes the new keys and clears the old ones, and reads back the same', () => {
    const model = {
      ...EMPTY_TASK_FILTERS,
      assignees: ['12', 'none'],
      projects: ['4'],
      due: ['overdue' as const],
      dueTo: '2026-12-31',
    };
    const stored = taskFiltersToViewFilters(model);
    expect(stored.assigneeId).toBeUndefined();
    expect(stored.projectId).toBeUndefined();
    expect(viewFiltersToTaskFilters(JSON.parse(JSON.stringify(stored)))).toEqual(model);
  });

  it('never carries a search into a saved view', () => {
    const stored = taskFiltersToViewFilters({ ...EMPTY_TASK_FILTERS, query: 'tax' });
    expect(JSON.stringify(stored)).not.toContain('tax');
  });

  it('treats a missing filter object as no filters', () => {
    expect(viewFiltersToTaskFilters(undefined)).toEqual(EMPTY_TASK_FILTERS);
  });
});
