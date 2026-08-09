import { describe, it, expect } from 'vitest';

import {
  rangesOverlap,
  clampToRange,
  spanInDays,
  milestonesInRange,
  projectsInRange,
  earliestProjectStart,
  latestProjectEnd,
} from './roadmapUtils';
import type { Milestone, Project } from '../stores/flowStore';

function makeMilestone(overrides: Partial<Milestone>): Milestone {
  return {
    id: 'm1',
    projectId: 'p1',
    name: 'Milestone',
    targetDate: '2026-06-15',
    progress: 0,
    computedProgress: 0,
    doneCount: 0,
    taskCount: 10,
    status: 'active',
    ...overrides,
  } as Milestone;
}

function makeProject(overrides: Partial<Project>): Project {
  return {
    id: 'p1',
    workspaceId: 'w1',
    name: 'Project',
    color: '#3b82f6',
    taskCount: 10,
    completedCount: 0,
    memberAvatars: [],
    health: 'on_track',
    progress: 0,
    startDate: '2026-05-01',
    targetDate: '2026-07-31',
    lead: '',
    ...overrides,
  } as Project;
}

describe('rangesOverlap', () => {
  it('returns true for overlapping ranges', () => {
    expect(
      rangesOverlap(
        { start: '2026-01-01', end: '2026-06-30' },
        { start: '2026-05-01', end: '2026-12-31' }
      )
    ).toBe(true);
  });

  it('returns false for non-overlapping ranges', () => {
    expect(
      rangesOverlap(
        { start: '2026-01-01', end: '2026-03-31' },
        { start: '2026-05-01', end: '2026-12-31' }
      )
    ).toBe(false);
  });

  it('returns true for adjacent ranges (touching)', () => {
    expect(
      rangesOverlap(
        { start: '2026-01-01', end: '2026-05-01' },
        { start: '2026-05-01', end: '2026-12-31' }
      )
    ).toBe(true);
  });
});

describe('clampToRange', () => {
  it('clamps below min to min', () => {
    expect(clampToRange('2026-01-01', '2026-03-01', '2026-12-31')).toBe('2026-03-01');
  });

  it('clamps above max to max', () => {
    expect(clampToRange('2027-01-01', '2026-01-01', '2026-12-31')).toBe('2026-12-31');
  });

  it('returns date unchanged if within range', () => {
    expect(clampToRange('2026-06-15', '2026-01-01', '2026-12-31')).toBe('2026-06-15');
  });
});

describe('spanInDays', () => {
  it('returns correct day span', () => {
    expect(spanInDays('2026-05-01', '2026-05-08')).toBe(7);
  });

  it('returns 0 for same date', () => {
    expect(spanInDays('2026-05-01', '2026-05-01')).toBe(0);
  });

  it('returns 0 when end is before start', () => {
    expect(spanInDays('2026-05-08', '2026-05-01')).toBe(0);
  });
});

describe('milestonesInRange', () => {
  it('returns milestones within range', () => {
    const milestones = [
      makeMilestone({ id: 'a', targetDate: '2026-05-15' }),
      makeMilestone({ id: 'b', targetDate: '2026-08-01' }),
    ];
    const result = milestonesInRange(milestones, { start: '2026-01-01', end: '2026-06-30' });
    expect(result.map((m) => m.id)).toEqual(['a']);
  });
});

describe('projectsInRange', () => {
  it('returns projects overlapping range', () => {
    const projects = [
      makeProject({ id: 'a', startDate: '2026-01-01', targetDate: '2026-06-30' }),
      makeProject({ id: 'b', startDate: '2026-08-01', targetDate: '2026-12-31' }),
    ];
    const result = projectsInRange(projects, { start: '2026-05-01', end: '2026-07-31' });
    expect(result.map((p) => p.id)).toEqual(['a']);
  });
});

describe('earliestProjectStart', () => {
  it('returns earliest start date', () => {
    const projects = [
      makeProject({ startDate: '2026-03-01' }),
      makeProject({ startDate: '2026-01-01' }),
      makeProject({ startDate: '2026-06-01' }),
    ];
    expect(earliestProjectStart(projects)).toBe('2026-01-01');
  });

  it('returns empty string for empty list', () => {
    expect(earliestProjectStart([])).toBe('');
  });
});

describe('latestProjectEnd', () => {
  it('returns latest target date', () => {
    const projects = [
      makeProject({ targetDate: '2026-06-30' }),
      makeProject({ targetDate: '2026-12-31' }),
      makeProject({ targetDate: '2026-09-30' }),
    ];
    expect(latestProjectEnd(projects)).toBe('2026-12-31');
  });

  it('returns empty string for empty list', () => {
    expect(latestProjectEnd([])).toBe('');
  });
});
