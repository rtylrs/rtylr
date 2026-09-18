import { describe, it, expect } from 'vitest';
import {
  isBusinessHour,
  partOfDay,
  formatMmSs,
  secondsToMinutes,
  isToday,
} from './timeUtils';

describe('isBusinessHour', () => {
  it('returns true when hour is within range', () => {
    expect(isBusinessHour(10, 9, 22)).toBe(true);
  });

  it('returns false when hour is before open', () => {
    expect(isBusinessHour(7, 9, 22)).toBe(false);
  });

  it('returns false when hour equals close', () => {
    expect(isBusinessHour(22, 9, 22)).toBe(false);
  });

  it('handles overnight hours correctly', () => {
    // Open 22, close 2: hours 22, 23, 0, 1 are open
    expect(isBusinessHour(23, 22, 2)).toBe(true);
    expect(isBusinessHour(0, 22, 2)).toBe(true);
    expect(isBusinessHour(3, 22, 2)).toBe(false);
  });

  it('returns false for non-finite inputs', () => {
    expect(isBusinessHour(NaN, 9, 22)).toBe(false);
  });
});

describe('partOfDay', () => {
  it('returns "morning" for hours 5-11', () => {
    expect(partOfDay(8)).toBe('morning');
    expect(partOfDay(5)).toBe('morning');
    expect(partOfDay(11)).toBe('morning');
  });

  it('returns "afternoon" for hours 12-16', () => {
    expect(partOfDay(14)).toBe('afternoon');
  });

  it('returns "evening" for hours 17-20', () => {
    expect(partOfDay(19)).toBe('evening');
  });

  it('returns "night" for hours 21-4', () => {
    expect(partOfDay(22)).toBe('night');
    expect(partOfDay(3)).toBe('night');
  });

  it('returns "night" for NaN', () => {
    expect(partOfDay(NaN)).toBe('night');
  });
});

describe('formatMmSs', () => {
  it('formats 0 seconds as "00:00"', () => {
    expect(formatMmSs(0)).toBe('00:00');
  });

  it('formats 65 seconds as "01:05"', () => {
    expect(formatMmSs(65)).toBe('01:05');
  });

  it('formats 3600 seconds as "60:00"', () => {
    expect(formatMmSs(3600)).toBe('60:00');
  });

  it('returns "00:00" for negative values', () => {
    expect(formatMmSs(-1)).toBe('00:00');
  });

  it('returns "00:00" for NaN', () => {
    expect(formatMmSs(NaN)).toBe('00:00');
  });
});

describe('secondsToMinutes', () => {
  it('returns the number of full minutes', () => {
    expect(secondsToMinutes(90)).toBe(1);
    expect(secondsToMinutes(120)).toBe(2);
  });

  it('returns 0 for NaN', () => {
    expect(secondsToMinutes(NaN)).toBe(0);
  });
});

describe('isToday', () => {
  it('returns true for the current date', () => {
    const now = new Date().toISOString();
    expect(isToday(now)).toBe(true);
  });

  it('returns false for a past date', () => {
    expect(isToday('2020-01-01T00:00:00Z')).toBe(false);
  });

  it('returns false for an invalid date string', () => {
    expect(isToday('not-a-date')).toBe(false);
  });
});
