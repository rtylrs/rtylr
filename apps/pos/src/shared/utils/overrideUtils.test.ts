import { describe, it, expect } from 'vitest';
import {
  overrideSummary,
  isValidOverrideReason,
  overridesByAction,
  sortOverridesNewestFirst,
  overrideCountByManager,
} from './overrideUtils';
import type { ManagerOverride } from '../types/pos';

const makeOverride = (overrides: Partial<ManagerOverride> = {}): ManagerOverride => ({
  id: 'ov1',
  managerId: 'mgr1',
  reason: 'Customer request',
  action: 'discount',
  createdAt: '2024-06-01T10:00:00Z',
  ...overrides,
});

describe('overrideSummary', () => {
  it('returns a human-readable string with all key fields', () => {
    const o = makeOverride({ managerId: 'M1', action: 'void', reason: 'Test reason' });
    expect(overrideSummary(o)).toBe('Manager M1 authorized "void" (Test reason)');
  });
});

describe('isValidOverrideReason', () => {
  it('returns true for a reason with 3 or more characters', () => {
    expect(isValidOverrideReason('Yes')).toBe(true);
  });

  it('returns false for a reason shorter than 3 characters', () => {
    expect(isValidOverrideReason('AB')).toBe(false);
  });

  it('returns false for a blank string', () => {
    expect(isValidOverrideReason('   ')).toBe(false);
  });
});

describe('overridesByAction', () => {
  it('groups overrides by their action type', () => {
    const overrides = [
      makeOverride({ id: 'a', action: 'discount' }),
      makeOverride({ id: 'b', action: 'void' }),
      makeOverride({ id: 'c', action: 'discount' }),
    ];
    const grouped = overridesByAction(overrides);
    expect(grouped['discount']).toHaveLength(2);
    expect(grouped['void']).toHaveLength(1);
  });

  it('returns an empty object for an empty array', () => {
    expect(overridesByAction([])).toEqual({});
  });
});

describe('sortOverridesNewestFirst', () => {
  it('sorts overrides newest-first', () => {
    const overrides = [
      makeOverride({ id: 'old', createdAt: '2024-06-01T10:00:00Z' }),
      makeOverride({ id: 'new', createdAt: '2024-06-01T12:00:00Z' }),
    ];
    const sorted = sortOverridesNewestFirst(overrides);
    expect(sorted[0].id).toBe('new');
    expect(sorted[1].id).toBe('old');
  });

  it('does not mutate the original array', () => {
    const overrides = [
      makeOverride({ id: 'old', createdAt: '2024-06-01T10:00:00Z' }),
      makeOverride({ id: 'new', createdAt: '2024-06-01T12:00:00Z' }),
    ];
    sortOverridesNewestFirst(overrides);
    expect(overrides[0].id).toBe('old');
  });
});

describe('overrideCountByManager', () => {
  it('counts overrides per manager', () => {
    const overrides = [
      makeOverride({ managerId: 'M1' }),
      makeOverride({ managerId: 'M2' }),
      makeOverride({ managerId: 'M1' }),
    ];
    const counts = overrideCountByManager(overrides);
    expect(counts['M1']).toBe(2);
    expect(counts['M2']).toBe(1);
  });

  it('returns an empty object for an empty array', () => {
    expect(overrideCountByManager([])).toEqual({});
  });
});
