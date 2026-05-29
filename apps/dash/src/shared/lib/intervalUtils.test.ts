import { describe, it, expect } from 'vitest';
import {
  toMs,
  msToHuman,
  safePollingInterval,
  isStale,
  clampInterval,
  backoffDelay,
} from './intervalUtils';

describe('toMs', () => {
  it('converts seconds to milliseconds', () => {
    expect(toMs(5, 's')).toBe(5000);
  });

  it('converts minutes to milliseconds', () => {
    expect(toMs(2, 'min')).toBe(120_000);
  });

  it('converts hours to milliseconds', () => {
    expect(toMs(1, 'h')).toBe(3_600_000);
  });

  it('converts days to milliseconds', () => {
    expect(toMs(1, 'd')).toBe(86_400_000);
  });

  it('leaves milliseconds unchanged', () => {
    expect(toMs(500, 'ms')).toBe(500);
  });
});

describe('msToHuman', () => {
  it('formats milliseconds', () => {
    expect(msToHuman(500)).toBe('500ms');
  });

  it('formats seconds', () => {
    expect(msToHuman(5000)).toBe('5s');
  });

  it('formats minutes', () => {
    expect(msToHuman(120_000)).toBe('2min');
  });

  it('formats hours', () => {
    expect(msToHuman(7_200_000)).toBe('2h');
  });

  it('formats days', () => {
    expect(msToHuman(86_400_000)).toBe('1d');
  });
});

describe('safePollingInterval', () => {
  it('returns at least 5000ms', () => {
    expect(safePollingInterval(100)).toBe(5000);
  });

  it('rounds to the nearest 5 seconds', () => {
    expect(safePollingInterval(12_000)).toBe(10_000);
  });
});

describe('isStale', () => {
  it('returns true when timestamp is older than threshold', () => {
    const now = Date.now();
    expect(isStale(now - 10_000, 5_000, now)).toBe(true);
  });

  it('returns false when timestamp is within threshold', () => {
    const now = Date.now();
    expect(isStale(now - 1_000, 5_000, now)).toBe(false);
  });
});

describe('clampInterval', () => {
  it('clamps to minimum', () => {
    expect(clampInterval(100, 1000, 10_000)).toBe(1000);
  });

  it('clamps to maximum', () => {
    expect(clampInterval(100_000, 1000, 10_000)).toBe(10_000);
  });

  it('returns the value when within bounds', () => {
    expect(clampInterval(5000, 1000, 10_000)).toBe(5000);
  });
});

describe('backoffDelay', () => {
  it('returns baseMs for attempt 0', () => {
    expect(backoffDelay(0, 1000, 30_000)).toBe(1000);
  });

  it('doubles the delay for each attempt', () => {
    expect(backoffDelay(1, 1000, 30_000)).toBe(2000);
    expect(backoffDelay(2, 1000, 30_000)).toBe(4000);
  });

  it('caps at maxMs', () => {
    expect(backoffDelay(10, 1000, 30_000)).toBe(30_000);
  });
});
