/** Return `value` when non-nullish, else `fallback`. */
export function withDefault<T>(value: T | null | undefined, fallback: T): T {
  return value == null ? fallback : value;
}
