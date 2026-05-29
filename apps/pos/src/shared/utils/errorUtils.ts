// Pure helpers for error message extraction and classification.

/** Extract a human-readable message from an unknown error value. */
export function extractErrorMessage(error: unknown, fallback = 'An unexpected error occurred'): string {
  if (typeof error === 'string') return error || fallback;
  if (error instanceof Error) return error.message || fallback;
  if (
    error !== null &&
    typeof error === 'object' &&
    'message' in error &&
    typeof (error as { message: unknown }).message === 'string'
  ) {
    return (error as { message: string }).message || fallback;
  }
  return fallback;
}

/** Return true if an error looks like a network connectivity error. */
export function isNetworkError(error: unknown): boolean {
  const msg = extractErrorMessage(error, '').toLowerCase();
  return (
    msg.includes('network') ||
    msg.includes('fetch') ||
    msg.includes('offline') ||
    msg.includes('connection') ||
    msg.includes('econnrefused')
  );
}

/** Return true if an error looks like an HTTP 4xx client error. */
export function isClientError(statusCode: number): boolean {
  return Number.isFinite(statusCode) && statusCode >= 400 && statusCode < 500;
}

/** Return true if an error looks like an HTTP 5xx server error. */
export function isServerError(statusCode: number): boolean {
  return Number.isFinite(statusCode) && statusCode >= 500 && statusCode < 600;
}

/** Return a user-friendly label for a given HTTP status code. */
export function httpStatusLabel(statusCode: number): string {
  const labels: Record<number, string> = {
    400: 'Bad request',
    401: 'Unauthorized',
    403: 'Forbidden',
    404: 'Not found',
    409: 'Conflict',
    422: 'Validation error',
    429: 'Too many requests',
    500: 'Server error',
    502: 'Bad gateway',
    503: 'Service unavailable',
  };
  return labels[statusCode] ?? `HTTP ${statusCode}`;
}
