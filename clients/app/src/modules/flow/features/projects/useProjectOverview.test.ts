import { describe, it, expect, vi, afterEach } from 'vitest';
import type { ApiStatus, ApiTask } from '../../shared/services/flowApi';
import { projectTaskStats } from './useProjectOverview';

function task(over: Partial<ApiTask>): ApiTask {
  return {
    id: '1',
    projectId: '3',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'RTYLR',
    ...over,
  };
}

const shipped: ApiStatus = {
  id: 's9',
  projectId: '3',
  name: 'Shipped',
  bucket: 'done',
  legacyStatus: 'in_review',
  color: '#000',
  sortOrder: 4,
  isDefault: false,
  createdAt: '2026-01-01T00:00:00Z',
};

afterEach(() => vi.useRealTimers());

describe('projectTaskStats', () => {
  it('counts top-level tasks that are not archived, leaving cancelled out of the total', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 10, 25, 12));
    const stats = projectTaskStats(
      [
        task({ id: '1', status: 'done' }),
        // Done through a custom status whose legacy string is not "done".
        task({ id: '2', status: 'in_review', statusId: 's9' }),
        task({ id: '3', status: 'in_progress', dueDate: '2026-11-20' }),
        task({ id: '4', status: 'todo' }),
        task({ id: '5', status: 'cancelled', dueDate: '2026-11-01' }),
        task({ id: '6', status: 'done', archived: true }),
        task({ id: '7', status: 'todo', parentTaskId: '4' }),
      ],
      [shipped]
    );
    expect(stats).toEqual({ total: 4, done: 2, inProgress: 1, overdue: 1, progress: 50 });
  });
});
