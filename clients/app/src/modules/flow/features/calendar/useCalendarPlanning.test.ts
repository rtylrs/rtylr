import { describe, expect, it } from 'vitest';
import type { ApiCycle, ApiMilestone } from '../../shared/services/flowApi';
import { leaveByDay, toPlanningCycles, toPlanningMilestones } from './useCalendarPlanning';

const cycle = (overrides: Partial<ApiCycle>): ApiCycle => ({
  id: '1',
  name: 'Sprint 1',
  startDate: '2026-11-02',
  endDate: '2026-11-13',
  status: 'active',
  scope: 0,
  completed: 0,
  ...overrides,
});

const milestone = (overrides: Partial<ApiMilestone>): ApiMilestone => ({
  id: '7',
  projectId: '8',
  name: 'Beta',
  targetDate: '2026-11-20',
  progress: 0,
  computedProgress: 0,
  doneCount: 0,
  taskCount: 0,
  status: 'active',
  ...overrides,
});

describe('toPlanningCycles', () => {
  it('keeps the project and the day range of each cycle', () => {
    expect(toPlanningCycles('8', [cycle({ endDate: '2026-11-13T00:00:00Z' })])).toEqual([
      {
        id: '1',
        projectId: '8',
        name: 'Sprint 1',
        range: { start: '2026-11-02', end: '2026-11-13' },
        status: 'active',
      },
    ]);
  });

  it('drops a cycle without both dates and straightens a reversed one', () => {
    const out = toPlanningCycles('8', [
      cycle({ id: '1', endDate: '' }),
      cycle({ id: '2', startDate: '2026-11-13', endDate: '2026-11-02' }),
    ]);
    expect(out.map((c) => c.id)).toEqual(['2']);
    expect(out[0].range).toEqual({ start: '2026-11-02', end: '2026-11-13' });
  });
});

describe('toPlanningMilestones', () => {
  it('places a milestone on its target day and skips one without', () => {
    expect(toPlanningMilestones([milestone({}), milestone({ id: '9', targetDate: '' })])).toEqual([
      { id: '7', name: 'Beta', day: '2026-11-20', status: 'active' },
    ]);
  });
});

describe('leaveByDay', () => {
  it('counts distinct people on leave per day inside the window', () => {
    const days = leaveByDay(
      [
        { employeeId: '1', startDate: '2026-11-18', endDate: '2026-11-20' },
        { employeeId: '2', startDate: '2026-11-20', endDate: '2026-11-25' },
        // The same person twice on one day is still one person.
        { employeeId: '1', startDate: '2026-11-20', endDate: '2026-11-20' },
      ],
      '2026-11-19',
      '2026-11-21'
    );
    expect(Object.fromEntries(days)).toEqual({
      '2026-11-19': { people: 1 },
      '2026-11-20': { people: 2 },
      '2026-11-21': { people: 1 },
    });
  });

  it('treats a request with no end as a single day and ignores one with no start', () => {
    const days = leaveByDay(
      [
        { employeeId: '1', startDate: '2026-11-19' },
        { employeeId: '2', endDate: '2026-11-19' },
      ],
      '2026-11-01',
      '2026-11-30'
    );
    expect(Object.fromEntries(days)).toEqual({ '2026-11-19': { people: 1 } });
  });
});
