import { describe, it, expect } from 'vitest';
import {
  extractErrorMessage,
  extractStatusCode,
  isUnauthorised,
  isNotFound,
  isServerError,
  buildQueryString,
} from './apiUtils';

const makeError = (status?: number, message?: string) => ({
  response: { status, data: { message } },
  message: 'Network error',
});

describe('extractErrorMessage', () => {
  it('returns the response data message when present', () => {
    const err = makeError(400, 'Bad request');
    expect(extractErrorMessage(err)).toBe('Bad request');
  });

  it('falls back to the error message property', () => {
    expect(extractErrorMessage({ message: 'timeout' })).toBe('timeout');
  });

  it('uses the fallback when no message is available', () => {
    expect(extractErrorMessage({})).toBe('Something went wrong.');
  });

  it('uses a custom fallback', () => {
    expect(extractErrorMessage({}, 'Custom error')).toBe('Custom error');
  });
});

describe('extractStatusCode', () => {
  it('returns the HTTP status code', () => {
    expect(extractStatusCode(makeError(404))).toBe(404);
  });

  it('returns null when no status is present', () => {
    expect(extractStatusCode({ message: 'oops' })).toBeNull();
  });
});

describe('isUnauthorised', () => {
  it('returns true for a 401 error', () => {
    expect(isUnauthorised(makeError(401))).toBe(true);
  });

  it('returns false for a 403 error', () => {
    expect(isUnauthorised(makeError(403))).toBe(false);
  });
});

describe('isNotFound', () => {
  it('returns true for a 404 error', () => {
    expect(isNotFound(makeError(404))).toBe(true);
  });

  it('returns false for a 200 status', () => {
    expect(isNotFound(makeError(200))).toBe(false);
  });
});

describe('isServerError', () => {
  it('returns true for a 500 error', () => {
    expect(isServerError(makeError(500))).toBe(true);
  });

  it('returns true for a 503 error', () => {
    expect(isServerError(makeError(503))).toBe(true);
  });

  it('returns false for a 400 error', () => {
    expect(isServerError(makeError(400))).toBe(false);
  });
});

describe('buildQueryString', () => {
  it('builds a query string from params', () => {
    expect(buildQueryString({ page: 1, limit: 20 })).toBe('?page=1&limit=20');
  });

  it('skips null and undefined values', () => {
    expect(buildQueryString({ a: 'x', b: null, c: undefined })).toBe('?a=x');
  });

  it('skips empty string values', () => {
    expect(buildQueryString({ q: '' })).toBe('');
  });

  it('returns an empty string for empty params', () => {
    expect(buildQueryString({})).toBe('');
  });

  it('encodes special characters', () => {
    const qs = buildQueryString({ q: 'hello world' });
    expect(qs).toBe('?q=hello%20world');
  });
});
