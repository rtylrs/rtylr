import { describe, expect, it } from 'vitest';
import { isHrRequestStatus, normalizeHrRequestStatus } from './requestStatus';

describe('normalizeHrRequestStatus', () => {
  it('normalizes API and legacy title-cased statuses to lowercase', () => {
    expect(normalizeHrRequestStatus('pending')).toBe('pending');
    expect(normalizeHrRequestStatus('Approved')).toBe('approved');
    expect(normalizeHrRequestStatus(' rejected ')).toBe('rejected');
  });

  it('returns an empty status for missing values', () => {
    expect(normalizeHrRequestStatus(undefined)).toBe('');
    expect(normalizeHrRequestStatus(null)).toBe('');
  });
});

describe('isHrRequestStatus', () => {
  it('matches lowercase API values against UI status labels', () => {
    expect(isHrRequestStatus('pending', 'Pending')).toBe(true);
    expect(isHrRequestStatus('approved', 'Approved')).toBe(true);
    expect(isHrRequestStatus('rejected', 'Rejected')).toBe(true);
  });

  it('does not match different statuses', () => {
    expect(isHrRequestStatus('approved', 'Pending')).toBe(false);
  });
});
