/** Compare two arrays of objects keyed by `keyOf`, return added/removed sets. */
export interface DiffResult<T> {
  added: T[];
  removed: T[];
  unchanged: T[];
}

export function diffById<T, K extends string | number>(
  prev: T[],
  next: T[],
  keyOf: (item: T) => K
): DiffResult<T> {
  const prevKeys = new Set(prev.map(keyOf));
  const nextKeys = new Set(next.map(keyOf));
  return {
    added: next.filter((n) => !prevKeys.has(keyOf(n))),
    removed: prev.filter((p) => !nextKeys.has(keyOf(p))),
    unchanged: next.filter((n) => prevKeys.has(keyOf(n))),
  };
}
