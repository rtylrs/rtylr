import { describe, it, expect } from 'vitest';
import {
  channelLabel,
  channelIcon,
  maxMessageLength,
  isWithinLimit,
  smsSegments,
  allChannels,
} from './channelUtils';

describe('channelLabel', () => {
  it('returns correct labels', () => {
    expect(channelLabel('email')).toBe('Email');
    expect(channelLabel('sms')).toBe('SMS');
    expect(channelLabel('push')).toBe('Push');
  });
});

describe('channelIcon', () => {
  it('returns icon names', () => {
    expect(channelIcon('email')).toBe('mail');
    expect(channelIcon('sms')).toBe('message-circle');
  });
});

describe('maxMessageLength', () => {
  it('returns 160 for sms', () => expect(maxMessageLength('sms')).toBe(160));
  it('returns 256 for push', () => expect(maxMessageLength('push')).toBe(256));
});

describe('isWithinLimit', () => {
  it('returns true for short message', () => expect(isWithinLimit('Hi', 'sms')).toBe(true));
  it('returns false when over limit', () => {
    expect(isWithinLimit('x'.repeat(300), 'sms')).toBe(false);
  });
});

describe('smsSegments', () => {
  it('returns 1 for short message', () => expect(smsSegments('Hello')).toBe(1));
  it('returns 2 for 161 char message', () => expect(smsSegments('x'.repeat(161))).toBe(2));
});

describe('allChannels', () => {
  it('returns all three channels', () => {
    expect(allChannels()).toContain('email');
    expect(allChannels()).toContain('sms');
    expect(allChannels()).toContain('push');
  });
});
