/**
 * Error handling utilities for the Dash app.
 * All functions are pure with no side-effects.
 */

/** Extract a human-readable message from any thrown value. */
export function extractMessage(error: unknown): string {
  if (error instanceof Error) return error.message;
  if (typeof error === 'string') return error;
  if (typeof error === 'object' && error !== null && 'message' in error) {
    return String((error as { message: unknown }).message);
  }
  return 'An unexpected error occurred';
}

/** Return true if the error looks like a network connectivity failure. */
export function isNetworkError(error: unknown): boolean {
  const message = extractMessage(error).toLowerCase();
  return (
    message.includes('network') ||
    message.includes('fetch') ||
    message.includes('connection') ||
    message.includes('timeout') ||
    message.includes('offline')
  );
}

/** Return true if the error indicates an HTTP 401 Unauthorized response. */
export function isUnauthorizedError(error: unknown): boolean {
  if (typeof error === 'object' && error !== null && 'statusCode' in error) {
    return (error as { statusCode: unknown }).statusCode === 401;
  }
  return false;
}

/** Return true if the error indicates an HTTP 403 Forbidden response. */
export function isForbiddenError(error: unknown): boolean {
  if (typeof error === 'object' && error !== null && 'statusCode' in error) {
    return (error as { statusCode: unknown }).statusCode === 403;
  }
  return false;
}

/** Return true if the error indicates an HTTP 404 Not Found response. */
export function isNotFoundError(error: unknown): boolean {
  if (typeof error === 'object' && error !== null && 'statusCode' in error) {
    return (error as { statusCode: unknown }).statusCode === 404;
  }
  return false;
}

/** Return a short human-readable label for a common HTTP status code. */
export function httpStatusLabel(statusCode: number): string {
  if (statusCode === 200) return 'OK';
  if (statusCode === 201) return 'Created';
  if (statusCode === 400) return 'Bad Request';
  if (statusCode === 401) return 'Unauthorized';
  if (statusCode === 403) return 'Forbidden';
  if (statusCode === 404) return 'Not Found';
  if (statusCode === 409) return 'Conflict';
  if (statusCode === 422) return 'Unprocessable';
  if (statusCode === 429) return 'Rate Limited';
  if (statusCode >= 500 && statusCode < 600) return 'Server Error';
  return `HTTP ${statusCode}`;
}
