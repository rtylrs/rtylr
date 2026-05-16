/** True when the value is null/undefined or an all-whitespace string. */
export function isBlank(v: string | null | undefined): boolean {
  return v == null || v.trim().length === 0;
}
