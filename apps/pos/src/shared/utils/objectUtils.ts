// Pure helpers for object manipulation and safe property access.

/** Return a shallow clone of an object with the specified keys omitted. */
export function omit<T extends object, K extends keyof T>(
  obj: T,
  keys: K[],
): Omit<T, K> {
  const result = { ...obj };
  for (const key of keys) {
    delete result[key];
  }
  return result as Omit<T, K>;
}

/** Return a shallow clone of an object with only the specified keys included. */
export function pick<T extends object, K extends keyof T>(
  obj: T,
  keys: K[],
): Pick<T, K> {
  const result = {} as Pick<T, K>;
  for (const key of keys) {
    if (key in obj) result[key] = obj[key];
  }
  return result;
}

/** Return true if an object has no own enumerable keys. */
export function isEmpty(obj: object): boolean {
  return Object.keys(obj).length === 0;
}

/** Merge two objects shallowly, with the second overriding the first. */
export function merge<T extends object>(base: T, overrides: Partial<T>): T {
  return { ...base, ...overrides };
}

/** Deeply compare two plain objects for value equality (one level deep). */
export function shallowEqual<T extends object>(a: T, b: T): boolean {
  const keysA = Object.keys(a) as (keyof T)[];
  const keysB = Object.keys(b) as (keyof T)[];
  if (keysA.length !== keysB.length) return false;
  return keysA.every((k) => a[k] === b[k]);
}

/** Return the value at a key, or a default if undefined. */
export function getOrDefault<T, K extends keyof T>(obj: T, key: K, fallback: T[K]): T[K] {
  return obj[key] !== undefined ? obj[key] : fallback;
}

/** Return an object with all falsy values removed. */
export function compact<T extends object>(obj: T): Partial<T> {
  const result: Partial<T> = {};
  for (const key of Object.keys(obj) as (keyof T)[]) {
    if (obj[key]) result[key] = obj[key];
  }
  return result;
}
