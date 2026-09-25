import { describe, expect, it } from 'vitest';
import { isRetryableTransportError } from './errorClassify';

const apiError = (statusCode: number, message = `unexpected status ${statusCode}`): Error =>
  Object.assign(new Error(message), { statusCode, name: 'ApiError' });

describe('isRetryableTransportError', () => {
  it('treats missing HTTP responses as retryable', () => {
    expect(isRetryableTransportError(new TypeError('fetch failed'))).toBe(true);
    expect(isRetryableTransportError(undefined)).toBe(true);
    expect(isRetryableTransportError('boom')).toBe(true);
  });

  it('treats 5xx, 408 and 429 as retryable', () => {
    expect(isRetryableTransportError(apiError(500))).toBe(true);
    expect(isRetryableTransportError(apiError(503))).toBe(true);
    expect(isRetryableTransportError(apiError(408))).toBe(true);
    expect(isRetryableTransportError(apiError(429))).toBe(true);
  });

  it('treats definitive 4xx business rejections as permanent', () => {
    expect(isRetryableTransportError(apiError(400))).toBe(false);
    expect(isRetryableTransportError(apiError(401))).toBe(false);
    expect(isRetryableTransportError(apiError(403))).toBe(false);
    expect(isRetryableTransportError(apiError(422))).toBe(false);
  });
});
