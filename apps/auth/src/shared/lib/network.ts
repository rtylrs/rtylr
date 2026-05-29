/** Network and HTTP status helpers for the auth app. */

/**
 * Returns true if the given HTTP status code indicates success (2xx).
 */
export function isSuccess(status: number): boolean {
  return status >= 200 && status < 300;
}

/**
 * Returns true if the status code indicates a client error (4xx).
 */
export function isClientError(status: number): boolean {
  return status >= 400 && status < 500;
}

/**
 * Returns true if the status code indicates a server error (5xx).
 */
export function isServerError(status: number): boolean {
  return status >= 500 && status < 600;
}

/**
 * Returns true if the status indicates a redirect (3xx).
 */
export function isRedirect(status: number): boolean {
  return status >= 300 && status < 400;
}

/**
 * Returns a human-readable label for common HTTP status codes.
 */
export function statusLabel(status: number): string {
  const labels: Record<number, string> = {
    200: 'OK',
    201: 'Created',
    204: 'No Content',
    301: 'Moved Permanently',
    302: 'Found',
    400: 'Bad Request',
    401: 'Unauthorized',
    403: 'Forbidden',
    404: 'Not Found',
    409: 'Conflict',
    422: 'Unprocessable Entity',
    429: 'Too Many Requests',
    500: 'Internal Server Error',
    502: 'Bad Gateway',
    503: 'Service Unavailable',
  };
  return labels[status] ?? `HTTP ${status}`;
}

/**
 * Returns true if the network is likely online (requires browser environment).
 */
export function isOnline(): boolean {
  if (typeof navigator === 'undefined') return true;
  return navigator.onLine;
}
