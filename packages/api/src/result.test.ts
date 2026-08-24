import { describe, it, expect, vi } from 'vitest';
import { safeFetch, unwrap, oneKitResult } from './result';
import type { ApiError } from './result';
import axios from 'axios';

// ── safeFetch ──────────────────────────────────────────────────────

describe('safeFetch', () => {
  it('wraps axios GET into ResultAsync', async () => {
    const client = axios.create();
    vi.spyOn(client, 'get').mockResolvedValue({ data: { id: 1, name: 'Test' } });

    const safe = safeFetch(client);
    const result = await safe.get('/test');

    expect(result.isOk()).toBe(true);
    if (result.isOk()) {
      expect(result.value).toEqual({ id: 1, name: 'Test' });
    }
  });

  it('wraps axios POST into ResultAsync', async () => {
    const client = axios.create();
    vi.spyOn(client, 'post').mockResolvedValue({ data: { success: true } });

    const safe = safeFetch(client);
    const result = await safe.post('/test', { name: 'New' });

    expect(result.isOk()).toBe(true);
    if (result.isOk()) {
      expect(result.value).toEqual({ success: true });
    }
  });

  it('returns ApiError on network failure', async () => {
    const client = axios.create();
    const error = new Error('Network Error');
    (error as any).isAxiosError = true;
    (error as any).response = undefined;
    vi.spyOn(client, 'get').mockRejectedValue(error);

    const safe = safeFetch(client);
    const result = await safe.get('/fail');

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.code).toBe('NETWORK_ERROR');
      expect(result.error.status).toBe(0);
      expect(result.error.message).toContain('Unable to reach');
    }
  });

  it('returns UNAUTHORIZED on 401', async () => {
    const client = axios.create();
    const error = new Error('Unauthorized');
    (error as any).isAxiosError = true;
    (error as any).response = { status: 401, data: { message: 'Token expired' } };
    vi.spyOn(client, 'get').mockRejectedValue(error);

    const safe = safeFetch(client);
    const result = await safe.get('/auth');

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.code).toBe('UNAUTHORIZED');
      expect(result.error.status).toBe(401);
      expect(result.error.message).toBe('Token expired');
    }
  });

  it('returns NOT_FOUND on 404', async () => {
    const client = axios.create();
    const error = new Error('Not Found');
    (error as any).isAxiosError = true;
    (error as any).response = { status: 404, data: {} };
    vi.spyOn(client, 'get').mockRejectedValue(error);

    const safe = safeFetch(client);
    const result = await safe.get('/missing');

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.code).toBe('NOT_FOUND');
      expect(result.error.status).toBe(404);
    }
  });

  it('returns SERVER_ERROR on 500', async () => {
    const client = axios.create();
    const error = new Error('Internal');
    (error as any).isAxiosError = true;
    (error as any).response = { status: 500, data: { error: 'DB down' } };
    vi.spyOn(client, 'get').mockRejectedValue(error);

    const safe = safeFetch(client);
    const result = await safe.get('/broken');

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.code).toBe('SERVER_ERROR');
      expect(result.error.message).toBe('DB down');
    }
  });

  it('returns TIMEOUT on ECONNABORTED', async () => {
    const client = axios.create();
    const error = new Error('timeout');
    (error as any).isAxiosError = true;
    (error as any).code = 'ECONNABORTED';
    vi.spyOn(client, 'get').mockRejectedValue(error);

    const safe = safeFetch(client);
    const result = await safe.get('/slow');

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.code).toBe('TIMEOUT');
    }
  });

  it('returns VALIDATION on 422', async () => {
    const client = axios.create();
    const error = new Error('Validation');
    (error as any).isAxiosError = true;
    (error as any).response = { status: 422, data: { message: 'Email required' } };
    vi.spyOn(client, 'post').mockRejectedValue(error);

    const safe = safeFetch(client);
    const result = await safe.post('/validate', {});

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.code).toBe('VALIDATION');
      expect(result.error.message).toBe('Email required');
    }
  });
});

// ── unwrap ──────────────────────────────────────────────────────────

describe('unwrap', () => {
  it('returns data on success', async () => {
    const client = axios.create();
    vi.spyOn(client, 'get').mockResolvedValue({ data: [1, 2, 3] });

    const safe = safeFetch(client);
    const data = await unwrap(safe.get('/list'));

    expect(data).toEqual([1, 2, 3]);
  });

  it('throws ApiError on failure', async () => {
    const client = axios.create();
    const error = new Error('fail');
    (error as any).isAxiosError = true;
    (error as any).response = { status: 500, data: {} };
    vi.spyOn(client, 'get').mockRejectedValue(error);

    const safe = safeFetch(client);

    await expect(unwrap(safe.get('/fail'))).rejects.toMatchObject({
      code: 'SERVER_ERROR',
      status: 500,
    });
  });

  it('thrown error has message property', async () => {
    const client = axios.create();
    const error = new Error('fail');
    (error as any).isAxiosError = true;
    (error as any).response = { status: 403, data: { message: 'Forbidden' } };
    vi.spyOn(client, 'get').mockRejectedValue(error);

    const safe = safeFetch(client);

    try {
      await unwrap(safe.get('/forbidden'));
      expect.fail('should have thrown');
    } catch (e) {
      const apiError = e as ApiError;
      expect(apiError.message).toBe('Forbidden');
      expect(apiError.code).toBe('FORBIDDEN');
    }
  });
});

// ── oneKitResult: client-side validation vs transport failure ──────

/**
 * rtylr#950. Generated clients validate the request and throw
 * `TypeError('invalid request: ...')` before any network call. That was falling
 * through to the generic branch, which reports `status: 0` - the value reserved
 * for "could not reach the server".
 *
 * Consequence, confirmed live: dash's `shouldRetryQuery` treats `status === 0`
 * as retryable, so a request that can never succeed was retried, and in logs a
 * programming error was indistinguishable from an outage.
 */
describe('oneKitResult error classification', () => {
  it('classifies a generated-client validation throw as CLIENT_VALIDATION', async () => {
    const result = await oneKitResult(() => {
      throw new TypeError('invalid request: body.description is required');
    });

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.code).toBe('CLIENT_VALIDATION');
      // Not 0: that is the transport-failure value other code branches on.
      expect(result.error.status).not.toBe(0);
      expect(result.error.message).toContain('body.description is required');
    }
  });

  it('keeps a real transport failure on status 0', async () => {
    const result = await oneKitResult(() => {
      throw new Error('Network request failed');
    });

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.status).toBe(0);
      expect(result.error.code).not.toBe('CLIENT_VALIDATION');
    }
  });

  it('does not claim validation for an unrelated TypeError', async () => {
    // A genuine programming bug must not be relabelled as a request problem.
    const result = await oneKitResult(() => {
      throw new TypeError('undefined is not a function');
    });

    expect(result.isErr()).toBe(true);
    if (result.isErr()) expect(result.error.code).not.toBe('CLIENT_VALIDATION');
  });

  it('still maps a generated-client HTTP error by its statusCode', async () => {
    const result = await oneKitResult(() => {
      const err = new Error('boom') as Error & { statusCode: number; body: string };
      err.statusCode = 404;
      err.body = JSON.stringify({ message: 'Campaign not found' });
      throw err;
    });

    expect(result.isErr()).toBe(true);
    if (result.isErr()) {
      expect(result.error.status).toBe(404);
      expect(result.error.code).toBe('NOT_FOUND');
    }
  });
});
