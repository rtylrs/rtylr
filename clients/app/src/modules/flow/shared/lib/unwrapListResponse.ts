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
 *
 * One exception: the server omits an empty array from its JSON (proto3 default
 * for a repeated field), so a user with no rows gets `{ "pagination": {...} }`
 * with no `tasks` key at all. That is an empty list. It is told apart from a
 * wrong-shape response by the rest of the envelope: only an envelope that holds
 * nothing but the standard `pagination` block (or nothing at all) reads as
 * empty. Anything else, such as the data under a mistyped key, still throws.
 */
export function unwrapListResponse<T>(res: unknown, key: string): T[] {
  if (Array.isArray(res)) return res as T[];
  if (res && typeof res === 'object') {
    const envelope = res as Record<string, unknown>;
    const value = envelope[key];
    if (Array.isArray(value)) return value as T[];
    const onlyPagination = Object.keys(envelope).every((k) => k === key || k === 'pagination');
    if (value === undefined && onlyPagination) return [];
  }
  throw new TypeError(`Expected a list response with "${key}"`);
}
