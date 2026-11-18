import { describe, expect, it, vi } from 'vitest';
import { EMPTY_TASK_FILTERS } from '../../shared/filters/taskFilters';
import type { ApiStatus } from '../../shared/services/flowApi';

vi.mock('./useViews', () => ({ useCreateView: () => ({ mutate: vi.fn(), isPending: false }) }));

import { projectFiltersForView } from './useSaveFiltersAsView';

const status = (id: string, legacyStatus: string): ApiStatus =>
  ({ id, legacyStatus, name: id, projectId: '4' }) as unknown as ApiStatus;

describe('projectFiltersForView', () => {
  it('adds the project, turns columns into built-in statuses and drops the search', () => {
    const out = projectFiltersForView(
      { ...EMPTY_TASK_FILTERS, statuses: ['31', '32', '33'], priorities: ['high'], query: 'tax' },
      '4',
      [status('31', 'in_progress'), status('32', 'in_progress'), status('33', 'in_review')]
    );
    expect(out.projects).toEqual(['4']);
    expect(out.statuses).toEqual(['in_progress', 'in_review']);
    expect(out.priorities).toEqual(['high']);
    expect(out.query).toBe('');
  });
});
