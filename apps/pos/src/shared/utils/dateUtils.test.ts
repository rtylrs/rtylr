import { describe, it, expect } from 'vitest';
import {
  todayISO,
  isToday,
  startOfDay,
  endOfDay,
  minutesBetween,
  secondsBetween,
  formatDuration,
} from './dateUtils';

describe('todayISO', () => {
  it('returns a string in YYYY-MM-DD format', () => {
    expect(todayISO()).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });

  it('returns the current date', () => {
    const today = new Date();
    const expected = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;
    expect(todayISO()).toBe(expected);
  });
});

describe('isToday', () => {
  it('returns true for the current ISO timestamp', () => {
    expect(isToday(new Date().toISOString())).toBe(true);
  });

  it('returns false for a past date', () => {
    expect(isToday('2020-01-01T00:00:00Z')).toBe(false);
  });

  it('returns false for a future date', () => {
    expect(isToday('2099-01-01T00:00:00Z')).toBe(false);
  });
});

describe('startOfDay', () => {
  it('sets hours, minutes, seconds, ms to zero', () => {
    const d = new Date('2024-06-15T14:30:00Z');
    const result = startOfDay(d);
    expect(result.getHours()).toBe(0);
    expect(result.getMinutes()).toBe(0);
    expect(result.getSeconds()).toBe(0);
    expect(result.getMilliseconds()).toBe(0);
  });

  it('does not mutate the original date', () => {
    const d = new Date('2024-06-15T14:30:00Z');
    const original = d.getTime();
    startOfDay(d);
    expect(d.getTime()).toBe(original);
  });
});

describe('endOfDay', () => {
  it('sets hours to 23, minutes to 59, seconds to 59, ms to 999', () => {
    const d = new Date('2024-06-15T14:30:00Z');
    const result = endOfDay(d);
    expect(result.getHours()).toBe(23);
    expect(result.getMinutes()).toBe(59);
    expect(result.getSeconds()).toBe(59);
    expect(result.getMilliseconds()).toBe(999);
  });
});

describe('minutesBetween', () => {
  it('returns the number of full minutes between two dates', () => {
    const from = new Date('2024-06-15T10:00:00Z');
    const to = new Date('2024-06-15T10:05:30Z');
    expect(minutesBetween(from, to)).toBe(5);
  });

  it('returns 0 when the dates are the same', () => {
    const d = new Date('2024-06-15T10:00:00Z');
    expect(minutesBetween(d, d)).toBe(0);
  });

  it('returns a negative value when to is before from', () => {
    const from = new Date('2024-06-15T10:05:00Z');
    const to = new Date('2024-06-15T10:00:00Z');
    expect(minutesBetween(from, to)).toBe(-5);
  });
});

describe('secondsBetween', () => {
  it('returns the number of full seconds between two dates', () => {
    const from = new Date('2024-06-15T10:00:00Z');
    const to = new Date('2024-06-15T10:00:45Z');
    expect(secondsBetween(from, to)).toBe(45);
  });

  it('returns 0 for the same date', () => {
    const d = new Date('2024-06-15T10:00:00Z');
    expect(secondsBetween(d, d)).toBe(0);
  });
});

describe('formatDuration', () => {
  it('formats zero seconds', () => {
    expect(formatDuration(0)).toBe('0 s');
  });

  it('formats seconds only when under a minute', () => {
    expect(formatDuration(45)).toBe('45 s');
  });

  it('formats exactly 1 minute', () => {
    expect(formatDuration(60)).toBe('1 min');
  });

  it('formats minutes and seconds', () => {
    expect(formatDuration(90)).toBe('1 min 30 s');
  });

  it('formats exactly 5 minutes', () => {
    expect(formatDuration(300)).toBe('5 min');
  });

  it('returns "0 s" for negative input', () => {
    expect(formatDuration(-10)).toBe('0 s');
  });

  it('returns "0 s" for NaN', () => {
    expect(formatDuration(NaN)).toBe('0 s');
  });
});
