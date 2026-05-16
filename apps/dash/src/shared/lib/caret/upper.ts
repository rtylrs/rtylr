/** Uppercase helper that's safe for nullish values. */
export function upper(s: string | null | undefined): string {
  return (s ?? '').toUpperCase();
}
