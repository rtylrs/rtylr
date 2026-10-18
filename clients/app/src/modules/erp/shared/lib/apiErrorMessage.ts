/**
 * The reason the server gave for rejecting a request, or the translated
 * fallback. Transport failures ("Failed to fetch") are not shown verbatim.
 */
export function apiErrorMessage(error: unknown, fallback: string): string {
  if (error && typeof error === 'object' && 'body' in error && typeof error.body === 'string') {
    try {
      const parsed: unknown = JSON.parse(error.body);
      if (parsed && typeof parsed === 'object' && 'message' in parsed) {
        const { message } = parsed as { message?: unknown };
        if (typeof message === 'string' && message.trim()) return message;
      }
    } catch {
      // Not a JSON error body.
    }
  }
  return fallback;
}
