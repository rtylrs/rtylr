/** Lowercase helper that's safe for nullish values. */
export function lower(s: string | null | undefined): string {
  return (s ?? '').toLowerCase();
}
