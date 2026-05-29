import { describe, it, expect } from 'vitest';
import {
  formatDuration,
  relativeTime,
  startOfDay,
  endOfDay,
  isSameDay,
  addDays,
  daysBetween,
} from './timeUtils';

describe('formatDuration', () => {
  it('formats hours, minutes, and seconds', () => expect(formatDuration(3690)).toBe('1h 1m 30s'));
  it('omits zero parts', () => expect(formatDuration(3600)).toBe('1h'));
  it('minutes only', () => expect(formatDuration(90)).toBe('1m 30s'));
  it('seconds only', () => expect(formatDuration(45)).toBe('45s'));
  it('zero returns 0s', () => expect(formatDuration(0)).toBe('0s'));
});

describe('relativeTime', () => {
  const now = new Date('2024-01-15T12:00:00Z').getTime();

  it('just now within 30 seconds', () => {
    expect(relativeTime(now - 10_000, now)).toBe('just now');
  });
  it('minutes ago', () => {
    expect(relativeTime(now - 2 * 60_000, now)).toBe('2 minutes ago');
  });
  it('hours ago', () => {
    expect(relativeTime(now - 3 * 3_600_000, now)).toBe('3 hours ago');
  });
  it('days ago', () => {
    expect(relativeTime(now - 2 * 86_400_000, now)).toBe('2 days ago');
  });
  it('in minutes', () => {
    expect(relativeTime(now + 5 * 60_000, now)).toBe('in 5 minutes');
  });
  it('in hours', () => {
    expect(relativeTime(now + 2 * 3_600_000, now)).toBe('in 2 hours');
  });
});

describe('startOfDay', () => {
  it('sets time to midnight', () => {
    const d = startOfDay(new Date('2024-03-10T15:30:00'));
    expect(d.getHours()).toBe(0);
    expect(d.getMinutes()).toBe(0);
    expect(d.getSeconds()).toBe(0);
  });
});

describe('endOfDay', () => {
  it('sets time to 23:59:59.999', () => {
    const d = endOfDay(new Date('2024-03-10T08:00:00'));
    expect(d.getHours()).toBe(23);
    expect(d.getMinutes()).toBe(59);
    expect(d.getMilliseconds()).toBe(999);
  });
});

describe('isSameDay', () => {
  it('same day', () =>
    expect(isSameDay(new Date('2024-01-01T00:00'), new Date('2024-01-01T23:59'))).toBe(true));
  it('different day', () =>
    expect(isSameDay(new Date('2024-01-01'), new Date('2024-01-02'))).toBe(false));
});

describe('addDays', () => {
  it('adds positive days', () => {
    const result = addDays(new Date('2024-01-01'), 5);
    expect(result.getDate()).toBe(6);
  });
  it('subtracts when negative', () => {
    const result = addDays(new Date('2024-01-10'), -3);
    expect(result.getDate()).toBe(7);
  });
});

describe('daysBetween', () => {
  it('counts days', () =>
    expect(daysBetween(new Date('2024-01-01'), new Date('2024-01-11'))).toBe(10));
  it('order does not matter', () =>
    expect(daysBetween(new Date('2024-01-11'), new Date('2024-01-01'))).toBe(10));
  it('same day is 0', () =>
    expect(daysBetween(new Date('2024-06-01'), new Date('2024-06-01'))).toBe(0));
});
