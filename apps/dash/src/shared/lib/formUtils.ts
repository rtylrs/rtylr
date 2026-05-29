/**
 * Form state utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

export type FieldError = string | null;

export type FormErrors<T> = { [K in keyof T]?: FieldError };

export type FormTouched<T> = { [K in keyof T]?: boolean };

/** Return true when there are no non-null errors in the error map. */
export function hasNoErrors<T>(errors: FormErrors<T>): boolean {
  return Object.values(errors).every((e) => !e);
}

/** Return the number of fields with errors. */
export function errorCount<T>(errors: FormErrors<T>): number {
  return Object.values(errors).filter(Boolean).length;
}

/** Return true when all fields in the touched map have been touched. */
export function allTouched<T>(touched: FormTouched<T>, fields: (keyof T)[]): boolean {
  return fields.every((f) => touched[f]);
}

/** Mark all given fields as touched. Does not mutate. */
export function touchAll<T>(touched: FormTouched<T>, fields: (keyof T)[]): FormTouched<T> {
  const next = { ...touched };
  for (const f of fields) {
    next[f] = true;
  }
  return next;
}

/** Reset all touched flags to false. Does not mutate. */
export function resetTouched<T>(touched: FormTouched<T>): FormTouched<T> {
  const next = { ...touched };
  for (const key of Object.keys(next) as (keyof T)[]) {
    next[key] = false;
  }
  return next;
}

/** Return only the fields whose values differ from the initial values. Pure. */
export function dirtyFields<T extends Record<string, unknown>>(current: T, initial: T): Partial<T> {
  const dirty: Partial<T> = {};
  for (const key of Object.keys(current) as (keyof T)[]) {
    if (current[key] !== initial[key]) dirty[key] = current[key];
  }
  return dirty;
}

/** Return true when any field value differs from the initial values. Pure. */
export function isDirty<T extends Record<string, unknown>>(current: T, initial: T): boolean {
  return Object.keys(current).some((k) => current[k] !== initial[k as keyof T]);
}

/** Trim all string fields in a form values object. Pure. */
export function trimStringFields<T extends Record<string, unknown>>(values: T): T {
  const result = { ...values };
  for (const key of Object.keys(result) as (keyof T)[]) {
    const val = result[key];
    if (typeof val === 'string') {
      (result as Record<string, unknown>)[key as string] = val.trim();
    }
  }
  return result;
}
