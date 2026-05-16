/**
 * Normalize line endings and trim trailing whitespace. Conservative — does
 * not touch any unicode characters so it stays safe with multi-byte input.
 */
export function cleanupInput(s: string): string {
  return s.replace(/\r\n?/g, '\n').trimEnd();
}
