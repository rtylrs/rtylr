export const JOIN_RETRY_INTERVAL_MS = 5_000;
export const JOIN_RETRY_WINDOW_MS = 120_000;

export interface JoinWithRetryOptions<T> {
  join: () => Promise<T>;
  /** True for the answer that means "not yet": the cashier has not entered the number. */
  isNotYet: (error: unknown) => boolean;
  /** Called once, when the first "not yet" arrives. */
  onWaiting?: () => void;
  signal?: AbortSignal;
  intervalMs?: number;
  windowMs?: number;
}

function sleep(ms: number, signal: AbortSignal | undefined): Promise<void> {
  return new Promise((resolve, reject) => {
    const abort = (): void => {
      clearTimeout(timer);
      reject(new DOMException('Join cancelled', 'AbortError'));
    };
    const timer = setTimeout(() => {
      signal?.removeEventListener('abort', abort);
      resolve();
    }, ms);
    if (signal?.aborted) abort();
    else signal?.addEventListener('abort', abort, { once: true });
  });
}

/**
 * At a venue that only accepts numbers its staff entered, the customer is
 * often a few seconds ahead of the cashier. Keeps asking until the number
 * exists or the window closes, then rethrows the last "not yet".
 */
export async function joinWithRetry<T>(options: JoinWithRetryOptions<T>): Promise<T> {
  const intervalMs = options.intervalMs ?? JOIN_RETRY_INTERVAL_MS;
  const deadline = Date.now() + (options.windowMs ?? JOIN_RETRY_WINDOW_MS);
  let waiting = false;

  for (;;) {
    try {
      return await options.join();
    } catch (error) {
      if (options.signal?.aborted) throw new DOMException('Join cancelled', 'AbortError');
      if (!options.isNotYet(error) || Date.now() + intervalMs > deadline) throw error;
      if (!waiting) {
        waiting = true;
        options.onWaiting?.();
      }
      await sleep(intervalMs, options.signal);
    }
  }
}

export function isAbort(error: unknown): boolean {
  return (error as { name?: unknown } | null)?.name === 'AbortError';
}
