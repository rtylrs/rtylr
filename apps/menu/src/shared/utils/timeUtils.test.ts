import { describe, it, expect } from 'vitest';
import { isWithinHours, formatHour, formatDuration, waitLabel } from './timeUtils';

describe('isWithinHours', () => {
  it('returns true when hour is within the range', () => {
    expect(isWithinHours(12, { open: 9, close: 17 })).toBe(true);
  });

  it('returns true at the open hour (inclusive)', () => {
    expect(isWithinHours(9, { open: 9, close: 17 })).toBe(true);
  });

  it('returns false at the close hour (exclusive)', () => {
    expect(isWithinHours(17, { open: 9, close: 17 })).toBe(false);
  });

  it('returns false when hour is before open', () => {
    expect(isWithinHours(8, { open: 9, close: 17 })).toBe(false);
  });

  it('returns false when hour is after close', () => {
    expect(isWithinHours(22, { open: 9, close: 17 })).toBe(false);
  });
});

describe('formatHour', () => {
  it('formats midnight as 12:00 AM', () => {
    expect(formatHour(0)).toBe('12:00 AM');
  });

  it('formats noon as 12:00 PM', () => {
    expect(formatHour(12)).toBe('12:00 PM');
  });

  it('formats 1 PM as 1:00 PM', () => {
    expect(formatHour(13)).toBe('1:00 PM');
  });

  it('formats 9 AM as 9:00 AM', () => {
    expect(formatHour(9)).toBe('9:00 AM');
  });
});

describe('formatDuration', () => {
  it('formats minutes under an hour', () => {
    expect(formatDuration(45)).toBe('45m');
  });

  it('formats exactly one hour', () => {
    expect(formatDuration(60)).toBe('1h');
  });

  it('formats hours and remaining minutes', () => {
    expect(formatDuration(90)).toBe('1h 30m');
  });

  it('formats multiple hours with no minutes', () => {
    expect(formatDuration(120)).toBe('2h');
  });
});

describe('waitLabel', () => {
  it('returns "Ready now" for 0 minutes', () => {
    expect(waitLabel(0)).toBe('Ready now');
  });

  it('returns minutes for under an hour', () => {
    expect(waitLabel(30)).toBe('30 min');
  });

  it('returns hour estimate for 60+ minutes', () => {
    expect(waitLabel(90)).toBe('About 2h');
  });

  it('returns "Ready now" for negative values', () => {
    expect(waitLabel(-5)).toBe('Ready now');
  });
});
