import { describe, it, expect } from 'vitest';
import {
  countCompleted,
  completionPercent,
  allRequiredComplete,
  pendingRequiredTasks,
  taskStatusLabel,
  isOverdue,
} from './onboardingUtils';
import type { OnboardingTask } from './onboardingUtils';

const tasks: OnboardingTask[] = [
  { id: '1', title: 'Sign NDA', status: 'completed', required: true },
  { id: '2', title: 'Setup laptop', status: 'in_progress', required: true },
  { id: '3', title: 'Office tour', status: 'not_started', required: false },
  { id: '4', title: 'Meet team', status: 'skipped', required: false },
];

describe('countCompleted', () => {
  it('counts only completed tasks', () => {
    expect(countCompleted(tasks)).toBe(1);
  });

  it('returns 0 for empty array', () => {
    expect(countCompleted([])).toBe(0);
  });
});

describe('completionPercent', () => {
  it('calculates percent correctly', () => {
    expect(completionPercent(tasks)).toBe(25);
  });

  it('returns 0 for empty task list', () => {
    expect(completionPercent([])).toBe(0);
  });
});

describe('allRequiredComplete', () => {
  it('returns false when a required task is not completed', () => {
    expect(allRequiredComplete(tasks)).toBe(false);
  });

  it('returns true when all required tasks are completed', () => {
    const done: OnboardingTask[] = [
      { id: '1', title: 'Sign NDA', status: 'completed', required: true },
      { id: '2', title: 'Office tour', status: 'not_started', required: false },
    ];
    expect(allRequiredComplete(done)).toBe(true);
  });
});

describe('pendingRequiredTasks', () => {
  it('returns required tasks that are not completed', () => {
    const pending = pendingRequiredTasks(tasks);
    expect(pending).toHaveLength(1);
    expect(pending[0].title).toBe('Setup laptop');
  });
});

describe('taskStatusLabel', () => {
  it('returns Not Started for not_started', () => {
    expect(taskStatusLabel('not_started')).toBe('Not Started');
  });

  it('returns In Progress for in_progress', () => {
    expect(taskStatusLabel('in_progress')).toBe('In Progress');
  });

  it('returns Completed for completed', () => {
    expect(taskStatusLabel('completed')).toBe('Completed');
  });

  it('returns Skipped for skipped', () => {
    expect(taskStatusLabel('skipped')).toBe('Skipped');
  });
});

describe('isOverdue', () => {
  it('returns true for a past due date on an incomplete task', () => {
    const task: OnboardingTask = {
      id: '1',
      title: 'Sign NDA',
      status: 'not_started',
      required: true,
      dueDate: '2025-01-01',
    };
    expect(isOverdue(task, '2025-06-01')).toBe(true);
  });

  it('returns false for completed tasks', () => {
    const task: OnboardingTask = {
      id: '1',
      title: 'Sign NDA',
      status: 'completed',
      required: true,
      dueDate: '2025-01-01',
    };
    expect(isOverdue(task, '2025-06-01')).toBe(false);
  });

  it('returns false when task has no due date', () => {
    const task: OnboardingTask = {
      id: '1',
      title: 'Sign NDA',
      status: 'not_started',
      required: true,
    };
    expect(isOverdue(task, '2025-06-01')).toBe(false);
  });
});
