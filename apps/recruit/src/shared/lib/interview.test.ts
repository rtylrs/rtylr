import { describe, it, expect } from 'vitest';
import {
  interviewTypeLabel,
  defaultDuration,
  allInterviewTypes,
  formatDuration,
  isValidMeetingUrl,
} from './interview';

describe('interviewTypeLabel', () => {
  it('returns correct label for known types', () => {
    expect(interviewTypeLabel('phone')).toBe('Phone Screen');
    expect(interviewTypeLabel('video')).toBe('Video Call');
    expect(interviewTypeLabel('onsite')).toBe('On-site');
  });

  it('returns the type itself for unknown values', () => {
    expect(interviewTypeLabel('custom')).toBe('custom');
  });
});

describe('defaultDuration', () => {
  it('returns expected defaults', () => {
    expect(defaultDuration('phone')).toBe(30);
    expect(defaultDuration('technical')).toBe(90);
  });

  it('returns 60 for unknown types', () => {
    expect(defaultDuration('other')).toBe(60);
  });
});

describe('allInterviewTypes', () => {
  it('returns all five types', () => {
    const types = allInterviewTypes();
    expect(types).toContain('phone');
    expect(types).toContain('panel');
    expect(types).toHaveLength(5);
  });
});

describe('formatDuration', () => {
  it('formats minutes-only durations', () => {
    expect(formatDuration(30)).toBe('30min');
    expect(formatDuration(45)).toBe('45min');
  });

  it('formats hour-only durations', () => {
    expect(formatDuration(60)).toBe('1h');
    expect(formatDuration(120)).toBe('2h');
  });

  it('formats mixed durations', () => {
    expect(formatDuration(90)).toBe('1h 30min');
    expect(formatDuration(75)).toBe('1h 15min');
  });
});

describe('isValidMeetingUrl', () => {
  it('returns true for valid https URLs', () => {
    expect(isValidMeetingUrl('https://meet.google.com/abc')).toBe(true);
    expect(isValidMeetingUrl('https://zoom.us/j/123')).toBe(true);
  });

  it('returns false for invalid URLs', () => {
    expect(isValidMeetingUrl('not-a-url')).toBe(false);
    expect(isValidMeetingUrl('http://insecure.com')).toBe(false);
  });
});
