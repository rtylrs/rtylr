import type { UseQueryResult } from '@tanstack/react-query';

/**
 * A query result that has genuinely failed - whether TanStack Query has
 * settled it into `status: 'error'`, or it is stuck at `status: 'pending'`,
 * `fetchStatus: 'paused'` after already failing at least once.
 *
 * The default `networkMode: 'online'` (see `main.tsx`) makes a query pause
 * instead of erroring out whenever the `onlineManager` believes the browser
 * is offline. A real failure - a 403 from a permission the caller lacks,
 * for example - can retry into that paused state and never reach `error`,
 * so a plain `isLoading` / `isError` check silently falls through to
 * whatever the "no data" branch renders. That is indistinguishable from a
 * genuinely empty result unless callers also treat "paused after a failed
 * attempt" as an error.
 *
 * `isPaused` alone is not enough: a query can be paused before its very
 * first attempt (nothing has failed yet), and that case must still render
 * as loading, not as an error. `failureCount > 0` is what tells the two
 * apart.
 */
export function hasQueryFailed(
  query: Pick<UseQueryResult, 'isError' | 'isPaused' | 'failureCount'>
): boolean {
  return query.isError || (query.isPaused && query.failureCount > 0);
}
