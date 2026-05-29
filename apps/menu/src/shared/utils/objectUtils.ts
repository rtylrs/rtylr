/**
 * Generic object utility functions for the Menu app.
 */

/** Return true if the object has no own enumerable keys. */
export function isEmptyObject(obj: Record<string, unknown>): boolean {
  return Object.keys(obj).length === 0;
}

/** Return a copy of the object without the specified keys. */
export function omit<T extends object, K extends keyof T>(obj: T, keys: K[]): Omit<T, K> {
  const result = { ...obj };
  for (const key of keys) {
    delete result[key];
  }
  return result as Omit<T, K>;
}

/** Return a new object containing only the specified keys. */
export function pick<T extends object, K extends keyof T>(obj: T, keys: K[]): Pick<T, K> {
  const result = {} as Pick<T, K>;
  for (const key of keys) {
    if (key in obj) {
      result[key] = obj[key];
    }
  }
  return result;
}

/** Return a copy of the object with all falsy values removed. */
export function compactObject<T extends object>(obj: T): Partial<T> {
  const result: Partial<T> = {};
  for (const key of Object.keys(obj) as Array<keyof T>) {
    if (obj[key]) {
      result[key] = obj[key];
    }
  }
  return result;
}

/** Return a deep clone of the value using structured clone. */
export function deepClone<T>(value: T): T {
  return structuredClone(value);
}

/** Return the value at a dot-notation path, or undefined if not found. */
export function getByPath(obj: Record<string, unknown>, path: string): unknown {
  return path.split('.').reduce<unknown>((current, segment) => {
    if (current === null || current === undefined || typeof current !== 'object') {
      return undefined;
    }
    return (current as Record<string, unknown>)[segment];
  }, obj);
}
