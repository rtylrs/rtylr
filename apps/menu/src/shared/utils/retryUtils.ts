/**
 * Retry and backoff utilities for the Menu app.
 */

/** Sleep for a given number of milliseconds. */
export function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/** Calculate exponential backoff delay with optional jitter. */
export function backoffDelay(attempt: number, baseMs = 500, maxMs = 30_000, jitter = true): number {
  const delay = Math.min(baseMs * Math.pow(2, attempt), maxMs);
  if (!jitter) return delay;
  return delay * (0.5 + Math.random() * 0.5);
}

/**
 * Retry an async function up to maxAttempts times with exponential backoff.
 * Throws the last error if all attempts fail.
 */
export async function withRetry<T>(
  fn: () => Promise<T>,
  maxAttempts = 3,
  baseMs = 500
): Promise<T> {
  let lastError: unknown;
  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (err) {
      lastError = err;
      if (attempt < maxAttempts - 1) {
        await sleep(backoffDelay(attempt, baseMs, 30_000, false));
      }
    }
  }
  throw lastError;
}

/**
 * Wrap a function so it can only be called once.
 * Subsequent calls return the result of the first invocation.
 */
export function once<T>(fn: () => T): () => T {
  let called = false;
  let result: T;
  return () => {
    if (!called) {
      called = true;
      result = fn();
    }
    return result;
  };
}
