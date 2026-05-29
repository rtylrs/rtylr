import { describe, it, expect } from 'vitest';
import {
  extractErrorMessage,
  isNetworkError,
  isClientError,
  isServerError,
  httpStatusLabel,
} from './errorUtils';

describe('extractErrorMessage', () => {
  it('extracts the message from an Error instance', () => {
    expect(extractErrorMessage(new Error('Something went wrong'))).toBe('Something went wrong');
  });

  it('returns a string error directly', () => {
    expect(extractErrorMessage('Direct error')).toBe('Direct error');
  });

  it('extracts message from an object with a message property', () => {
    expect(extractErrorMessage({ message: 'Object error' })).toBe('Object error');
  });

  it('returns the fallback for null', () => {
    expect(extractErrorMessage(null)).toBe('An unexpected error occurred');
  });

  it('returns the fallback for undefined', () => {
    expect(extractErrorMessage(undefined)).toBe('An unexpected error occurred');
  });

  it('accepts a custom fallback', () => {
    expect(extractErrorMessage(null, 'Custom fallback')).toBe('Custom fallback');
  });
});

describe('isNetworkError', () => {
  it('returns true for a network-related message', () => {
    expect(isNetworkError(new Error('Network request failed'))).toBe(true);
    expect(isNetworkError('fetch error')).toBe(true);
    expect(isNetworkError('ECONNREFUSED')).toBe(true);
  });

  it('returns false for non-network errors', () => {
    expect(isNetworkError(new Error('Validation failed'))).toBe(false);
  });
});

describe('isClientError', () => {
  it('returns true for 4xx status codes', () => {
    expect(isClientError(400)).toBe(true);
    expect(isClientError(404)).toBe(true);
    expect(isClientError(499)).toBe(true);
  });

  it('returns false for 5xx status codes', () => {
    expect(isClientError(500)).toBe(false);
  });

  it('returns false for 2xx status codes', () => {
    expect(isClientError(200)).toBe(false);
  });

  it('returns false for non-finite input', () => {
    expect(isClientError(NaN)).toBe(false);
  });
});

describe('isServerError', () => {
  it('returns true for 5xx status codes', () => {
    expect(isServerError(500)).toBe(true);
    expect(isServerError(503)).toBe(true);
  });

  it('returns false for 4xx status codes', () => {
    expect(isServerError(400)).toBe(false);
  });
});

describe('httpStatusLabel', () => {
  it('returns the correct label for known status codes', () => {
    expect(httpStatusLabel(404)).toBe('Not found');
    expect(httpStatusLabel(500)).toBe('Server error');
    expect(httpStatusLabel(401)).toBe('Unauthorized');
  });

  it('returns a generic label for unknown status codes', () => {
    expect(httpStatusLabel(418)).toBe('HTTP 418');
  });
});
