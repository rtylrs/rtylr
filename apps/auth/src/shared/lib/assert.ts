/**
 * Lightweight assertion helpers for the auth app.
 */

/**
 * Throws an error with the given message when the condition is falsy.
 */
export function assert(condition: unknown, message: string): asserts condition {
  if (!condition) {
    throw new Error(message);
  }
}

/**
 * Throws an error when the value is null or undefined,
 * and narrows its type to NonNullable<T>.
 */
export function assertDefined<T>(value: T | null | undefined, message: string): T {
  if (value == null) {
    throw new Error(message);
  }
  return value;
}

/**
 * Asserts that the given string is non-empty (after trimming).
 */
export function assertNonEmpty(value: string, message: string): void {
  if (!value || value.trim().length === 0) {
    throw new Error(message);
  }
}

/**
 * Used in switch exhaustiveness checks.
 * Throws if the value is anything other than never.
 */
export function assertNever(value: never, message = 'Unexpected value'): never {
  throw new Error(`${message}: ${JSON.stringify(value)}`);
}
