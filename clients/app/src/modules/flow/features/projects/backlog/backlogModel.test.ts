import { describe, expect, it } from 'vitest';
import type { Task } from '../../../shared/stores/flowStore';
import type { ApiCycle, ApiStatus } from '../../../shared/services/flowApi';
import {
  BACKLOG_SECTION_ID,
  buildBacklogSections,
  capacityTone,
  hoursLabel,
  orderByPlanRank,
  planMove,
  sectionTotals,
} from './backlogModel';

function task(id: string, patch: Partial<Task> = {}): Task {
  return {
    id,
    projectId: 'p1',
    title: `Task ${id}`,
    status: 'todo',
    priority: 'medium',
    assignee: '',
    assigneeInitials: '',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'P',
    identifier: id,
    createdAt: `2026-11-0${id}T00:00:00Z`,
    updatedAt: '',
    ...patch,
  } as Task;
}

function cycle(id: string, status: ApiCycle['status'], startDate: string): ApiCycle {
  return { id, name: `Cycle ${id}`, status, startDate, endDate: '', scope: 0, completed: 0 };
}

const noStatuses = new Map<string, ApiStatus>();

describe('buildBacklogSections', () => {
  it('orders the active cycle, then upcoming by start date, then the backlog', () => {
    const sections = buildBacklogSections(
      [
        cycle('late', 'upcoming', '2026-12-21'),
        cycle('done', 'completed', '2026-11-01'),
        cycle('now', 'active', '2026-11-23'),
        cycle('soon', 'upcoming', '2026-12-07'),
      ],
      [],
      noStatuses
    );
    expect(sections.map((s) => s.id)).toEqual(['now', 'soon', 'late', BACKLOG_SECTION_ID]);
  });

  it('puts tasks in their cycle by plan rank and leaves finished work out of the backlog', () => {
    const sections = buildBacklogSections(
      [cycle('c1', 'active', '2026-11-23'), cycle('old', 'completed', '2026-11-01')],
      [
        task('1', { cycleId: 'c1', planRank: 2048 }),
        task('2', { cycleId: 'c1', planRank: 1024 }),
        task('3'),
        task('4', { status: 'done' }),
        task('5', { cycleId: 'old', status: 'done' }),
        task('6', { archived: true }),
        task('7', { planRank: -1024 }),
      ],
      noStatuses
    );
    expect(sections[0].tasks.map((t) => t.id)).toEqual(['2', '1']);
    // A ranked backlog task comes before one that was never placed.
    expect(sections[1].tasks.map((t) => t.id)).toEqual(['7', '3']);
  });

  it('uses a custom status bucket over the built-in status', () => {
    const statuses = new Map<string, ApiStatus>([
      ['s-done', { id: 's-done', bucket: 'done' } as ApiStatus],
    ]);
    const sections = buildBacklogSections([], [task('1', { statusId: 's-done' })], statuses);
    expect(sections[0].tasks).toHaveLength(0);
  });
});

describe('sectionTotals', () => {
  it('counts to do, in progress and done, and sums estimates', () => {
    const totals = sectionTotals(
      [
        task('1', { effortMinutes: 60 }),
        task('2', { status: 'in_review', effortMinutes: 30 }),
        task('3', { status: 'done' }),
        task('4', { status: 'cancelled', effortMinutes: 15 }),
      ],
      noStatuses
    );
    expect(totals).toEqual({ todo: 1, active: 1, done: 2, estimateMinutes: 105 });
  });
});

describe('orderByPlanRank', () => {
  it('sorts ranked tasks by rank and unranked ones oldest first', () => {
    const ordered = orderByPlanRank([task('3'), task('2', { planRank: 5 }), task('1')]);
    expect(ordered.map((t) => t.id)).toEqual(['2', '1', '3']);
  });
});

describe('planMove', () => {
  const section = {
    id: 'c1',
    tasks: [
      task('1', { planRank: 1024 }),
      task('2', { planRank: 2048 }),
      task('3', { planRank: 3072 }),
    ],
    totals: { todo: 0, active: 0, done: 0, estimateMinutes: 0 },
  };

  it('anchors before the task the drop lands on', () => {
    expect(planMove(section, '3', 1)).toEqual({
      taskId: '3',
      cycleId: 'c1',
      beforeTaskId: '2',
      optimisticRank: 1536,
    });
  });

  it('anchors after the last task when dropped at the end', () => {
    expect(planMove(section, '9', 3)).toEqual({
      taskId: '9',
      cycleId: 'c1',
      afterTaskId: '3',
      optimisticRank: 4096,
    });
  });

  it('sends no anchor into an empty section', () => {
    const empty = { ...section, id: BACKLOG_SECTION_ID, tasks: [] };
    expect(planMove(empty, '9', 0)).toEqual({ taskId: '9', cycleId: '', optimisticRank: 0 });
  });
});

describe('capacityTone', () => {
  it('turns amber at 90% and red only once the time is exceeded', () => {
    expect(capacityTone(80, 100)).toBe('ok');
    expect(capacityTone(90, 100)).toBe('warn');
    expect(capacityTone(100, 100)).toBe('warn');
    expect(capacityTone(101, 100)).toBe('over');
    expect(capacityTone(30, 0)).toBe('over');
    expect(capacityTone(0, 0)).toBe('ok');
  });

  it('labels minutes as hours with one decimal', () => {
    expect(hoursLabel(90)).toBe('1.5h');
    expect(hoursLabel(0)).toBe('0h');
  });
});
