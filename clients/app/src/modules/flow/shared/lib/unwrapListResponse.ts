/**
 * Every Go handler in `onion` that returns a list wraps the array in an
 * envelope: `{ tasks: [...] }`, `{ projects: [...] }`, `{ members: [...] }`,
 * etc. This matches the voxie proto `ListXResponse` shape.
 *
 * The frontend has historically typed those endpoints as raw arrays
 * (`ApiTask[]`) which bypasses the compile-time check and explodes at
 * runtime with "X.map is not a function".
 *
 * Use this helper at the boundary between `flowApi.getX()` and the hook /
 * component:
 *
 *     const res = await unwrap(flowApi.getMyTasks());
 *     return unwrapListResponse<ApiTask>(res, 'tasks');
 *
 * It accepts either a raw array (legacy) or the enveloped object so the
 * switch is gradual.
 *
 * A successful response with the wrong shape is an error, not an empty list.
 */
export function unwrapListResponse<T>(res: unknown, key: string): T[] {
  if (Array.isArray(res)) return res as T[];
  if (res && typeof res === 'object') {
    const value = (res as Record<string, unknown>)[key];
    if (Array.isArray(value)) return value as T[];
  }
  throw new TypeError(`Expected a list response with "${key}"`);
}
