import { sleep } from './sleep';

export interface RetryOptions {
  /** Number of attempts before failing. */
  attempts?: number;
  /** Initial backoff delay in ms (doubled per attempt). */
  baseMs?: number;
  /** Cap the delay. */
  maxMs?: number;
}

/** Run `fn` up to N times with exponential backoff. */
export async function retry<T>(fn: () => Promise<T>, opts: RetryOptions = {}): Promise<T> {
  const attempts = opts.attempts ?? 3;
  const baseMs = opts.baseMs ?? 200;
  const maxMs = opts.maxMs ?? 5_000;
  let lastError: unknown;
  for (let i = 0; i < attempts; i++) {
    try {
      return await fn();
    } catch (err) {
      lastError = err;
      if (i === attempts - 1) break;
      const delay = Math.min(maxMs, baseMs * 2 ** i);
      await sleep(delay);
    }
  }
  throw lastError;
}
