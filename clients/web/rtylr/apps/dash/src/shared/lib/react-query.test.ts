import { describe, expect, it } from 'vitest';
import { shouldRetryQuery } from './react-query';

describe('shouldRetryQuery', () => {
  it.each([408, 425, 429, 500, 503])('retries transient HTTP status %s once', (statusCode) => {
    expect(shouldRetryQuery(0, { statusCode })).toBe(true);
    expect(shouldRetryQuery(1, { statusCode })).toBe(false);
  });

  it.each([400, 401, 403, 404, 409, 422])('does not retry permanent HTTP status %s', (status) => {
    expect(shouldRetryQuery(0, { response: { status } })).toBe(false);
  });

  it('retries a network failure once', () => {
    expect(shouldRetryQuery(0, new Error('Network error'))).toBe(true);
    expect(shouldRetryQuery(1, new Error('Network error'))).toBe(false);
  });
});
