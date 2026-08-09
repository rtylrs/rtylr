import { describe, it, expect } from 'vitest';
import {
  headcountGap,
  isAtCapacity,
  totalCurrentHeadcount,
  totalTargetHeadcount,
  openRoles,
  departmentsUnderTarget,
  hiringProgress,
} from './headcount';
import type { HeadcountPlan } from './headcount';

function makePlan(department: string, current: number, target: number): HeadcountPlan {
  return { department, current, target };
}

const plans = [makePlan('Engineering', 8, 10), makePlan('Sales', 5, 5), makePlan('Design', 2, 4)];

describe('headcountGap', () => {
  it('returns difference between target and current', () => {
    expect(headcountGap(plans[0])).toBe(2);
  });

  it('returns negative when over target', () => {
    expect(headcountGap(makePlan('X', 6, 5))).toBe(-1);
  });
});

describe('isAtCapacity', () => {
  it('returns true when current meets target', () => {
    expect(isAtCapacity(plans[1])).toBe(true);
  });

  it('returns false when below target', () => {
    expect(isAtCapacity(plans[0])).toBe(false);
  });
});

describe('totalCurrentHeadcount', () => {
  it('sums current headcount', () => {
    expect(totalCurrentHeadcount(plans)).toBe(15);
  });
});

describe('totalTargetHeadcount', () => {
  it('sums target headcount', () => {
    expect(totalTargetHeadcount(plans)).toBe(19);
  });
});

describe('openRoles', () => {
  it('returns total open roles', () => {
    expect(openRoles(plans)).toBe(4);
  });
});

describe('departmentsUnderTarget', () => {
  it('returns departments with open roles', () => {
    const under = departmentsUnderTarget(plans);
    expect(under).toHaveLength(2);
    expect(under.map((p) => p.department)).toContain('Engineering');
  });
});

describe('hiringProgress', () => {
  it('returns percentage of target filled', () => {
    expect(hiringProgress(plans)).toBeCloseTo(78.9, 0);
  });

  it('returns 0 for empty plans', () => {
    expect(hiringProgress([])).toBe(0);
  });
});
