import { describe, it, expect } from 'vitest';

import {
  projectCompletionPercent,
  milestoneCompletionPercent,
  projectHealthLabel,
  projectHealthColor,
  inferProjectHealth,
  aggregateProgress,
} from './progressUtils';
import type { Project, Milestone } from '../stores/flowStore';

function makeProject(overrides: Partial<Project>): Project {
  return {
    id: 'p1',
    workspaceId: 'w1',
    name: 'Project',
    color: '#000',
    taskCount: 10,
    completedCount: 0,
    memberAvatars: [],
    health: 'on_track',
    progress: 0,
    startDate: '',
    targetDate: '',
    lead: '',
    ...overrides,
  } as Project;
}

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

// ---------------------------------------------------------------------------
// projectCompletionPercent
// ---------------------------------------------------------------------------

describe('projectCompletionPercent', () => {
  it('returns 0 when taskCount is 0', () => {
    expect(projectCompletionPercent(makeProject({ taskCount: 0, completedCount: 0 }))).toBe(0);
  });

  it('returns 50 when half the tasks are done', () => {
    expect(projectCompletionPercent(makeProject({ taskCount: 10, completedCount: 5 }))).toBe(50);
  });

  it('returns 100 when all tasks are done', () => {
    expect(projectCompletionPercent(makeProject({ taskCount: 10, completedCount: 10 }))).toBe(100);
  });

  it('caps at 100', () => {
    expect(projectCompletionPercent(makeProject({ taskCount: 10, completedCount: 15 }))).toBe(100);
  });
});

// ---------------------------------------------------------------------------
// milestoneCompletionPercent
// ---------------------------------------------------------------------------

describe('milestoneCompletionPercent', () => {
  it('returns 0 for empty milestone', () => {
    expect(milestoneCompletionPercent(makeMilestone({ taskCount: 0, doneCount: 0 }))).toBe(0);
  });

  it('returns correct percent', () => {
    expect(milestoneCompletionPercent(makeMilestone({ taskCount: 4, doneCount: 3 }))).toBe(75);
  });
});

// ---------------------------------------------------------------------------
// projectHealthLabel
// ---------------------------------------------------------------------------

describe('projectHealthLabel', () => {
  it('returns "On Track" for on_track', () => {
    expect(projectHealthLabel('on_track')).toBe('On Track');
  });

  it('returns "At Risk" for at_risk', () => {
    expect(projectHealthLabel('at_risk')).toBe('At Risk');
  });

  it('returns "Off Track" for off_track', () => {
    expect(projectHealthLabel('off_track')).toBe('Off Track');
  });
});

// ---------------------------------------------------------------------------
// projectHealthColor
// ---------------------------------------------------------------------------

describe('projectHealthColor', () => {
  it('returns the approved token for on_track, not a hex', () => {
    // This asserted a hex, which pinned the hardcoded #22c55e. Flow's positive
    // tone is the teal --flow-status-approved: the token set has no green, and
    // #22c55e measured 2.28:1 as text.
    expect(projectHealthColor('on_track')).toBe('var(--flow-status-approved)');
  });

  it('returns a different color for at_risk vs on_track', () => {
    expect(projectHealthColor('at_risk')).not.toBe(projectHealthColor('on_track'));
  });
});

// ---------------------------------------------------------------------------
// inferProjectHealth
// ---------------------------------------------------------------------------

describe('inferProjectHealth', () => {
  it('returns on_track for >= 75%', () => {
    expect(inferProjectHealth(100)).toBe('on_track');
    expect(inferProjectHealth(75)).toBe('on_track');
  });

  it('returns at_risk for 40-74%', () => {
    expect(inferProjectHealth(74)).toBe('at_risk');
    expect(inferProjectHealth(40)).toBe('at_risk');
  });

  it('returns off_track for < 40%', () => {
    expect(inferProjectHealth(39)).toBe('off_track');
    expect(inferProjectHealth(0)).toBe('off_track');
  });
});

// ---------------------------------------------------------------------------
// aggregateProgress
// ---------------------------------------------------------------------------

describe('aggregateProgress', () => {
  it('returns 0 for empty list', () => {
    expect(aggregateProgress([])).toBe(0);
  });

  it('returns 0 when all projects have 0 tasks', () => {
    expect(
      aggregateProgress([
        makeProject({ taskCount: 0, completedCount: 0 }),
        makeProject({ taskCount: 0, completedCount: 0 }),
      ])
    ).toBe(0);
  });

  it('returns weighted average across projects', () => {
    const projects = [
      makeProject({ taskCount: 10, completedCount: 10 }),
      makeProject({ taskCount: 10, completedCount: 0 }),
    ];
    expect(aggregateProgress(projects)).toBe(50);
  });

  it('weights larger projects more heavily', () => {
    const projects = [
      makeProject({ taskCount: 90, completedCount: 90 }),
      makeProject({ taskCount: 10, completedCount: 0 }),
    ];
    expect(aggregateProgress(projects)).toBe(90);
  });
});
