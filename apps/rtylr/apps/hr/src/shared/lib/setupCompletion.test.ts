import { describe, expect, it } from 'vitest';
import { isHRSetupComplete } from './setupCompletion';

describe('isHRSetupComplete', () => {
  it('accepts a real completion timestamp', () => {
    expect(isHRSetupComplete({ setupCompletedAt: '1699999999' })).toBe(true);
  });

  it('rejects a config that was never finalised', () => {
    expect(isHRSetupComplete({ setupCompletedAt: undefined })).toBe(false);
    expect(isHRSetupComplete({})).toBe(false);
  });

  it('rejects a missing config so callers need no separate null branch', () => {
    expect(isHRSetupComplete(null)).toBe(false);
    expect(isHRSetupComplete(undefined)).toBe(false);
  });

  // int64 crosses the wire as a string, so these are the shapes an unset
  // timestamp can actually arrive in. "0" matters most: it is truthy in JS, so
  // a plain `!config.setupCompletedAt` check would call it complete.
  it('rejects zero and empty string wire values', () => {
    expect(isHRSetupComplete({ setupCompletedAt: '0' })).toBe(false);
    expect(isHRSetupComplete({ setupCompletedAt: '' })).toBe(false);
  });

  it('rejects a value that is not a number', () => {
    expect(isHRSetupComplete({ setupCompletedAt: 'not-a-timestamp' })).toBe(false);
  });
});
