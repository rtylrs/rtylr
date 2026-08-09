import { describe, it, expect } from 'vitest';
import {
  OT_MULTIPLIERS,
  calculateOvertimePay,
  overtimeCategoryLabel,
  overtimeStatusLabel,
  isMutable,
  totalOvertimeHours,
  approvedOvertimeRecords,
} from './overtimeUtils';

describe('OT_MULTIPLIERS', () => {
  it('has 1.5x for regular overtime', () => {
    expect(OT_MULTIPLIERS.regular).toBe(1.5);
  });

  it('has 2.0x for weekend overtime', () => {
    expect(OT_MULTIPLIERS.weekend).toBe(2.0);
  });

  it('has 2.5x for holiday overtime', () => {
    expect(OT_MULTIPLIERS.holiday).toBe(2.5);
  });

  it('has 1.25x for night shift overtime', () => {
    expect(OT_MULTIPLIERS.night).toBe(1.25);
  });
});

describe('calculateOvertimePay', () => {
  it('calculates regular OT pay correctly', () => {
    expect(calculateOvertimePay(2, 20, 'regular')).toBe(60);
  });

  it('calculates weekend OT pay correctly', () => {
    expect(calculateOvertimePay(3, 10, 'weekend')).toBe(60);
  });

  it('calculates holiday OT pay correctly', () => {
    expect(calculateOvertimePay(4, 10, 'holiday')).toBe(100);
  });
});

describe('overtimeCategoryLabel', () => {
  it('returns Regular OT for regular', () => {
    expect(overtimeCategoryLabel('regular')).toBe('Regular OT');
  });

  it('returns Weekend OT for weekend', () => {
    expect(overtimeCategoryLabel('weekend')).toBe('Weekend OT');
  });
});

describe('overtimeStatusLabel', () => {
  it('returns Pending for pending', () => {
    expect(overtimeStatusLabel('pending')).toBe('Pending');
  });

  it('returns Approved for approved', () => {
    expect(overtimeStatusLabel('approved')).toBe('Approved');
  });
});

describe('isMutable', () => {
  it('returns true for pending status', () => {
    expect(isMutable('pending')).toBe(true);
  });

  it('returns false for approved status', () => {
    expect(isMutable('approved')).toBe(false);
  });
});

describe('totalOvertimeHours', () => {
  it('sums hours from multiple records', () => {
    const records = [{ hours: 2 }, { hours: 3 }, { hours: 1.5 }];
    expect(totalOvertimeHours(records)).toBe(6.5);
  });

  it('returns 0 for empty array', () => {
    expect(totalOvertimeHours([])).toBe(0);
  });
});

describe('approvedOvertimeRecords', () => {
  it('filters to only approved records', () => {
    const records = [
      { id: 1, status: 'approved' as const, hours: 2 },
      { id: 2, status: 'pending' as const, hours: 3 },
      { id: 3, status: 'approved' as const, hours: 1 },
    ];
    const result = approvedOvertimeRecords(records);
    expect(result).toHaveLength(2);
  });

  it('returns empty array when no approved records', () => {
    const records = [{ id: 1, status: 'pending' as const, hours: 2 }];
    expect(approvedOvertimeRecords(records)).toHaveLength(0);
  });
});
