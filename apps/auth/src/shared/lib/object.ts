/**
 * Generic object helpers used across the auth app.
 */

/**
 * Returns a new object with the specified keys omitted.
 */
export function omit<T extends Record<string, unknown>, K extends keyof T>(
  obj: T,
  keys: K[]
): Omit<T, K> {
  const result = { ...obj };
  for (const key of keys) {
    delete result[key];
  }
  return result as Omit<T, K>;
}

/**
 * Returns a new object containing only the specified keys.
 */
export function pick<T extends Record<string, unknown>, K extends keyof T>(
  obj: T,
  keys: K[]
): Pick<T, K> {
  const result = {} as Pick<T, K>;
  for (const key of keys) {
    if (key in obj) {
      result[key] = obj[key];
    }
  }
  return result;
}

/**
 * Returns true when an object has no enumerable own properties.
 */
export function isEmpty(obj: Record<string, unknown>): boolean {
  return Object.keys(obj).length === 0;
}

/**
 * Performs a shallow merge of two objects, with b overriding a.
 */
export function merge<T extends Record<string, unknown>>(a: T, b: Partial<T>): T {
  return { ...a, ...b };
}

/**
 * Returns all enumerable [key, value] pairs of an object as a typed array.
 */
export function entries<T>(obj: Record<string, T>): [string, T][] {
  return Object.entries(obj);
}

/**
 * Builds an object from an array of [key, value] pairs.
 */
export function fromEntries<T>(pairs: [string, T][]): Record<string, T> {
  return Object.fromEntries(pairs);
}
