/**
 * CRM list endpoints return `{ <key>: [...] }` envelopes from the Go backend.
 * This helper accepts either a raw array or an enveloped response, so the
 * schema parse never sees undefined (which would throw).
 *
 * Pass one or more candidate keys - the first matching array wins.
 */
export function unwrapList<T>(res: unknown, ...keys: string[]): T[] {
  if (Array.isArray(res)) return res as T[];
  if (res && typeof res === 'object') {
    for (const key of keys) {
      const value = (res as Record<string, unknown>)[key];
      if (Array.isArray(value)) return value as T[];
    }
  }
  return [];
}
