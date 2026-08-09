import { describe, it, expect } from 'vitest';
import {
  extractErrorMessage,
  isNetworkError,
  isServerError,
  isUnauthorizedError,
  shouldRetry,
} from './errorHandling';

describe('extractErrorMessage', () => {
  it('returns the message from an Error object', () => {
    expect(extractErrorMessage(new Error('bad request'))).toBe('bad request');
  });

  it('returns a string error directly', () => {
    expect(extractErrorMessage('timeout')).toBe('timeout');
  });

  it('reads message from a plain object', () => {
    expect(extractErrorMessage({ message: 'not found' })).toBe('not found');
  });

  it('reads error property from a plain object', () => {
    expect(extractErrorMessage({ error: 'unauthorized' })).toBe('unauthorized');
  });

  it('returns the fallback for null', () => {
    expect(extractErrorMessage(null)).toBe('Something went wrong.');
  });

  it('returns a custom fallback for unknown error types', () => {
    expect(extractErrorMessage({}, 'Custom fallback')).toBe('Custom fallback');
  });
});

describe('isNetworkError', () => {
  it('returns true for "failed to fetch" error', () => {
    expect(isNetworkError(new Error('Failed to fetch'))).toBe(true);
  });

  it('returns true for timeout errors', () => {
    expect(isNetworkError(new Error('Request timeout'))).toBe(true);
  });

  it('returns false for non-Error values', () => {
    expect(isNetworkError('string error')).toBe(false);
  });

  it('returns false for a generic Error', () => {
    expect(isNetworkError(new Error('bad input'))).toBe(false);
  });
});

describe('isServerError', () => {
  it('returns true for status 500', () => {
    expect(isServerError({ status: 500 })).toBe(true);
  });

  it('returns true for status 503', () => {
    expect(isServerError({ status: 503 })).toBe(true);
  });

  it('returns false for status 404', () => {
    expect(isServerError({ status: 404 })).toBe(false);
  });

  it('returns false for non-object errors', () => {
    expect(isServerError('error string')).toBe(false);
  });
});

describe('isUnauthorizedError', () => {
  it('returns true for status 401', () => {
    expect(isUnauthorizedError({ status: 401 })).toBe(true);
  });

  it('returns false for status 403', () => {
    expect(isUnauthorizedError({ status: 403 })).toBe(false);
  });

  it('returns false for non-object errors', () => {
    expect(isUnauthorizedError(null)).toBe(false);
  });
});

describe('shouldRetry', () => {
  it('returns true for a network error on first attempt', () => {
    expect(shouldRetry(new Error('Failed to fetch'), 0)).toBe(true);
  });

  it('returns true for a server error below max retries', () => {
    expect(shouldRetry({ status: 503 }, 1)).toBe(true);
  });

  it('returns false when max retries are reached', () => {
    expect(shouldRetry(new Error('Failed to fetch'), 3)).toBe(false);
  });

  it('returns false for a 401 error', () => {
    expect(shouldRetry({ status: 401 }, 0)).toBe(false);
  });

  it('uses custom max retries', () => {
    expect(shouldRetry(new Error('Failed to fetch'), 1, 1)).toBe(false);
  });
});
