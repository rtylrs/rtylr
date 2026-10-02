import { describe, it, expect } from 'vitest';

import {
  labelsForTask,
  tasksWithLabel,
  tasksWithAnyLabel,
  tasksWithAllLabels,
  unlabeledTasks,
  labeledTasks,
  taskCountByLabel,
  labelsByUsage,
} from './taskLabelUtils';
import type { Label, Task } from '../stores/flowStore';

function makeLabel(overrides: Partial<Label>): Label {
  return { id: 'l1', name: 'Bug', color: '#ef4444', ...overrides };
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
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Project',
    createdAt: '2026-01-01',
    updatedAt: '2026-01-01',
    identifier: 'FLOW-1',
    labelIds: [],
    ...overrides,
  } as Task;
}

describe('labelsForTask', () => {
  it('returns labels assigned to the task', () => {
    const labels = [makeLabel({ id: 'l1' }), makeLabel({ id: 'l2', name: 'Feature' })];
    const task = makeTask({ labelIds: ['l1'] });
    expect(labelsForTask(labels, task).map((l) => l.id)).toEqual(['l1']);
  });

  // ApiLabel declares id as a number while the wire sends the int64 as a
  // string, so both shapes reach this function and Array.includes does not
  // coerce between them.
  it("matches a numeric label id against the task's string ids", () => {
    const labels = [
      { id: 994343, name: 'qa-temp', color: '#f00' },
      { id: 994344, name: 'blocked', color: '#0f0' },
    ];
    const task = makeTask({ labelIds: ['994343'] });

    expect(labelsForTask(labels, task).map((l) => l.name)).toEqual(['qa-temp']);
  });

  it('matches a string label id against numeric task ids', () => {
    const labels = [makeLabel({ id: '994343', name: 'qa-temp' })];
    const task = makeTask({ labelIds: [994343] as unknown as string[] });

    expect(labelsForTask(labels, task).map((l) => l.name)).toEqual(['qa-temp']);
  });

  it('returns nothing for a task with no labels', () => {
    const labels = [makeLabel({ id: 'l1' })];

    expect(labelsForTask(labels, makeTask({ labelIds: [] }))).toEqual([]);
    expect(labelsForTask(labels, makeTask({ labelIds: undefined }))).toEqual([]);
  });

  it('does not match a partial id', () => {
    const labels = [{ id: 99, name: 'short', color: '#f00' }];
    const task = makeTask({ labelIds: ['994343'] });

    expect(labelsForTask(labels, task)).toEqual([]);
  });
});

describe('tasksWithLabel', () => {
  it('returns tasks with the label', () => {
    const tasks = [makeTask({ labelIds: ['l1'] }), makeTask({ labelIds: ['l2'] })];
    expect(tasksWithLabel(tasks, 'l1').length).toBe(1);
  });
});

describe('tasksWithAnyLabel', () => {
  it('returns tasks with any of the labels', () => {
    const tasks = [
      makeTask({ id: 'a', labelIds: ['l1'] }),
      makeTask({ id: 'b', labelIds: ['l3'] }),
    ];
    expect(tasksWithAnyLabel(tasks, ['l1', 'l2']).map((t) => t.id)).toEqual(['a']);
  });

  it('returns all tasks for empty labelIds', () => {
    const tasks = [makeTask({}), makeTask({})];
    expect(tasksWithAnyLabel(tasks, []).length).toBe(2);
  });
});

describe('tasksWithAllLabels', () => {
  it('returns tasks with all labels', () => {
    const tasks = [
      makeTask({ id: 'a', labelIds: ['l1', 'l2'] }),
      makeTask({ id: 'b', labelIds: ['l1'] }),
    ];
    expect(tasksWithAllLabels(tasks, ['l1', 'l2']).map((t) => t.id)).toEqual(['a']);
  });
});

describe('unlabeledTasks', () => {
  it('returns tasks with no labels', () => {
    const tasks = [makeTask({ labelIds: [] }), makeTask({ labelIds: ['l1'] })];
    expect(unlabeledTasks(tasks).length).toBe(1);
  });
});

describe('labeledTasks', () => {
  it('returns tasks with labels', () => {
    const tasks = [makeTask({ labelIds: [] }), makeTask({ labelIds: ['l1'] })];
    expect(labeledTasks(tasks).length).toBe(1);
  });
});

describe('taskCountByLabel', () => {
  it('counts tasks per label', () => {
    const labels = [makeLabel({ id: 'l1' }), makeLabel({ id: 'l2', name: 'Feature' })];
    const tasks = [makeTask({ labelIds: ['l1'] }), makeTask({ labelIds: ['l1', 'l2'] })];
    const counts = taskCountByLabel(tasks, labels);
    expect(counts.get('l1')).toBe(2);
    expect(counts.get('l2')).toBe(1);
  });
});

describe('labelsByUsage', () => {
  it('sorts most used labels first', () => {
    const labels = [makeLabel({ id: 'l1' }), makeLabel({ id: 'l2', name: 'Feature' })];
    const tasks = [
      makeTask({ labelIds: ['l2'] }),
      makeTask({ labelIds: ['l2'] }),
      makeTask({ labelIds: ['l1'] }),
    ];
    expect(labelsByUsage(labels, tasks)[0].id).toBe('l2');
  });
});
