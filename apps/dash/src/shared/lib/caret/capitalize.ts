/** Capitalize the first character of `s`. Returns "" for empty input. */
export function capitalize(s: string | null | undefined): string {
  if (!s) return '';
  return s[0].toUpperCase() + s.slice(1);
}
