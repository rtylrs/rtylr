import { describe, it, expect } from 'vitest';
import { formatElapsed, elapsedSeconds } from './useKdsTimer';

describe('formatElapsed', () => {
  it('formats zero seconds as 00:00 with ok urgency', () => {
    expect(formatElapsed(0)).toEqual({ label: '00:00', urgency: 'ok' });
  });

  it('formats 59 seconds as 00:59 with ok urgency', () => {
    expect(formatElapsed(59)).toEqual({ label: '00:59', urgency: 'ok' });
  });

  it('formats 60 seconds as 01:00 with ok urgency', () => {
    expect(formatElapsed(60)).toEqual({ label: '01:00', urgency: 'ok' });
  });

  it('formats 299 seconds as 04:59 with ok urgency', () => {
    expect(formatElapsed(299)).toEqual({ label: '04:59', urgency: 'ok' });
  });

  it('formats 300 seconds as 05:00 with warn urgency', () => {
    expect(formatElapsed(300)).toEqual({ label: '05:00', urgency: 'warn' });
  });

  it('formats 599 seconds as 09:59 with warn urgency', () => {
    expect(formatElapsed(599)).toEqual({ label: '09:59', urgency: 'warn' });
  });

  it('formats 600 seconds as 10:00 with crit urgency', () => {
    expect(formatElapsed(600)).toEqual({ label: '10:00', urgency: 'crit' });
  });

  it('formats 3600 seconds (1 hour) as 60:00 with crit urgency', () => {
    expect(formatElapsed(3600)).toEqual({ label: '60:00', urgency: 'crit' });
  });

  it('zero-pads single-digit minutes', () => {
    const { label } = formatElapsed(65);
    expect(label).toBe('01:05');
  });

  it('zero-pads single-digit seconds', () => {
    const { label } = formatElapsed(121);
    expect(label).toBe('02:01');
  });
});

describe('elapsedSeconds', () => {
  it('returns 0 for a future timestamp', () => {
    const future = new Date(Date.now() + 10_000).toISOString();
    expect(elapsedSeconds(future)).toBe(0);
  });

  it('returns a positive integer for a past timestamp', () => {
    const past = new Date(Date.now() - 5_000).toISOString();
    const result = elapsedSeconds(past);
    expect(result).toBeGreaterThanOrEqual(4);
    expect(result).toBeLessThanOrEqual(6);
  });

  it('returns a whole number (floors sub-second precision)', () => {
    const past = new Date(Date.now() - 1_500).toISOString();
    const result = elapsedSeconds(past);
    expect(Number.isInteger(result)).toBe(true);
  });
});
