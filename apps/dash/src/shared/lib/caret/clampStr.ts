/** Truncate a string to `maxLen` chars, optionally adding an ellipsis. */
export function clampString(s: string, maxLen: number, suffix = '…'): string {
  if (s.length <= maxLen) return s;
  if (maxLen <= suffix.length) return suffix.slice(0, maxLen);
  return s.slice(0, maxLen - suffix.length) + suffix;
}
