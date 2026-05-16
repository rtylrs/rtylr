import { describe, expect, it } from 'vitest';
import { friendlyError } from './errorMessages';

describe('friendlyError', () => {
  it('handles null/undefined', () => {
    expect(friendlyError(null)).toBe('Something went wrong.');
  });

  it('detects offline errors', () => {
    expect(friendlyError(new Error('Failed to fetch'))).toContain('connection');
  });

  it('detects rate-limit errors', () => {
    expect(friendlyError(new Error('HTTP 429'))).toContain('Too many');
  });

  it('detects auth expiry', () => {
    expect(friendlyError(new Error('401: Unauthorized'))).toContain('expired');
  });
});
