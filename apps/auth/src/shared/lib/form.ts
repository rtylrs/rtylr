/** Form state and validation helpers for the auth app. */

export type FieldErrors = Record<string, string>;

/**
 * Returns true if there are no field errors.
 */
export function isFormValid(errors: FieldErrors): boolean {
  return Object.values(errors).every((e) => !e);
}

/**
 * Returns the error message for a given field, or empty string if none.
 */
export function fieldError(errors: FieldErrors, field: string): string {
  return errors[field] ?? '';
}

/**
 * Returns true if a specific field has an error.
 */
export function hasFieldError(errors: FieldErrors, field: string): boolean {
  return Boolean(errors[field]);
}

/**
 * Returns a new errors object with the given field cleared.
 */
export function clearFieldError(errors: FieldErrors, field: string): FieldErrors {
  const next = { ...errors };
  delete next[field];
  return next;
}

/**
 * Returns all field names that have errors.
 */
export function invalidFields(errors: FieldErrors): string[] {
  return Object.entries(errors)
    .filter(([, v]) => Boolean(v))
    .map(([k]) => k);
}

/**
 * Merges two error objects, with the second taking priority on conflicts.
 */
export function mergeErrors(a: FieldErrors, b: FieldErrors): FieldErrors {
  return { ...a, ...b };
}
