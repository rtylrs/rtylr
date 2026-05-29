import { describe, it, expect } from 'vitest';
import {
  timeToMinutes,
  shiftDurationMinutes,
  isOvernightShift,
  shiftsOverlap,
  shiftStatusLabel,
  isShiftActive,
  formatShiftRange,
} from './shiftUtils';

describe('timeToMinutes', () => {
  it('converts midnight to 0', () => {
    expect(timeToMinutes('00:00')).toBe(0);
  });

  it('converts 01:00 to 60', () => {
    expect(timeToMinutes('01:00')).toBe(60);
  });

  it('converts 09:30 to 570', () => {
    expect(timeToMinutes('09:30')).toBe(570);
  });

  it('converts 23:59 to 1439', () => {
    expect(timeToMinutes('23:59')).toBe(1439);
  });
});

describe('shiftDurationMinutes', () => {
  it('calculates an 8-hour shift', () => {
    expect(shiftDurationMinutes('09:00', '17:00')).toBe(480);
  });

  it('calculates a 30-minute shift', () => {
    expect(shiftDurationMinutes('12:00', '12:30')).toBe(30);
  });

  it('handles overnight shifts (wraps past midnight)', () => {
    expect(shiftDurationMinutes('22:00', '06:00')).toBe(8 * 60);
  });

  it('returns 0 for same start and end', () => {
    expect(shiftDurationMinutes('09:00', '09:00')).toBe(0);
  });
});

describe('isOvernightShift', () => {
  it('returns false for a daytime shift', () => {
    expect(isOvernightShift('09:00', '17:00')).toBe(false);
  });

  it('returns true for a midnight-crossing shift', () => {
    expect(isOvernightShift('22:00', '06:00')).toBe(true);
  });

  it('returns true for a shift ending at midnight (midnight is less than 20:00)', () => {
    expect(isOvernightShift('20:00', '00:00')).toBe(true);
  });
});

describe('shiftsOverlap', () => {
  it('returns false when shifts are back-to-back', () => {
    expect(shiftsOverlap('09:00', '12:00', '12:00', '15:00')).toBe(false);
  });

  it('returns true when shifts partially overlap', () => {
    expect(shiftsOverlap('09:00', '13:00', '12:00', '16:00')).toBe(true);
  });

  it('returns true when one shift contains another', () => {
    expect(shiftsOverlap('08:00', '18:00', '10:00', '14:00')).toBe(true);
  });

  it('returns false when shifts are non-overlapping', () => {
    expect(shiftsOverlap('09:00', '12:00', '13:00', '16:00')).toBe(false);
  });
});

describe('shiftStatusLabel', () => {
  it('returns "Scheduled" for Scheduled', () => {
    expect(shiftStatusLabel('Scheduled')).toBe('Scheduled');
  });

  it('returns "In Progress" for In Progress', () => {
    expect(shiftStatusLabel('In Progress')).toBe('In Progress');
  });

  it('returns "Completed" for Completed', () => {
    expect(shiftStatusLabel('Completed')).toBe('Completed');
  });
});

describe('isShiftActive', () => {
  it('returns true for In Progress', () => {
    expect(isShiftActive('In Progress')).toBe(true);
  });

  it('returns false for Scheduled', () => {
    expect(isShiftActive('Scheduled')).toBe(false);
  });

  it('returns false for Completed', () => {
    expect(isShiftActive('Completed')).toBe(false);
  });
});

describe('formatShiftRange', () => {
  it('formats a standard shift range', () => {
    expect(formatShiftRange('09:00', '17:00')).toBe('09:00 - 17:00');
  });

  it('formats an overnight range', () => {
    expect(formatShiftRange('22:00', '06:00')).toBe('22:00 - 06:00');
  });
});
