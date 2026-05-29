import { describe, it, expect } from 'vitest';
import {
  leaveStatusColor,
  workingDaysBetween,
  isApprovable,
  remainingBalance,
  formatLeaveDuration,
  hasOverlap,
  monthlyAccrualRate,
  includesWeekend,
} from './leaveUtils';

describe('leaveStatusColor', () => {
  it('returns yellow for Pending', () => {
    expect(leaveStatusColor('Pending')).toBe('yellow');
  });

  it('returns green for Approved', () => {
    expect(leaveStatusColor('Approved')).toBe('green');
  });

  it('returns red for Rejected', () => {
    expect(leaveStatusColor('Rejected')).toBe('red');
  });
});

describe('workingDaysBetween', () => {
  it('returns 1 for a single weekday', () => {
    expect(workingDaysBetween('2026-05-25', '2026-05-25')).toBe(1);
  });

  it('returns 5 for Mon-Fri', () => {
    expect(workingDaysBetween('2026-05-25', '2026-05-29')).toBe(5);
  });

  it('returns 5 for Mon-Sun (skips weekend)', () => {
    expect(workingDaysBetween('2026-05-25', '2026-05-31')).toBe(5);
  });

  it('returns 0 for a Saturday-only range', () => {
    expect(workingDaysBetween('2026-05-30', '2026-05-30')).toBe(0);
  });

  it('handles two-week ranges correctly', () => {
    expect(workingDaysBetween('2026-05-25', '2026-06-05')).toBe(10);
  });
});

describe('isApprovable', () => {
  it('returns true for Pending', () => {
    expect(isApprovable('Pending')).toBe(true);
  });

  it('returns false for Approved', () => {
    expect(isApprovable('Approved')).toBe(false);
  });

  it('returns false for Rejected', () => {
    expect(isApprovable('Rejected')).toBe(false);
  });
});

describe('remainingBalance', () => {
  it('subtracts used and pending from total', () => {
    expect(remainingBalance(20, 5, 3)).toBe(12);
  });

  it('returns 0 when overdrawn', () => {
    expect(remainingBalance(10, 8, 5)).toBe(0);
  });

  it('returns total when nothing used', () => {
    expect(remainingBalance(15, 0, 0)).toBe(15);
  });

  it('returns 0 for all zeros', () => {
    expect(remainingBalance(0, 0, 0)).toBe(0);
  });
});

describe('formatLeaveDuration', () => {
  it('returns "0 days" for zero', () => {
    expect(formatLeaveDuration(0)).toBe('0 days');
  });

  it('returns singular for 1 day', () => {
    expect(formatLeaveDuration(1)).toBe('1 day');
  });

  it('returns plural for more than 1', () => {
    expect(formatLeaveDuration(5)).toBe('5 days');
  });

  it('returns "0 days" for negative input', () => {
    expect(formatLeaveDuration(-2)).toBe('0 days');
  });
});

describe('hasOverlap', () => {
  it('returns true when ranges are identical', () => {
    expect(hasOverlap('2026-06-01', '2026-06-05', '2026-06-01', '2026-06-05')).toBe(true);
  });

  it('returns true when new range starts inside existing', () => {
    expect(hasOverlap('2026-06-03', '2026-06-08', '2026-06-01', '2026-06-05')).toBe(true);
  });

  it('returns true when new range ends inside existing', () => {
    expect(hasOverlap('2026-05-29', '2026-06-02', '2026-06-01', '2026-06-05')).toBe(true);
  });

  it('returns false when ranges are adjacent (no overlap)', () => {
    expect(hasOverlap('2026-06-06', '2026-06-10', '2026-06-01', '2026-06-05')).toBe(false);
  });

  it('returns false when new range is entirely before existing', () => {
    expect(hasOverlap('2026-05-01', '2026-05-15', '2026-06-01', '2026-06-05')).toBe(false);
  });

  it('returns true when new range completely contains existing', () => {
    expect(hasOverlap('2026-05-28', '2026-06-10', '2026-06-01', '2026-06-05')).toBe(true);
  });
});

describe('monthlyAccrualRate', () => {
  it('calculates monthly accrual from annual allowance', () => {
    expect(monthlyAccrualRate(12)).toBe(1);
  });

  it('rounds to two decimal places', () => {
    expect(monthlyAccrualRate(15)).toBe(1.25);
  });
});

describe('includesWeekend', () => {
  it('returns true when range spans a Saturday', () => {
    expect(includesWeekend('2025-01-11', '2025-01-12')).toBe(true);
  });

  it('returns false for a weekday-only range', () => {
    expect(includesWeekend('2025-01-06', '2025-01-10')).toBe(false);
  });

  it('returns true for a range including Sunday', () => {
    expect(includesWeekend('2025-01-12', '2025-01-13')).toBe(true);
  });
});
