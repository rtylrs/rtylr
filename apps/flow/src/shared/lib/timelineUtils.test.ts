import { describe, it, expect } from 'vitest';

import {
  taskToTimelineItem,
  milestoneToTimelineItem,
  projectToTimelineItem,
  sortTimelineItems,
  itemsInRange,
  earliestStart,
  latestEnd,
} from './timelineUtils';
import type { Task, Milestone, Project } from '../stores/flowStore';

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
    dueDate: '2026-06-30',
    startDate: '2026-06-01',
    subtaskCount: 0,
    projectName: 'Project',
    createdAt: '2026-01-01',
    updatedAt: '2026-01-01',
    identifier: 'FLOW-1',
    ...overrides,
  } as Task;
}

function makeMilestone(overrides: Partial<Milestone>): Milestone {
  return {
    id: 'm1',
    projectId: 'p1',
    name: 'Beta Launch',
    targetDate: '2026-07-01',
    progress: 0,
    computedProgress: 0,
    doneCount: 0,
    taskCount: 5,
    status: 'active',
    ...overrides,
  } as Milestone;
}

function makeProject(overrides: Partial<Project>): Project {
  return {
    id: 'p1',
    workspaceId: 'w1',
    name: 'Alpha Project',
    color: '#3b82f6',
    taskCount: 10,
    completedCount: 5,
    memberAvatars: [],
    health: 'on_track',
    progress: 50,
    startDate: '2026-01-01',
    targetDate: '2026-12-31',
    lead: '',
    ...overrides,
  } as Project;
}

describe('taskToTimelineItem', () => {
  it('converts task with dates to timeline item', () => {
    const item = taskToTimelineItem(makeTask({}));
    expect(item?.type).toBe('task');
    expect(item?.startDate).toBe('2026-06-01');
    expect(item?.endDate).toBe('2026-06-30');
  });

  it('returns null when start date is missing', () => {
    expect(taskToTimelineItem(makeTask({ startDate: undefined }))).toBeNull();
  });

  it('returns null when due date is missing', () => {
    expect(taskToTimelineItem(makeTask({ dueDate: '' }))).toBeNull();
  });
});

describe('milestoneToTimelineItem', () => {
  it('creates a point item at the target date', () => {
    const item = milestoneToTimelineItem(makeMilestone({}));
    expect(item.type).toBe('milestone');
    expect(item.startDate).toBe('2026-07-01');
    expect(item.endDate).toBe('2026-07-01');
  });
});

describe('projectToTimelineItem', () => {
  it('converts project to timeline item', () => {
    const item = projectToTimelineItem(makeProject({}));
    expect(item?.type).toBe('project');
    expect(item?.color).toBe('#3b82f6');
  });

  it('returns null when start or target date missing', () => {
    expect(projectToTimelineItem(makeProject({ startDate: '' }))).toBeNull();
  });
});

describe('sortTimelineItems', () => {
  it('sorts by start date ascending', () => {
    const items = [
      {
        id: 'a',
        label: 'A',
        startDate: '2026-06-01',
        endDate: '2026-06-30',
        type: 'task' as const,
      },
      {
        id: 'b',
        label: 'B',
        startDate: '2026-01-01',
        endDate: '2026-01-31',
        type: 'task' as const,
      },
    ];
    expect(sortTimelineItems(items)[0].id).toBe('b');
  });
});

describe('itemsInRange', () => {
  it('returns items overlapping the range', () => {
    const items = [
      {
        id: 'a',
        label: 'A',
        startDate: '2026-06-01',
        endDate: '2026-06-30',
        type: 'task' as const,
      },
      {
        id: 'b',
        label: 'B',
        startDate: '2026-12-01',
        endDate: '2026-12-31',
        type: 'task' as const,
      },
    ];
    expect(itemsInRange(items, '2026-06-01', '2026-07-31').length).toBe(1);
  });
});

describe('earliestStart', () => {
  it('returns the earliest start date', () => {
    const items = [
      {
        id: 'a',
        label: 'A',
        startDate: '2026-06-01',
        endDate: '2026-06-30',
        type: 'task' as const,
      },
      {
        id: 'b',
        label: 'B',
        startDate: '2026-01-01',
        endDate: '2026-01-31',
        type: 'task' as const,
      },
    ];
    expect(earliestStart(items)).toBe('2026-01-01');
  });

  it('returns null for empty list', () => {
    expect(earliestStart([])).toBeNull();
  });
});

describe('latestEnd', () => {
  it('returns the latest end date', () => {
    const items = [
      {
        id: 'a',
        label: 'A',
        startDate: '2026-01-01',
        endDate: '2026-06-30',
        type: 'task' as const,
      },
      {
        id: 'b',
        label: 'B',
        startDate: '2026-01-01',
        endDate: '2026-12-31',
        type: 'task' as const,
      },
    ];
    expect(latestEnd(items)).toBe('2026-12-31');
  });

  it('returns null for empty list', () => {
    expect(latestEnd([])).toBeNull();
  });
});
