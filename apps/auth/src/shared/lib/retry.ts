/** Retry and backoff helpers for the auth app. */

export interface RetryOptions {
  maxAttempts?: number;
  baseDelayMs?: number;
  maxDelayMs?: number;
  shouldRetry?: (error: unknown) => boolean;
}

/**
 * Runs fn with exponential backoff. Throws the last error after all attempts fail.
 */
export async function withExponentialBackoff<T>(
  fn: () => Promise<T>,
  options: RetryOptions = {}
): Promise<T> {
  const { maxAttempts = 3, baseDelayMs = 100, maxDelayMs = 10_000, shouldRetry } = options;
  let lastError: unknown;
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (err) {
      lastError = err;
      if (attempt === maxAttempts) break;
      if (shouldRetry && !shouldRetry(err)) break;
      const delay = Math.min(baseDelayMs * 2 ** (attempt - 1), maxDelayMs);
      await new Promise<void>((resolve) => setTimeout(resolve, delay));
    }
  }
  throw lastError;
}

/**
 * Runs fn with a fixed delay between attempts. Throws the last error if all fail.
 */
export async function withFixedRetry<T>(
  fn: () => Promise<T>,
  maxAttempts = 3,
  delayMs = 100
): Promise<T> {
  let lastError: unknown;
  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      return await fn();
    } catch (err) {
      lastError = err;
      if (attempt < maxAttempts) {
        await new Promise<void>((resolve) => setTimeout(resolve, delayMs));
      }
    }
  }
  throw lastError;
}

/**
 * Wraps a function so it cannot be called again until cooldownMs has elapsed.
 * Calls during the cooldown period are silently ignored.
 */
export function withCooldown<T extends (...args: unknown[]) => unknown>(
  fn: T,
  cooldownMs: number
): T {
  let lastCallAt = 0;
  return ((...args: unknown[]) => {
    const now = Date.now();
    if (now - lastCallAt < cooldownMs) return;
    lastCallAt = now;
    return fn(...args);
  }) as T;
}
