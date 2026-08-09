// Error handling utilities for the insights app.
// These helpers extract user-friendly messages from API errors,
// classify network vs server errors, and build retry policies.

/**
 * extractErrorMessage returns a human-readable error message from an
 * unknown error value. Falls back to a generic message.
 */
export function extractErrorMessage(error: unknown, fallback = 'Something went wrong.'): string {
  if (!error) return fallback;
  if (typeof error === 'string') return error || fallback;
  if (error instanceof Error) return error.message || fallback;
  if (typeof error === 'object' && error !== null) {
    const e = error as Record<string, unknown>;
    if (typeof e['message'] === 'string') return e['message'] || fallback;
    if (typeof e['error'] === 'string') return e['error'] || fallback;
  }
  return fallback;
}

/**
 * isNetworkError returns true when the error looks like a connectivity
 * failure (timeout, DNS, etc.) rather than a server-side error.
 */
export function isNetworkError(error: unknown): boolean {
  if (!(error instanceof Error)) return false;
  const msg = error.message.toLowerCase();
  return (
    msg.includes('network') ||
    msg.includes('failed to fetch') ||
    msg.includes('load failed') ||
    msg.includes('timeout') ||
    msg.includes('abort')
  );
}

/**
 * isServerError returns true when the error has an HTTP status >= 500.
 */
export function isServerError(error: unknown): boolean {
  if (typeof error === 'object' && error !== null) {
    const e = error as Record<string, unknown>;
    const status = e['status'] ?? e['statusCode'];
    if (typeof status === 'number') return status >= 500;
  }
  return false;
}

/**
 * isUnauthorizedError returns true when the error has an HTTP 401 status.
 */
export function isUnauthorizedError(error: unknown): boolean {
  if (typeof error === 'object' && error !== null) {
    const e = error as Record<string, unknown>;
    const status = e['status'] ?? e['statusCode'];
    if (typeof status === 'number') return status === 401;
  }
  return false;
}

/**
 * shouldRetry returns true when the error type warrants an automatic
 * retry. Network errors and 5xx server errors are retried; 4xx
 * client errors are not.
 */
export function shouldRetry(error: unknown, attemptIndex: number, maxRetries = 3): boolean {
  if (attemptIndex >= maxRetries) return false;
  if (isUnauthorizedError(error)) return false;
  if (isNetworkError(error)) return true;
  if (isServerError(error)) return true;
  return false;
}
