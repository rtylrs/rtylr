import { describe, it, expect } from 'vitest';

import {
  sprintVelocities,
  averageVelocity,
  scopeVariance,
  predictability,
  carryoverTasks,
  carryoverRate,
  sortCyclesDesc,
} from './sprintUtils';
import type { Cycle, Task } from '../stores/flowStore';

function makeCycle(overrides: Partial<Cycle>): Cycle {
  return {
    id: 'c1',
    workspaceId: 'w1',
    name: 'Sprint 1',
    startDate: '2026-01-01',
    endDate: '2026-01-14',
    status: 'completed',
    scope: 10,
    completed: 8,
    ...overrides,
  } as Cycle;
}

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
    cycleId: 'c1',
    ...overrides,
  } as Task;
}

describe('sprintVelocities', () => {
  it('maps cycles to velocity data', () => {
    const cycles = [makeCycle({ id: 'c1', name: 'Sprint 1', scope: 10, completed: 8 })];
    const velocities = sprintVelocities(cycles);
    expect(velocities[0].cycleId).toBe('c1');
    expect(velocities[0].completed).toBe(8);
  });
});

describe('averageVelocity', () => {
  it('computes average of completed cycles', () => {
    const cycles = [
      makeCycle({ status: 'completed', completed: 10 }),
      makeCycle({ id: 'c2', status: 'completed', completed: 6 }),
      makeCycle({ id: 'c3', status: 'active', completed: 3 }),
    ];
    expect(averageVelocity(cycles)).toBe(8);
  });

  it('returns 0 for no completed cycles', () => {
    expect(averageVelocity([makeCycle({ status: 'active' })])).toBe(0);
  });
});

describe('scopeVariance', () => {
  it('returns positive when over-delivered', () => {
    expect(scopeVariance(makeCycle({ scope: 8, completed: 10 }))).toBe(2);
  });

  it('returns negative when under-delivered', () => {
    expect(scopeVariance(makeCycle({ scope: 10, completed: 7 }))).toBe(-3);
  });
});

describe('predictability', () => {
  it('returns 1 for exact completion', () => {
    expect(predictability(makeCycle({ scope: 10, completed: 10 }))).toBe(1);
  });

  it('returns less than 1 for partial completion', () => {
    expect(predictability(makeCycle({ scope: 10, completed: 8 }))).toBe(0.8);
  });
});

describe('carryoverTasks', () => {
  it('returns incomplete tasks in cycle', () => {
    const tasks = [
      makeTask({ status: 'todo', cycleId: 'c1' }),
      makeTask({ id: 't2', status: 'done', cycleId: 'c1' }),
      makeTask({ id: 't3', status: 'todo', cycleId: 'c2' }),
    ];
    expect(carryoverTasks(tasks, 'c1').length).toBe(1);
  });
});

describe('carryoverRate', () => {
  it('returns ratio of carried tasks to scope', () => {
    const tasks = [
      makeTask({ status: 'todo', cycleId: 'c1' }),
      makeTask({ id: 't2', status: 'done', cycleId: 'c1' }),
    ];
    const cycle = makeCycle({ scope: 4 });
    expect(carryoverRate(tasks, cycle)).toBe(0.25);
  });

  it('returns 0 for zero scope', () => {
    expect(carryoverRate([], makeCycle({ scope: 0 }))).toBe(0);
  });
});

describe('sortCyclesDesc', () => {
  it('sorts by start date newest first', () => {
    const cycles = [
      makeCycle({ id: 'a', startDate: '2026-01-01' }),
      makeCycle({ id: 'b', startDate: '2026-06-01' }),
    ];
    expect(sortCyclesDesc(cycles)[0].id).toBe('b');
  });
});
