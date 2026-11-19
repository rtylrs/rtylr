import { describe, it, expect } from 'vitest';
import type { Task } from '../../shared/stores/flowStore';
import type { ApiCycle, ApiStatus } from '../../shared/services/flowApi';
import { UNASSIGNED } from '../../shared/filters/taskFilters';
import { computeProjectSummary, type SummaryInput } from './projectSummary';

const TODAY = '2026-11-18';
// Midday UTC lands on the same calendar day in every timezone the tests run in.
const at = (day: string): string => `${day}T12:00:00Z`;

function task(id: string, overrides: Partial<Task> = {}): Task {
  return {
    id,
    projectId: 'p1',
    title: `Task ${id}`,
    status: 'todo',
    statusId: '10',
    priority: 'medium',
    assignee: 'Ahmed',
    assigneeInitials: 'AH',
    assigneeId: 'u1',
    dueDate: '',
    subtaskCount: 0,
    projectName: 'Investment Hub',
    createdAt: at('2026-10-01'),
    updatedAt: at('2026-10-01'),
    identifier: `T-${id}`,
    ...overrides,
  };
}

function status(id: string, bucket: ApiStatus['bucket'], legacyStatus: string): ApiStatus {
  return { id, bucket, legacyStatus } as ApiStatus;
}

const STATUSES = [
  status('10', 'active', 'todo'),
  status('11', 'active', 'in_progress'),
  status('12', 'done', 'done'),
  status('13', 'cancelled', 'cancelled'),
];

const DONE = { status: 'done', statusId: '12' } as const;

function summarise(overrides: Partial<SummaryInput>) {
  return computeProjectSummary({
    tasks: [],
    loggedHours: new Map(),
    statuses: STATUSES,
    cycles: [],
    today: TODAY,
    ...overrides,
  });
}

describe('computeProjectSummary', () => {
  it('counts what changed in the last 7 days and what is due in the next 7', () => {
    const s = summarise({
      tasks: [
        task('1', { createdAt: at('2026-11-17'), updatedAt: at('2026-11-17') }),
        task('2', { ...DONE, completedAt: at('2026-11-16'), updatedAt: at('2026-11-16') }),
        task('3', { dueDate: '2026-11-20' }),
        task('4', { ...DONE, dueDate: '2026-11-20' }),
        task('5', { createdAt: at('2026-11-01') }),
      ],
    });
    expect(s.recent.created.count).toBe(1);
    expect(s.recent.completed.count).toBe(1);
    expect(s.recent.updated.count).toBe(2);
    // A done task due this week is not "due soon".
    expect(s.recent.dueSoon.count).toBe(1);
    expect(s.recent.created.filters.changed).toEqual(['created_7d']);
    expect(s.recent.dueSoon.filters).toMatchObject({ due: ['next_7'], statuses: ['10', '11'] });
  });

  it('flags open overdue work and links to it through the open statuses', () => {
    const s = summarise({
      tasks: [
        task('1', { dueDate: '2026-11-17' }),
        task('2', { ...DONE, dueDate: '2026-11-17' }),
        task('3', { dueDate: TODAY }),
      ],
    });
    expect(s.attention.overdue.count).toBe(1);
    expect(s.attention.overdue.filters).toMatchObject({
      due: ['overdue'],
      statuses: ['10', '11'],
    });
  });

  it('counts unassigned open tasks in each active cycle only', () => {
    const cycles = [
      { id: 7, name: 'Sprint 7', status: 'active' },
      { id: 6, name: 'Sprint 6', status: 'completed' },
    ] as unknown as ApiCycle[];
    const s = summarise({
      cycles,
      tasks: [
        task('1', { cycleId: '7', assigneeId: '', assignee: '' }),
        task('2', { cycleId: '7', assigneeId: '', assignee: '', ...DONE }),
        task('3', { cycleId: '7' }),
        task('4', { cycleId: '6', assigneeId: '', assignee: '' }),
      ],
    });
    expect(s.attention.unassignedInCycles).toHaveLength(1);
    expect(s.attention.unassignedInCycles[0]).toMatchObject({
      cycleId: '7',
      cycleName: 'Sprint 7',
      count: 1,
    });
    expect(s.attention.unassignedInCycles[0].filters).toMatchObject({
      assignees: [UNASSIGNED],
      cycles: ['7'],
    });
  });

  it('breaks open tasks down by priority, most urgent first', () => {
    const s = summarise({
      tasks: [
        task('1', { priority: 'urgent' }),
        task('2', { priority: 'urgent', ...DONE }),
        task('3', { priority: 'low' }),
      ],
    });
    expect(s.priorities.map((p) => [p.priority, p.metric.count])).toEqual([
      ['urgent', 1],
      ['high', 0],
      ['medium', 0],
      ['low', 1],
      ['no_priority', 0],
    ]);
  });

  it('gives each person their open tasks and the hours estimated on them', () => {
    const s = summarise({
      tasks: [
        task('1', { estimate: 3 }),
        task('2', { estimate: 2.5 }),
        task('3', { estimate: 8, ...DONE }),
        task('4', { assigneeId: 'u2', assignee: 'Mohammad', estimate: 1 }),
        task('5', { assigneeId: '', assignee: '' }),
      ],
    });
    expect(s.workload.map((r) => [r.key, r.name, r.total, r.open, r.openEstimateHours])).toEqual([
      ['u1', 'Ahmed', 3, 2, 5.5],
      ['u2', 'Mohammad', 1, 1, 1],
      [UNASSIGNED, '', 1, 1, 0],
    ]);
    expect(s.workload[0].filters).toMatchObject({ assignees: ['u1'], statuses: ['10', '11'] });
  });

  it('compares estimates with the hours logged on those same tasks', () => {
    const s = summarise({
      tasks: [
        task('1', { estimate: 4 }),
        task('2', { estimate: 6, ...DONE }),
        task('3', { estimate: 5, status: 'cancelled', statusId: '13' }),
        task('4'),
      ],
      loggedHours: new Map([
        ['1', 1.5],
        ['2', 7],
        ['3', 2],
        ['4', 3],
      ]),
    });
    expect(s.hours).toEqual({
      estimated: 10,
      loggedOnEstimated: 8.5,
      openEstimated: 4,
      openUnestimated: 1,
    });
  });

  it('maps each built-in status to the project statuses behind it', () => {
    const s = summarise({
      statuses: [...STATUSES, status('14', 'active', 'todo')],
    });
    expect(s.statusFilters.todo?.statuses).toEqual(['10', '14']);
    expect(s.statusFilters.in_review).toBeUndefined();
  });

  it('falls back to built-in statuses when the project statuses have not loaded', () => {
    const s = summarise({
      statuses: [],
      tasks: [task('1', { dueDate: '2026-11-01' }), task('2', { ...DONE, dueDate: '2026-11-01' })],
    });
    expect(s.attention.overdue.count).toBe(1);
    expect(s.attention.overdue.filters.statuses).toEqual([]);
  });
});
