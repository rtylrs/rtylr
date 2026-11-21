import { describe, it, expect } from 'vitest';
import type { Task } from '../../shared/stores/flowStore';
import { NO_EPIC } from '../../shared/lib/taskHierarchy';
import { epicLanes, laneKeyFor } from './taskGroups';

function task(overrides: Partial<Task> = {}): Task {
  return {
    id: '1',
    projectId: 'p1',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    assigneeId: '',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Project',
    createdAt: '2026-11-01',
    updatedAt: '2026-11-01',
    identifier: 'T-1',
    ...overrides,
  };
}

describe('epic swimlanes', () => {
  const epic = task({ id: '10', taskType: 'epic', title: 'Payments', epicColor: '#22c55e' });
  const child = task({ id: '11', epicId: '10' });
  const loose = task({ id: '12' });

  it('gives each epic a lane, in board order, then No epic', () => {
    expect(epicLanes([loose, epic, child])).toEqual([
      { key: '10', label: 'Payments', color: '#22c55e' },
      { key: NO_EPIC, labelKey: 'flow.epics.noEpic' },
    ]);
  });

  it('puts an epic and its tasks in the epic lane and the rest in No epic', () => {
    expect(laneKeyFor(epic, 'epic')).toBe('10');
    expect(laneKeyFor(child, 'epic')).toBe('10');
    expect(laneKeyFor(loose, 'epic')).toBe(NO_EPIC);
  });

  it('still has a No epic lane on a board without epics', () => {
    expect(epicLanes([loose])).toEqual([{ key: NO_EPIC, labelKey: 'flow.epics.noEpic' }]);
  });
});
