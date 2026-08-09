import { describe, it, expect } from 'vitest';

import {
  isMilestoneOverdue,
  daysUntilMilestone,
  sortMilestonesByDate,
  filterMilestonesByStatus,
  groupMilestonesByStatus,
  milestoneProgress,
} from './milestoneUtils';
import type { Milestone } from '../stores/flowStore';

function makeMilestone(overrides: Partial<Milestone>): Milestone {
  return {
    id: 'm1',
    projectId: 'p1',
    name: 'Milestone',
    targetDate: '2026-06-01',
    progress: 0,
    computedProgress: 0,
    doneCount: 0,
    taskCount: 10,
    status: 'active',
    ...overrides,
  } as Milestone;
}

const NOW = new Date('2026-05-15T12:00:00Z');

describe('isMilestoneOverdue', () => {
  it('returns true for past target date that is not completed', () => {
    const m = makeMilestone({ targetDate: '2026-05-01', status: 'active' });
    expect(isMilestoneOverdue(m, NOW)).toBe(true);
  });

  it('returns false for future target date', () => {
    const m = makeMilestone({ targetDate: '2026-06-01', status: 'active' });
    expect(isMilestoneOverdue(m, NOW)).toBe(false);
  });

  it('returns false for completed milestones even if past target', () => {
    const m = makeMilestone({ targetDate: '2026-01-01', status: 'completed' });
    expect(isMilestoneOverdue(m, NOW)).toBe(false);
  });
});

describe('daysUntilMilestone', () => {
  it('returns positive days for future milestone', () => {
    const m = makeMilestone({ targetDate: '2026-05-22' });
    expect(daysUntilMilestone(m, NOW)).toBe(7);
  });

  it('returns negative days for overdue milestone', () => {
    const m = makeMilestone({ targetDate: '2026-05-08' });
    expect(daysUntilMilestone(m, NOW)).toBe(-7);
  });

  it('returns 0 for today', () => {
    const today = new Date(NOW);
    const m = makeMilestone({ targetDate: '2026-05-15' });
    expect(daysUntilMilestone(m, today)).toBe(0);
  });
});

describe('sortMilestonesByDate', () => {
  it('sorts soonest first', () => {
    const milestones = [
      makeMilestone({ id: 'a', targetDate: '2026-08-01' }),
      makeMilestone({ id: 'b', targetDate: '2026-06-01' }),
      makeMilestone({ id: 'c', targetDate: '2026-07-01' }),
    ];
    const sorted = sortMilestonesByDate(milestones);
    expect(sorted.map((m) => m.id)).toEqual(['b', 'c', 'a']);
  });

  it('does not mutate the original array', () => {
    const milestones = [
      makeMilestone({ targetDate: '2026-12-01' }),
      makeMilestone({ targetDate: '2026-01-01' }),
    ];
    const original = [...milestones];
    sortMilestonesByDate(milestones);
    expect(milestones).toEqual(original);
  });
});

describe('filterMilestonesByStatus', () => {
  it('returns milestones with the given status', () => {
    const milestones = [
      makeMilestone({ id: 'a', status: 'active' }),
      makeMilestone({ id: 'b', status: 'completed' }),
      makeMilestone({ id: 'c', status: 'active' }),
    ];
    expect(filterMilestonesByStatus(milestones, 'active').map((m) => m.id)).toEqual(['a', 'c']);
  });
});

describe('groupMilestonesByStatus', () => {
  it('groups milestones by status', () => {
    const milestones = [
      makeMilestone({ status: 'active' }),
      makeMilestone({ status: 'completed' }),
      makeMilestone({ status: 'active' }),
    ];
    const map = groupMilestonesByStatus(milestones);
    expect(map.get('active')).toHaveLength(2);
    expect(map.get('completed')).toHaveLength(1);
  });
});

describe('milestoneProgress', () => {
  it('uses computedProgress when available', () => {
    const m = makeMilestone({ computedProgress: 65, progress: 50 });
    expect(milestoneProgress(m)).toBe(65);
  });

  it('falls back to progress when computedProgress is 0', () => {
    const m = makeMilestone({ computedProgress: 0, progress: 50 });
    expect(milestoneProgress(m)).toBe(0);
  });

  it('clamps to 100', () => {
    const m = makeMilestone({ computedProgress: 120 });
    expect(milestoneProgress(m)).toBe(100);
  });

  it('clamps to 0', () => {
    const m = makeMilestone({ computedProgress: -10 });
    expect(milestoneProgress(m)).toBe(0);
  });
});
