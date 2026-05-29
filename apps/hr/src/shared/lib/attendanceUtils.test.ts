import { describe, it, expect } from 'vitest';
import {
  attendanceStatusLabel,
  attendanceStatusColor,
  isWorkedDay,
  workedDayFraction,
  attendanceRate,
  totalEffectiveDays,
  isLate,
  countAbsences,
  isInRange,
} from './attendanceUtils';

describe('attendanceStatusLabel', () => {
  it('returns Present for Present', () => {
    expect(attendanceStatusLabel('Present')).toBe('Present');
  });

  it('returns Absent for Absent', () => {
    expect(attendanceStatusLabel('Absent')).toBe('Absent');
  });

  it('returns Late for Late', () => {
    expect(attendanceStatusLabel('Late')).toBe('Late');
  });

  it('returns Half Day for Half Day', () => {
    expect(attendanceStatusLabel('Half Day')).toBe('Half Day');
  });

  it('returns Remote for Remote', () => {
    expect(attendanceStatusLabel('Remote')).toBe('Remote');
  });
});

describe('attendanceStatusColor', () => {
  it('returns green for Present', () => {
    expect(attendanceStatusColor('Present')).toBe('green');
  });

  it('returns red for Absent', () => {
    expect(attendanceStatusColor('Absent')).toBe('red');
  });

  it('returns yellow for Late', () => {
    expect(attendanceStatusColor('Late')).toBe('yellow');
  });

  it('returns orange for Half Day', () => {
    expect(attendanceStatusColor('Half Day')).toBe('orange');
  });

  it('returns blue for Remote', () => {
    expect(attendanceStatusColor('Remote')).toBe('blue');
  });
});

describe('isWorkedDay', () => {
  it('returns true for Present', () => {
    expect(isWorkedDay('Present')).toBe(true);
  });

  it('returns true for Remote', () => {
    expect(isWorkedDay('Remote')).toBe(true);
  });

  it('returns true for Late', () => {
    expect(isWorkedDay('Late')).toBe(true);
  });

  it('returns true for Half Day', () => {
    expect(isWorkedDay('Half Day')).toBe(true);
  });

  it('returns false for Absent', () => {
    expect(isWorkedDay('Absent')).toBe(false);
  });
});

describe('workedDayFraction', () => {
  it('returns 0.5 for Half Day', () => {
    expect(workedDayFraction('Half Day')).toBe(0.5);
  });

  it('returns 0 for Absent', () => {
    expect(workedDayFraction('Absent')).toBe(0);
  });

  it('returns 1 for Present', () => {
    expect(workedDayFraction('Present')).toBe(1);
  });

  it('returns 1 for Remote', () => {
    expect(workedDayFraction('Remote')).toBe(1);
  });

  it('returns 1 for Late', () => {
    expect(workedDayFraction('Late')).toBe(1);
  });
});

describe('attendanceRate', () => {
  it('returns 100 when all days are present', () => {
    expect(attendanceRate(20, 20)).toBe(100);
  });

  it('returns 0 when no days are present', () => {
    expect(attendanceRate(0, 20)).toBe(0);
  });

  it('returns 0 when totalDays is 0', () => {
    expect(attendanceRate(0, 0)).toBe(0);
  });

  it('returns 75 for 15 of 20 days', () => {
    expect(attendanceRate(15, 20)).toBe(75);
  });

  it('rounds to one decimal place', () => {
    expect(attendanceRate(2, 3)).toBeCloseTo(66.7, 0);
  });
});

describe('totalEffectiveDays', () => {
  it('returns 0 for an empty array', () => {
    expect(totalEffectiveDays([])).toBe(0);
  });

  it('counts Present days as 1', () => {
    expect(totalEffectiveDays([{ status: 'Present' }, { status: 'Present' }])).toBe(2);
  });

  it('counts Half Day as 0.5', () => {
    expect(totalEffectiveDays([{ status: 'Present' }, { status: 'Half Day' }])).toBe(1.5);
  });

  it('counts Absent as 0', () => {
    expect(totalEffectiveDays([{ status: 'Absent' }])).toBe(0);
  });

  it('sums mixed statuses correctly', () => {
    const records = [
      { status: 'Present' as const },
      { status: 'Absent' as const },
      { status: 'Half Day' as const },
      { status: 'Remote' as const },
    ];
    expect(totalEffectiveDays(records)).toBe(2.5);
  });
});

describe('isLate', () => {
  it('returns true for Late', () => {
    expect(isLate('Late')).toBe(true);
  });

  it('returns false for Present', () => {
    expect(isLate('Present')).toBe(false);
  });

  it('returns false for Absent', () => {
    expect(isLate('Absent')).toBe(false);
  });
});

describe('countAbsences', () => {
  it('returns 0 for an empty array', () => {
    expect(countAbsences([])).toBe(0);
  });

  it('counts only Absent records', () => {
    const records = [
      { status: 'Present' as const },
      { status: 'Absent' as const },
      { status: 'Absent' as const },
      { status: 'Late' as const },
    ];
    expect(countAbsences(records)).toBe(2);
  });

  it('returns 0 when no absences', () => {
    expect(countAbsences([{ status: 'Present' }, { status: 'Remote' }])).toBe(0);
  });
});

describe('isInRange', () => {
  it('returns true for a date within range', () => {
    expect(isInRange('2025-06-10', '2025-06-01', '2025-06-30')).toBe(true);
  });

  it('returns true for start boundary', () => {
    expect(isInRange('2025-06-01', '2025-06-01', '2025-06-30')).toBe(true);
  });

  it('returns true for end boundary', () => {
    expect(isInRange('2025-06-30', '2025-06-01', '2025-06-30')).toBe(true);
  });

  it('returns false for a date before range', () => {
    expect(isInRange('2025-05-31', '2025-06-01', '2025-06-30')).toBe(false);
  });

  it('returns false for a date after range', () => {
    expect(isInRange('2025-07-01', '2025-06-01', '2025-06-30')).toBe(false);
  });
});
