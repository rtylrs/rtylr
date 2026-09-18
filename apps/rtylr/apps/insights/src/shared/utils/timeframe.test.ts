import { describe, it, expect } from 'vitest';
import {
  timeframeLabel,
  allTimeframes,
  timeframeDays,
  isCustomTimeframe,
  timeframeRangeValid,
  timeframeDayCount,
} from './timeframe';

describe('timeframeLabel', () => {
  it('returns display label for known timeframes', () => {
    expect(timeframeLabel('30d')).toBe('Last 30 Days');
    expect(timeframeLabel('ytd')).toBe('Year to Date');
  });

  it('returns raw value for unknown timeframe', () => {
    expect(timeframeLabel('forever')).toBe('forever');
  });
});

describe('allTimeframes', () => {
  it('returns all standard timeframe options', () => {
    const frames = allTimeframes();
    expect(frames).toContain('7d');
    expect(frames).toContain('custom');
    expect(frames).toHaveLength(7);
  });
});

describe('timeframeDays', () => {
  it('returns correct day count for standard timeframes', () => {
    expect(timeframeDays('7d')).toBe(7);
    expect(timeframeDays('1y')).toBe(365);
  });

  it('returns null for ytd and custom', () => {
    expect(timeframeDays('ytd')).toBeNull();
    expect(timeframeDays('custom')).toBeNull();
  });
});

describe('isCustomTimeframe', () => {
  it('returns true for custom timeframe', () => {
    expect(isCustomTimeframe('custom')).toBe(true);
  });

  it('returns false for standard timeframes', () => {
    expect(isCustomTimeframe('30d')).toBe(false);
  });
});

describe('timeframeRangeValid', () => {
  it('returns true when start is before end', () => {
    expect(timeframeRangeValid({ start: '2026-01-01', end: '2026-06-01' })).toBe(true);
  });

  it('returns false when start is after end', () => {
    expect(timeframeRangeValid({ start: '2026-06-01', end: '2026-01-01' })).toBe(false);
  });
});

describe('timeframeDayCount', () => {
  it('returns number of days in range', () => {
    expect(timeframeDayCount({ start: '2026-01-01', end: '2026-01-31' })).toBe(30);
  });
});
