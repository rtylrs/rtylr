import { describe, expect, it } from 'vitest';
import { detectTimezone, timezoneOffsetLabel } from './timezone';

describe('detectTimezone', () => {
  it('returns a non-empty string', () => {
    expect(detectTimezone().length).toBeGreaterThan(0);
  });
});

describe('timezoneOffsetLabel', () => {
  it('matches the +HH:MM / -HH:MM shape', () => {
    expect(timezoneOffsetLabel()).toMatch(/^[+-]\d{2}:\d{2}$/);
  });
});
