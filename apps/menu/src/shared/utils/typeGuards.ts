/**
 * Type guard utilities for the Menu app.
 * All functions narrow the TypeScript type when they return true.
 */

/** Return true (and narrow type) if the value is a non-null object. */
export function isObject(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
}

/** Return true (and narrow type) if the value is a string. */
export function isString(value: unknown): value is string {
  return typeof value === 'string';
}

/** Return true (and narrow type) if the value is a number (not NaN). */
export function isNumber(value: unknown): value is number {
  return typeof value === 'number' && !Number.isNaN(value);
}

/** Return true (and narrow type) if the value is a boolean. */
export function isBoolean(value: unknown): value is boolean {
  return typeof value === 'boolean';
}

/** Return true (and narrow type) if the value is an array. */
export function isArray(value: unknown): value is unknown[] {
  return Array.isArray(value);
}

/** Return true (and narrow type) if the value is null or undefined. */
export function isNullish(value: unknown): value is null | undefined {
  return value === null || value === undefined;
}

/** Return true (and narrow type) if the value is a function. */
export function isFunction(value: unknown): value is (...args: unknown[]) => unknown {
  return typeof value === 'function';
}

/** Return true (and narrow type) if the value is a non-empty string. */
export function isNonEmptyString(value: unknown): value is string {
  return typeof value === 'string' && value.trim().length > 0;
}
