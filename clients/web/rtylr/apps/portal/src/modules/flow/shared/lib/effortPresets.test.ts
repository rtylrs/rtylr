import { describe, it, expect } from 'vitest';

import { formatEffortMinutes, EFFORT_PRESETS } from './effortPresets';

describe('formatEffortMinutes', () => {
  it('returns "-" for undefined', () => {
    expect(formatEffortMinutes(undefined)).toBe('-');
  });

  it('returns "-" for 0', () => {
    expect(formatEffortMinutes(0)).toBe('-');
  });

  it('returns "-" for negative values', () => {
    expect(formatEffortMinutes(-10)).toBe('-');
  });

  it('formats minutes under 60 as "Xm"', () => {
    expect(formatEffortMinutes(15)).toBe('15m');
    expect(formatEffortMinutes(45)).toBe('45m');
    expect(formatEffortMinutes(59)).toBe('59m');
  });

  it('formats exactly 60 minutes as "1h"', () => {
    expect(formatEffortMinutes(60)).toBe('1h');
  });

  it('formats whole-hour values without decimal', () => {
    expect(formatEffortMinutes(120)).toBe('2h');
    expect(formatEffortMinutes(240)).toBe('4h');
  });

  it('formats fractional hours with one decimal', () => {
    expect(formatEffortMinutes(90)).toBe('1.5h');
    expect(formatEffortMinutes(150)).toBe('2.5h');
  });
});

describe('EFFORT_PRESETS', () => {
  it('contains the expected values in order', () => {
    expect([...EFFORT_PRESETS]).toEqual([15, 30, 60, 120, 240]);
  });
});
