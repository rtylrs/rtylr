import { describe, it, expect } from 'vitest';
import {
  filterByStatus,
  completionRate,
  inProgress,
  hasEnrolled,
  sortByEnrollment,
  totalEnrolled,
} from './automationUtils';
import type { Automation } from '../schemas/automations';

const makeAutomation = (overrides: Partial<Automation> = {}): Automation => ({
  id: 1,
  name: 'Birthday',
  description: 'Birthday reward',
  trigger: 'birthday',
  status: 'active',
  enrolled: 100,
  completed: 60,
  createdAt: '2024-01-01',
  ...overrides,
});

const automations: Automation[] = [
  makeAutomation({ id: 1, status: 'active', enrolled: 150, completed: 90 }),
  makeAutomation({ id: 2, status: 'paused', enrolled: 50, completed: 50 }),
  makeAutomation({ id: 3, status: 'active', enrolled: 200, completed: 120 }),
  makeAutomation({ id: 4, status: 'draft', enrolled: 0, completed: 0 }),
];

describe('filterByStatus', () => {
  it('returns only active automations', () => {
    const result = filterByStatus(automations, 'active');
    expect(result).toHaveLength(2);
    expect(result.every((a) => a.status === 'active')).toBe(true);
  });

  it('returns empty array for unmatched status', () => {
    expect(filterByStatus(automations, 'completed')).toHaveLength(0);
  });
});

describe('completionRate', () => {
  it('returns the correct fraction', () => {
    expect(completionRate(makeAutomation({ enrolled: 100, completed: 75 }))).toBeCloseTo(0.75);
  });

  it('returns 0 when enrolled is 0', () => {
    expect(completionRate(makeAutomation({ enrolled: 0, completed: 0 }))).toBe(0);
  });

  it('returns 1 when all enrolled have completed', () => {
    expect(completionRate(makeAutomation({ enrolled: 50, completed: 50 }))).toBe(1);
  });
});

describe('inProgress', () => {
  it('returns enrolled minus completed', () => {
    expect(inProgress(makeAutomation({ enrolled: 100, completed: 60 }))).toBe(40);
  });

  it('returns 0 when all completed', () => {
    expect(inProgress(makeAutomation({ enrolled: 50, completed: 50 }))).toBe(0);
  });

  it('returns 0 when completed exceeds enrolled (edge case)', () => {
    expect(inProgress(makeAutomation({ enrolled: 10, completed: 15 }))).toBe(0);
  });
});

describe('hasEnrolled', () => {
  it('returns true when enrolled > 0', () => {
    expect(hasEnrolled(makeAutomation({ enrolled: 10 }))).toBe(true);
  });

  it('returns false when enrolled is 0', () => {
    expect(hasEnrolled(makeAutomation({ enrolled: 0 }))).toBe(false);
  });
});

describe('sortByEnrollment', () => {
  it('sorts by enrolled count, highest first', () => {
    const sorted = sortByEnrollment(automations);
    expect(sorted[0].enrolled).toBe(200);
    expect(sorted[sorted.length - 1].enrolled).toBe(0);
  });

  it('does not mutate the original array', () => {
    const original = [...automations];
    sortByEnrollment(automations);
    expect(automations[0]).toEqual(original[0]);
  });
});

describe('totalEnrolled', () => {
  it('sums enrolled counts across all automations', () => {
    expect(totalEnrolled(automations)).toBe(400);
  });

  it('returns 0 for empty array', () => {
    expect(totalEnrolled([])).toBe(0);
  });
});
