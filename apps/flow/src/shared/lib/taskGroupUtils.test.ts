import { describe, it, expect } from 'vitest';

import {
  groupByProject,
  groupByAssignee,
  groupByCycle,
  groupByMilestone,
  groupByParent,
  rootTasks,
  childTasks,
} from './taskGroupUtils';
import type { Task } from '../stores/flowStore';

function makeTask(overrides: Partial<Task>): Task {
  return {
    id: 't1',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    assigneeId: 'u1',
    projectId: 'p1',
    workspaceId: 'w1',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Project',
    createdAt: '2026-01-01',
    updatedAt: '2026-01-01',
    identifier: 'FLOW-1',
    ...overrides,
  } as Task;
}

const tasks = [
  makeTask({ id: 'a', projectId: 'p1', assigneeId: 'u1' }),
  makeTask({ id: 'b', projectId: 'p2', assigneeId: 'u2' }),
  makeTask({ id: 'c', projectId: 'p1', assigneeId: 'u1' }),
];

describe('groupByProject', () => {
  it('groups tasks by project', () => {
    const map = groupByProject(tasks);
    expect(map.get('p1')).toHaveLength(2);
    expect(map.get('p2')).toHaveLength(1);
  });
});

describe('groupByAssignee', () => {
  it('groups tasks by assignee', () => {
    const map = groupByAssignee(tasks);
    expect(map.get('u1')).toHaveLength(2);
    expect(map.get('u2')).toHaveLength(1);
  });

  it('uses "unassigned" for tasks without assigneeId', () => {
    const t = makeTask({ assigneeId: '' });
    const map = groupByAssignee([t]);
    expect(map.get('unassigned')).toHaveLength(1);
  });
});

describe('groupByCycle', () => {
  it('groups by cycle ID', () => {
    const cycleTasks = [
      makeTask({ id: 'a', cycleId: 'c1' }),
      makeTask({ id: 'b', cycleId: 'c2' }),
      makeTask({ id: 'c', cycleId: 'c1' }),
    ];
    const map = groupByCycle(cycleTasks);
    expect(map.get('c1')).toHaveLength(2);
    expect(map.get('c2')).toHaveLength(1);
  });

  it('uses "none" for tasks without a cycle', () => {
    const t = makeTask({});
    const map = groupByCycle([t]);
    expect(map.get('none')).toHaveLength(1);
  });
});

describe('groupByMilestone', () => {
  it('groups by milestone ID', () => {
    const ms = [makeTask({ id: 'a', milestoneId: 'm1' }), makeTask({ id: 'b' })];
    const map = groupByMilestone(ms);
    expect(map.get('m1')).toHaveLength(1);
    expect(map.get('none')).toHaveLength(1);
  });
});

describe('groupByParent', () => {
  it('groups by parent ID', () => {
    const mixed = [makeTask({ id: 'a', parentId: 'parent1' }), makeTask({ id: 'b' })];
    const map = groupByParent(mixed);
    expect(map.get('parent1')).toHaveLength(1);
    expect(map.get('root')).toHaveLength(1);
  });
});

describe('rootTasks', () => {
  it('returns tasks without a parent', () => {
    const mixed = [makeTask({ id: 'a' }), makeTask({ id: 'b', parentId: 'a' })];
    expect(rootTasks(mixed).map((t) => t.id)).toEqual(['a']);
  });
});

describe('childTasks', () => {
  it('returns child tasks of a parent', () => {
    const mixed = [
      makeTask({ id: 'a' }),
      makeTask({ id: 'b', parentId: 'a' }),
      makeTask({ id: 'c', parentId: 'a' }),
    ];
    expect(childTasks(mixed, 'a').map((t) => t.id)).toEqual(['b', 'c']);
  });
});
