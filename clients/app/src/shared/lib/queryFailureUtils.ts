import type { UseQueryResult } from '@tanstack/react-query';

/**
 * Treat a query that failed and then paused offline as failed. In online mode,
 * TanStack Query can pause a retry after its first failure without setting
 * `isError`; `failureCount` distinguishes that case from a query paused before
 * its first attempt.
 */
export function hasQueryFailed(
  query: Pick<UseQueryResult, 'isError' | 'isPaused' | 'failureCount'>
): boolean {
  return query.isError || (query.isPaused && query.failureCount > 0);
}
