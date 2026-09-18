// Pure type assertion helpers for narrowing and runtime checks.

/** Assert that a value is not null or undefined, throwing if it is. */
export function assertDefined<T>(value: T | null | undefined, label = 'value'): T {
  if (value === null || value === undefined) {
    throw new Error(`Expected ${label} to be defined, but got ${value}`);
  }
  return value;
}

/** Return true if a value is a non-null object (not an array). */
export function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/** Return true if a value is a non-empty string. */
export function isNonEmptyString(value: unknown): value is string {
  return typeof value === 'string' && value.trim().length > 0;
}

/** Return true if a value is a finite number. */
export function isFiniteNumber(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value);
}

/** Return true if a value is a positive integer. */
export function isPositiveInteger(value: unknown): value is number {
  return isFiniteNumber(value) && Number.isInteger(value) && value > 0;
}

/** Narrow an unknown value to string, returning null if not a string. */
export function asString(value: unknown): string | null {
  return typeof value === 'string' ? value : null;
}

/** Narrow an unknown value to number, returning null if not a finite number. */
export function asNumber(value: unknown): number | null {
  if (typeof value === 'number' && Number.isFinite(value)) return value;
  if (typeof value === 'string') {
    const n = Number(value);
    return Number.isFinite(n) ? n : null;
  }
  return null;
}
