import { describe, expect, it } from 'vitest';
import { formatDuration, formatDurationClock, parseDurationInput } from './durationUtils';

describe('parseDurationInput', () => {
  it('reads bare minutes', () => {
    expect(parseDurationInput('90')).toBe(90);
    expect(parseDurationInput('90m')).toBe(90);
  });

  it('reads decimal hours', () => {
    expect(parseDurationInput('1.5h')).toBe(90);
    expect(parseDurationInput('2h')).toBe(120);
  });

  it('reads clock notation', () => {
    expect(parseDurationInput('1:30')).toBe(90);
    expect(parseDurationInput('0:45')).toBe(45);
    expect(parseDurationInput('12:05')).toBe(725);
  });

  it('tolerates surrounding whitespace and capitals', () => {
    expect(parseDurationInput('  1.5H ')).toBe(90);
    expect(parseDurationInput(' 45M')).toBe(45);
  });

  it('rejects input it cannot read rather than guessing', () => {
    expect(parseDurationInput('')).toBeNull();
    expect(parseDurationInput('soon')).toBeNull();
    expect(parseDurationInput('1h30')).toBeNull();
    expect(parseDurationInput('-30')).toBeNull();
  });

  it('rejects zero, because a zero-minute entry is never intentional', () => {
    expect(parseDurationInput('0')).toBeNull();
    expect(parseDurationInput('0:00')).toBeNull();
  });

  it('rejects clock notation with more than 59 minutes', () => {
    expect(parseDurationInput('1:60')).toBeNull();
    expect(parseDurationInput('1:75')).toBeNull();
  });

  it('rejects values outside the safe database duration range', () => {
    expect(parseDurationInput('9'.repeat(400))).toBeNull();
    expect(parseDurationInput('2147483648')).toBeNull();
    expect(parseDurationInput('999999999999999999999999999999999999h')).toBeNull();
  });
});

describe('formatDuration', () => {
  it('drops the empty unit', () => {
    expect(formatDuration(45)).toBe('45m');
    expect(formatDuration(120)).toBe('2h');
    expect(formatDuration(90)).toBe('1h 30m');
  });

  it('renders nothing logged as 0m rather than an empty string', () => {
    expect(formatDuration(0)).toBe('0m');
  });
});

describe('formatDurationClock', () => {
  it('pads minutes so totals line up in a column', () => {
    expect(formatDurationClock(405)).toBe('6:45');
    expect(formatDurationClock(480)).toBe('8:00');
    expect(formatDurationClock(5)).toBe('0:05');
  });

  it('does not roll hours over at 24', () => {
    expect(formatDurationClock(1500)).toBe('25:00');
  });
});
