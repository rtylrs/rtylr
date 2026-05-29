/**
 * Object utility helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

/** Pick specified keys from an object, returning a new object. */
export function pick<T extends object, K extends keyof T>(obj: T, keys: K[]): Pick<T, K> {
  const result = {} as Pick<T, K>;
  for (const key of keys) {
    if (key in obj) result[key] = obj[key];
  }
  return result;
}

/** Omit specified keys from an object, returning a new object. */
export function omit<T extends object, K extends keyof T>(obj: T, keys: K[]): Omit<T, K> {
  const result = { ...obj };
  for (const key of keys) {
    delete result[key];
  }
  return result as Omit<T, K>;
}

/** Perform a shallow merge of multiple objects (later keys win). */
export function merge<T extends object>(...objects: Partial<T>[]): Partial<T> {
  return Object.assign({}, ...objects);
}

/** Return true if the object has no own enumerable keys. */
export function isEmpty(obj: object): boolean {
  return Object.keys(obj).length === 0;
}

/** Recursively remove keys with null or undefined values from an object. */
export function compactObject<T extends Record<string, unknown>>(obj: T): Partial<T> {
  const result: Partial<T> = {};
  for (const key of Object.keys(obj) as (keyof T)[]) {
    const val = obj[key];
    if (val != null) {
      result[key] =
        typeof val === 'object' && !Array.isArray(val)
          ? (compactObject(val as Record<string, unknown>) as T[keyof T])
          : val;
    }
  }
  return result;
}

/** Map the values of an object using a transform function. */
export function mapValues<T extends Record<string, unknown>, R>(
  obj: T,
  fn: (value: T[keyof T], key: keyof T) => R
): Record<keyof T, R> {
  const result = {} as Record<keyof T, R>;
  for (const key of Object.keys(obj) as (keyof T)[]) {
    result[key] = fn(obj[key], key);
  }
  return result;
}

/** Return all keys of an object with a specific value type. */
export function keys<T extends object>(obj: T): (keyof T)[] {
  return Object.keys(obj) as (keyof T)[];
}

/** Deeply compare two plain objects for equality. */
export function deepEqual(a: unknown, b: unknown): boolean {
  if (a === b) return true;
  if (typeof a !== 'object' || typeof b !== 'object') return false;
  if (a === null || b === null) return false;
  const keysA = Object.keys(a as object);
  const keysB = Object.keys(b as object);
  if (keysA.length !== keysB.length) return false;
  return keysA.every((k) =>
    deepEqual((a as Record<string, unknown>)[k], (b as Record<string, unknown>)[k])
  );
}
