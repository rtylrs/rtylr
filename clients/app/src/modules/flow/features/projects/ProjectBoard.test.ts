import { describe, expect, it, afterEach, vi } from 'vitest';
import { laneKeyFor } from './ProjectBoard';
import type { Task } from '../../shared/stores/flowStore';

const task = (dueDate: string): Task =>
  ({ id: '1', projectId: 'p1', title: 'T', status: 'todo', dueDate }) as Task;

describe('board due-date swimlanes', () => {
  const todayAt = (iso: string): void => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(iso));
  };

  afterEach(() => {
    vi.useRealTimers();
  });

  // Late afternoon at UTC+3 is past 00:00 UTC, so the old
  // `Math.round((new Date(due) - new Date()) / msPerDay)` rounded a task due
  // today to -1 day and filed it under Overdue (rtylr#1077).
  it('files a task due today under Today, not Overdue', () => {
    todayAt('2026-05-02T16:00:00+03:00');

    expect(laneKeyFor(task('2026-05-02'), 'due')).toBe('Today');
  });

  it('files a task due yesterday under Overdue', () => {
    todayAt('2026-05-02T16:00:00+03:00');

    expect(laneKeyFor(task('2026-05-01'), 'due')).toBe('Overdue');
  });

  it('keeps the week, month and later boundaries on calendar days', () => {
    todayAt('2026-05-02T16:00:00+03:00');

    expect(laneKeyFor(task('2026-05-09'), 'due')).toBe('This week');
    expect(laneKeyFor(task('2026-05-10'), 'due')).toBe('This month');
    expect(laneKeyFor(task('2026-05-31'), 'due')).toBe('This month');
    // Exactly 30 days out is still inside the month lane.
    expect(laneKeyFor(task('2026-06-01'), 'due')).toBe('This month');
    expect(laneKeyFor(task('2026-06-05'), 'due')).toBe('Later');
  });

  it('ignores a timestamp suffix and a missing due date', () => {
    todayAt('2026-05-02T16:00:00+03:00');

    expect(laneKeyFor(task('2026-05-02T18:30:00.000Z'), 'due')).toBe('Today');
    expect(laneKeyFor(task(''), 'due')).toBe('No due date');
  });
});
