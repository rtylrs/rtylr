/** Type guard and coercion helpers for the auth app. */

/**
 * Returns true if value is a plain object (not null, array, or primitive).
 */
export function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/**
 * Returns true if value is a string.
 */
export function isString(value: unknown): value is string {
  return typeof value === 'string';
}

/**
 * Returns true if value is a finite number.
 */
export function isNumber(value: unknown): value is number {
  return typeof value === 'number' && isFinite(value);
}

/**
 * Returns true if value is a boolean.
 */
export function isBoolean(value: unknown): value is boolean {
  return typeof value === 'boolean';
}

/**
 * Returns true if value is null or undefined.
 */
export function isNullish(value: unknown): value is null | undefined {
  return value === null || value === undefined;
}

/**
 * Returns true if value is a non-empty array.
 */
export function isNonEmptyArray(value: unknown): value is unknown[] {
  return Array.isArray(value) && value.length > 0;
}

/**
 * Returns true if value is a function.
 */
export function isFunction(value: unknown): value is (...args: unknown[]) => unknown {
  return typeof value === 'function';
}

/**
 * Coerces an unknown value to a plain object, returning {} if not possible.
 */
export function toObject(value: unknown): Record<string, unknown> {
  if (isObject(value)) return value;
  return {};
}

/**
 * Coerces an unknown value to a string. Returns fallback if not a string.
 */
export function getString(value: unknown, fallback = ''): string {
  if (isString(value)) return value;
  return fallback;
}
