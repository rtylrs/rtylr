import { describe, it, expect } from 'vitest';

import {
  tasksForMember,
  completedTasksForMember,
  memberCompletionRate,
  totalLoggedMinutesForMember,
  averageEstimateHours,
  taskCountByMember,
  membersByWorkload,
  projectsByCompletion,
  overdueCount,
  estimatedTaskPercent,
} from './insightUtils';
import type { Task, Project, Member, TimeLog } from '../stores/flowStore';

function makeTask(overrides: Partial<Task>): Task {
  return {
    id: 't1',
    title: 'Task',
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    assigneeId: '',
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

function makeProject(overrides: Partial<Project>): Project {
  return {
    id: 'p1',
    workspaceId: 'w1',
    name: 'Project',
    color: '#3b82f6',
    taskCount: 10,
    completedCount: 5,
    memberAvatars: [],
    health: 'on_track',
    progress: 50,
    startDate: '2026-01-01',
    targetDate: '2026-12-31',
    lead: '',
    archived: false,
    ...overrides,
  } as Project;
}

function makeMember(overrides: Partial<Member>): Member {
  return {
    id: 'm1',
    name: 'Alice',
    email: 'alice@example.com',
    initials: 'A',
    role: 'member',
    capacity: 40,
    ...overrides,
  } as Member;
}

function makeLog(overrides: Partial<TimeLog>): TimeLog {
  return {
    id: 'l1',
    taskId: 't1',
    userId: 'u1',
    minutes: 60,
    loggedAt: '2026-01-01',
    ...overrides,
  } as TimeLog;
}

describe('tasksForMember', () => {
  it('returns tasks assigned to the member', () => {
    const tasks = [
      makeTask({ id: 'a', assigneeId: 'm1' }),
      makeTask({ id: 'b', assigneeId: 'm2' }),
    ];
    expect(tasksForMember(tasks, 'm1').map((t) => t.id)).toEqual(['a']);
  });
});

describe('completedTasksForMember', () => {
  it('counts done tasks for the member', () => {
    const tasks = [
      makeTask({ assigneeId: 'm1', status: 'done' }),
      makeTask({ assigneeId: 'm1', status: 'todo' }),
      makeTask({ assigneeId: 'm2', status: 'done' }),
    ];
    expect(completedTasksForMember(tasks, 'm1')).toBe(1);
  });
});

describe('memberCompletionRate', () => {
  it('returns percentage of done tasks', () => {
    const tasks = [
      makeTask({ id: 'a', assigneeId: 'm1', status: 'done' }),
      makeTask({ id: 'b', assigneeId: 'm1', status: 'todo' }),
    ];
    expect(memberCompletionRate(tasks, 'm1')).toBe(50);
  });

  it('returns 0 for member with no tasks', () => {
    expect(memberCompletionRate([], 'm1')).toBe(0);
  });
});

describe('totalLoggedMinutesForMember', () => {
  it('sums minutes for the member', () => {
    const logs = [
      makeLog({ userId: 'm1', minutes: 30 }),
      makeLog({ userId: 'm1', minutes: 45 }),
      makeLog({ userId: 'm2', minutes: 100 }),
    ];
    expect(totalLoggedMinutesForMember(logs, 'm1')).toBe(75);
  });
});

describe('averageEstimateHours', () => {
  it('returns average estimate for project tasks', () => {
    const tasks = [
      makeTask({ projectId: 'p1', estimate: 2 }),
      makeTask({ projectId: 'p1', estimate: 4 }),
      makeTask({ projectId: 'p2', estimate: 10 }),
    ];
    expect(averageEstimateHours(tasks, 'p1')).toBe(3);
  });

  it('returns 0 when no tasks have estimates', () => {
    expect(averageEstimateHours([], 'p1')).toBe(0);
  });
});

describe('taskCountByMember', () => {
  it('counts tasks per member', () => {
    const tasks = [
      makeTask({ assigneeId: 'm1' }),
      makeTask({ assigneeId: 'm1' }),
      makeTask({ assigneeId: 'm2' }),
    ];
    const map = taskCountByMember(tasks);
    expect(map.get('m1')).toBe(2);
    expect(map.get('m2')).toBe(1);
  });

  it('skips tasks with no assignee', () => {
    const tasks = [makeTask({ assigneeId: '' })];
    expect(taskCountByMember(tasks).size).toBe(0);
  });
});

describe('membersByWorkload', () => {
  it('sorts members by assigned task count descending', () => {
    const tasks = [
      makeTask({ assigneeId: 'm1' }),
      makeTask({ assigneeId: 'm2' }),
      makeTask({ assigneeId: 'm2' }),
    ];
    const members = [makeMember({ id: 'm1' }), makeMember({ id: 'm2', name: 'Bob' })];
    expect(membersByWorkload(tasks, members)[0].id).toBe('m2');
  });
});

describe('projectsByCompletion', () => {
  it('sorts projects by completion percentage descending', () => {
    const projects = [
      makeProject({ id: 'a', taskCount: 10, completedCount: 2 }),
      makeProject({ id: 'b', taskCount: 10, completedCount: 8 }),
    ];
    expect(projectsByCompletion(projects)[0].id).toBe('b');
  });
});

describe('overdueCount', () => {
  it('counts tasks past due date', () => {
    const tasks = [
      makeTask({ dueDate: '2026-01-01', status: 'todo' }),
      makeTask({ dueDate: '2026-01-01', status: 'done' }),
      makeTask({ dueDate: '2026-12-31', status: 'todo' }),
    ];
    expect(overdueCount(tasks, '2026-06-01')).toBe(1);
  });
});

describe('estimatedTaskPercent', () => {
  it('returns percentage of tasks with estimate', () => {
    const tasks = [makeTask({ estimate: 2 }), makeTask({ estimate: undefined })];
    expect(estimatedTaskPercent(tasks)).toBe(50);
  });

  it('returns 0 for empty list', () => {
    expect(estimatedTaskPercent([])).toBe(0);
  });
});
