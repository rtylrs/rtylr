/**
 * Error handling utilities for the auth app.
 */

/**
 * Extracts a human-readable message from an unknown error value.
 */
export function getErrorMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === 'string') return error;
  if (
    typeof error === 'object' &&
    error !== null &&
    'message' in error &&
    typeof (error as Record<string, unknown>).message === 'string'
  ) {
    return (error as { message: string }).message;
  }
  return 'An unexpected error occurred';
}

/**
 * Returns true when the value is an Error instance.
 */
export function isError(value: unknown): value is Error {
  return value instanceof Error;
}

/**
 * Creates a typed Error with an optional cause.
 */
export function createError(message: string, cause?: unknown): Error {
  const err = new Error(message);
  if (cause !== undefined) {
    (err as Error & { cause: unknown }).cause = cause;
  }
  return err;
}

/**
 * Returns a fallback value when the supplier function throws.
 */
export function safely<T>(fn: () => T, fallback: T): T {
  try {
    return fn();
  } catch {
    return fallback;
  }
}

/**
 * Checks whether an error represents an HTTP 401 Unauthorized response.
 */
export function isUnauthorizedError(error: unknown): boolean {
  if (typeof error !== 'object' || error === null) return false;
  const status =
    (error as { response?: { status?: number } }).response?.status ??
    (error as { status?: number }).status;
  return status === 401;
}
