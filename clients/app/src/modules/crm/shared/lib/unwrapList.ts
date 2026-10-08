/**
 * CRM list endpoints return `{ <key>: [...] }` envelopes from the Go backend.
 * This helper accepts either a raw array or an enveloped response. A malformed
 * success response is an error, not an empty list, so the query can show its
 * error state instead of hiding missing data.
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
  const expected = keys.length > 0 ? keys.map((key) => `"${key}"`).join(' or ') : 'an array';
  throw new TypeError(`Expected a list response with ${expected}`);
}
