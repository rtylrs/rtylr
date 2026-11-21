import { describe, expect, it } from 'vitest';
import { QueryClient } from '@tanstack/react-query';
import type { ApiTask, ApiTaskDetail } from '../../../shared/services/flowApi';
import { patchTaskInCaches } from './useListCellEdit';

const apiTask = (id: string, extra: Partial<ApiTask> = {}): ApiTask =>
  ({ id, projectId: '7', title: `Task ${id}`, priority: 'medium', ...extra }) as ApiTask;

describe('patchTaskInCaches', () => {
  it('updates the task in every list it appears in, matching numeric ids', () => {
    const qc = new QueryClient();
    // The wire sends numeric ids whatever the type says.
    qc.setQueryData(['tasks', { projectId: '7' }], [apiTask(5 as unknown as string), apiTask('6')]);
    qc.setQueryData(['tasks', {}], [apiTask('5')]);

    patchTaskInCaches(qc, '5', { priority: 'urgent' });

    expect(
      qc.getQueryData<ApiTask[]>(['tasks', { projectId: '7' }])!.map((t) => t.priority)
    ).toEqual(['urgent', 'medium']);
    expect(qc.getQueryData<ApiTask[]>(['tasks', {}])![0].priority).toBe('urgent');
  });

  it('updates a subtask inside its parent detail and leaves other details alone', () => {
    const qc = new QueryClient();
    const parent = {
      task: apiTask('1'),
      subtasks: [apiTask('2'), apiTask('3')],
      comments: [],
      activity: [],
    } as unknown as ApiTaskDetail;
    const other = { ...parent, task: apiTask('9'), subtasks: [] } as unknown as ApiTaskDetail;
    qc.setQueryData(['task', '1'], parent);
    qc.setQueryData(['task', '9'], other);

    patchTaskInCaches(qc, '3', { title: 'Renamed' });

    const after = qc.getQueryData<ApiTaskDetail>(['task', '1'])!;
    expect(after.subtasks.map((s) => s.title)).toEqual(['Task 2', 'Renamed']);
    expect(after.task.title).toBe('Task 1');
    expect(qc.getQueryData(['task', '9'])).toBe(other);
  });
});
