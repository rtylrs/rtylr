/**
 * CRM list endpoints return `{ <key>: [...] }` envelopes from the Go backend.
 * This helper accepts either a raw array or an enveloped response. A malformed
 * success response is an error, not an empty list, so the query can show its
 * error state instead of hiding missing data.
 *
 * An empty list has two encodings that are not malformed: generated OneKit
 * responses omit an empty repeated field (`{"total":"0"}`), and hand-written
 * handlers serialize a nil slice as `null`. Both read as `[]`; treating them
 * as errors showed every new business an error on every CRM list screen. An
 * envelope that carries an array under some other key still throws, since
 * that is a key mismatch rather than an empty page.
 *
 * Pass one or more candidate keys - the first matching array wins.
 */
export function unwrapList<T>(res: unknown, ...keys: string[]): T[] {
  if (Array.isArray(res)) return res as T[];
  if (res && typeof res === 'object') {
    const envelope = res as Record<string, unknown>;
    for (const key of keys) {
      const value = envelope[key];
      if (Array.isArray(value)) return value as T[];
      if (value === null) return [];
    }
    const hasCandidate = keys.some((key) => envelope[key] !== undefined);
    const hasOtherList = Object.values(envelope).some(Array.isArray);
    if (!hasCandidate && !hasOtherList) return [];
  }
  const expected = keys.length > 0 ? keys.map((key) => `"${key}"`).join(' or ') : 'an array';
  throw new TypeError(`Expected a list response with ${expected}`);
}
