import { describe, it, expect } from 'vitest';
import {
  onboardingTaskStatusLabel,
  completedTasks,
  pendingTasks,
  onboardingProgress,
  isOnboardingComplete,
  tasksDueToday,
} from './onboarding';
import type { OnboardingTask } from './onboarding';

function makeTask(
  id: string,
  status: 'pending' | 'in_progress' | 'completed' | 'skipped',
  dueDate?: string
): OnboardingTask {
  return { id, title: `Task ${id}`, status, dueDate };
}

const tasks = [
  makeTask('1', 'completed'),
  makeTask('2', 'in_progress'),
  makeTask('3', 'pending'),
  makeTask('4', 'skipped'),
];

describe('onboardingTaskStatusLabel', () => {
  it('returns display labels for known statuses', () => {
    expect(onboardingTaskStatusLabel('in_progress')).toBe('In Progress');
    expect(onboardingTaskStatusLabel('completed')).toBe('Completed');
  });

  it('returns raw value for unknown status', () => {
    expect(onboardingTaskStatusLabel('unknown')).toBe('unknown');
  });
});

describe('completedTasks', () => {
  it('returns only completed tasks', () => {
    expect(completedTasks(tasks)).toHaveLength(1);
  });
});

describe('pendingTasks', () => {
  it('returns pending and in_progress tasks', () => {
    expect(pendingTasks(tasks)).toHaveLength(2);
  });
});

describe('onboardingProgress', () => {
  it('calculates percentage of non-skipped tasks completed', () => {
    expect(onboardingProgress(tasks)).toBeCloseTo(33.3, 0);
  });

  it('returns 0 for all skipped tasks', () => {
    expect(onboardingProgress([makeTask('1', 'skipped')])).toBe(0);
  });
});

describe('isOnboardingComplete', () => {
  it('returns false when pending tasks exist', () => {
    expect(isOnboardingComplete(tasks)).toBe(false);
  });

  it('returns true when no pending tasks', () => {
    const done = [makeTask('1', 'completed'), makeTask('2', 'skipped')];
    expect(isOnboardingComplete(done)).toBe(true);
  });
});

describe('tasksDueToday', () => {
  it('returns non-completed tasks due today', () => {
    const withDates = [
      makeTask('1', 'pending', '2026-05-29'),
      makeTask('2', 'completed', '2026-05-29'),
      makeTask('3', 'in_progress', '2026-05-30'),
    ];
    expect(tasksDueToday(withDates, '2026-05-29')).toHaveLength(1);
  });
});
