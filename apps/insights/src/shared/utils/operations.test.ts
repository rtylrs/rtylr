import { describe, it, expect } from 'vitest';
import {
  noShowRate,
  completionRate,
  missingShifts,
  unusedHours,
  pendingTimeOffRatio,
  timeOffApprovalRate,
} from './operations';
import type {
  ShiftMetricsResponse,
  UtilizationResponse,
  TimeOffResponse,
} from '../hooks/queries/types';

function makeShifts(overrides: Partial<ShiftMetricsResponse> = {}): ShiftMetricsResponse {
  return { scheduled: 100, completed: 90, noShows: 5, completionRatePct: 90, ...overrides };
}

function makeUtil(overrides: Partial<UtilizationResponse> = {}): UtilizationResponse {
  return { scheduledHours: 160, availableHours: 200, utilizationPct: 80, ...overrides };
}

function makeTimeOff(overrides: Partial<TimeOffResponse> = {}): TimeOffResponse {
  return { pending: 3, approved: 7, ...overrides };
}

describe('noShowRate', () => {
  it('calculates no-show percentage', () => {
    expect(noShowRate(makeShifts({ scheduled: 100, noShows: 10 }))).toBe(10);
  });

  it('returns 0 when scheduled is 0', () => {
    expect(noShowRate(makeShifts({ scheduled: 0, noShows: 0 }))).toBe(0);
  });
});

describe('completionRate', () => {
  it('returns the precomputed rate', () => {
    expect(completionRate(makeShifts({ completionRatePct: 85 }))).toBe(85);
  });
});

describe('missingShifts', () => {
  it('returns scheduled minus completed minus no-shows', () => {
    const m = makeShifts({ scheduled: 100, completed: 80, noShows: 10 });
    expect(missingShifts(m)).toBe(10);
  });
});

describe('unusedHours', () => {
  it('returns scheduled minus utilized hours', () => {
    const u = makeUtil({ scheduledHours: 160, availableHours: 200, utilizationPct: 50 });
    expect(unusedHours(u)).toBe(60);
  });
});

describe('pendingTimeOffRatio', () => {
  it('returns pending fraction', () => {
    expect(pendingTimeOffRatio(makeTimeOff({ pending: 3, approved: 7 }))).toBeCloseTo(0.3);
  });

  it('returns 0 when no requests', () => {
    expect(pendingTimeOffRatio(makeTimeOff({ pending: 0, approved: 0 }))).toBe(0);
  });
});

describe('timeOffApprovalRate', () => {
  it('returns 100 when all approved', () => {
    expect(timeOffApprovalRate(makeTimeOff({ pending: 0, approved: 10 }))).toBe(100);
  });

  it('returns 0 for no requests', () => {
    expect(timeOffApprovalRate(makeTimeOff({ pending: 0, approved: 0 }))).toBe(0);
  });
});
