const RETRYABLE_HTTP_STATUS = new Set([408, 429]);

/**
 * True when a failure looks transport-related: no HTTP response at all
 * (network down, DNS, timeout - oneKit clients throw plain TypeErrors), or a
 * status the caller may legitimately retry (5xx, 408, 429). Definitive 4xx
 * business rejections must NOT be replayed offline - they would fail forever
 * as poison queue entries.
 */
export function isRetryableTransportError(error: unknown): boolean {
  if (error && typeof error === 'object' && 'statusCode' in error) {
    const status = (error as { statusCode?: unknown }).statusCode;
    if (typeof status === 'number') {
      return status >= 500 || RETRYABLE_HTTP_STATUS.has(status);
    }
  }
  return true;
}
