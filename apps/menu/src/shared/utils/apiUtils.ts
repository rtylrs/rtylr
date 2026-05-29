/**
 * API error and response helpers for the menu app.
 */

/**
 * Shape of an Axios-like error response.
 */
interface ApiError {
  response?: {
    status?: number;
    data?: {
      message?: string;
      error?: string;
    };
  };
  message?: string;
}

/**
 * Extracts a human-readable error message from an unknown API error.
 */
export function extractErrorMessage(error: unknown, fallback = 'Something went wrong.'): string {
  const err = error as ApiError;
  return err?.response?.data?.message ?? err?.response?.data?.error ?? err?.message ?? fallback;
}

/**
 * Returns the HTTP status code from an API error, or null if unavailable.
 */
export function extractStatusCode(error: unknown): number | null {
  const err = error as ApiError;
  return err?.response?.status ?? null;
}

/**
 * Returns true when the error indicates a 401 Unauthorised response.
 */
export function isUnauthorised(error: unknown): boolean {
  return extractStatusCode(error) === 401;
}

/**
 * Returns true when the error indicates a 404 Not Found response.
 */
export function isNotFound(error: unknown): boolean {
  return extractStatusCode(error) === 404;
}

/**
 * Returns true when the error indicates a 5xx server-side failure.
 */
export function isServerError(error: unknown): boolean {
  const code = extractStatusCode(error);
  return code !== null && code >= 500;
}

/**
 * Builds a URL query string from an object of parameters.
 * Skips keys whose value is null, undefined, or an empty string.
 */
export function buildQueryString(params: Record<string, unknown>): string {
  const parts: string[] = [];
  for (const [key, value] of Object.entries(params)) {
    if (value === null || value === undefined || value === '') continue;
    parts.push(`${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`);
  }
  return parts.length ? `?${parts.join('&')}` : '';
}
