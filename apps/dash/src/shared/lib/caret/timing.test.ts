import { describe, expect, it } from 'vitest';
import { formatDurationMs } from './timing';

describe('formatDurationMs', () => {
  it('uses ms for sub-second durations', () => {
    expect(formatDurationMs(250)).toBe('250ms');
  });
  it('uses seconds with one decimal for sub-minute', () => {
    expect(formatDurationMs(2300)).toBe('2.3s');
  });
  it('uses minutes/seconds for long durations', () => {
    expect(formatDurationMs(125_000)).toBe('2m 5s');
  });
  it('handles invalid input', () => {
    expect(formatDurationMs(NaN)).toBe('-');
  });
});
