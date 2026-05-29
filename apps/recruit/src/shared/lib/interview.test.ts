import { describe, it, expect } from 'vitest';
import {
  interviewTypeLabel,
  defaultDuration,
  allInterviewTypes,
  formatDuration,
  isValidMeetingUrl,
} from './interview';

describe('interviewTypeLabel', () => {
  it('returns correct labels', () => {
    expect(interviewTypeLabel('phone')).toBe('Phone Screen');
    expect(interviewTypeLabel('video')).toBe('Video Interview');
    expect(interviewTypeLabel('panel')).toBe('Panel Interview');
  });

  it('returns the type itself for unknown values', () => {
    expect(interviewTypeLabel('custom')).toBe('custom');
  });
});

describe('defaultDuration', () => {
  it('returns 30 for phone screens', () => {
    expect(defaultDuration('phone')).toBe(30);
  });

  it('returns 90 for onsite', () => {
    expect(defaultDuration('onsite')).toBe(90);
  });

  it('returns 60 for unknown types', () => {
    expect(defaultDuration('mystery')).toBe(60);
  });
});

describe('allInterviewTypes', () => {
  it('returns all 5 interview types', () => {
    expect(allInterviewTypes()).toHaveLength(5);
    expect(allInterviewTypes()).toContain('phone');
    expect(allInterviewTypes()).toContain('panel');
  });
});

describe('formatDuration', () => {
  it('formats minutes-only', () => {
    expect(formatDuration(30)).toBe('30m');
    expect(formatDuration(45)).toBe('45m');
  });

  it('formats whole hours', () => {
    expect(formatDuration(60)).toBe('1h');
    expect(formatDuration(120)).toBe('2h');
  });

  it('formats hours and minutes', () => {
    expect(formatDuration(90)).toBe('1h 30m');
    expect(formatDuration(75)).toBe('1h 15m');
  });
});

describe('isValidMeetingUrl', () => {
  it('accepts valid Zoom URLs', () => {
    expect(isValidMeetingUrl('https://zoom.us/j/123456789')).toBe(true);
    expect(isValidMeetingUrl('https://us06web.zoom.us/j/123')).toBe(true);
  });

  it('accepts valid Google Meet URLs', () => {
    expect(isValidMeetingUrl('https://meet.google.com/abc-def-ghi')).toBe(true);
  });

  it('accepts valid Teams URLs', () => {
    expect(isValidMeetingUrl('https://teams.microsoft.com/l/meetup-join/abc')).toBe(true);
  });

  it('rejects arbitrary URLs', () => {
    expect(isValidMeetingUrl('https://example.com/meeting')).toBe(false);
  });

  it('rejects empty string', () => {
    expect(isValidMeetingUrl('')).toBe(false);
  });
});
