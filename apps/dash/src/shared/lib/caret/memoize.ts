/** Single-slot memoization keyed on a string of the args. */
export function memoize<TArgs extends unknown[], TResult>(
  fn: (...args: TArgs) => TResult,
  keyOf: (...args: TArgs) => string = (...a) => JSON.stringify(a)
): (...args: TArgs) => TResult {
  const cache = new Map<string, TResult>();
  return (...args: TArgs): TResult => {
    const k = keyOf(...args);
    if (cache.has(k)) return cache.get(k) as TResult;
    const v = fn(...args);
    cache.set(k, v);
    return v;
  };
}
