/** Wrap a thunk so it only runs once, caching the result. */
export function once<T>(fn: () => T): () => T {
  let called = false;
  let result: T;
  return (): T => {
    if (!called) {
      result = fn();
      called = true;
    }
    return result;
  };
}
