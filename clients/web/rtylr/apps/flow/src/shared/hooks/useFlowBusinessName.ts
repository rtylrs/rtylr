import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';
import type { ApiBusinessSummary } from '../services/flowApi';

/**
 * The active business, fetched once for both projections below.
 *
 * TanStack Query keeps one cache entry per queryKey, and that entry has a
 * single queryFn: two observers on the same key with different projections
 * did not each get their own - the first-mounted observer's queryFn produced
 * the value both hooks read, and a later refetch re-applied whichever
 * observer triggered it. So the id hook returned the business name, the
 * realtime client subscribed to `business:<name>`, and a refetch could flip
 * the shell title to the numeric id. One shared value with a per-observer
 * `select` keeps each projection correct.
 */
function activeBusinessQueryOptions() {
  return {
    queryKey: ['flow-active-business'],
    queryFn: async (): Promise<ApiBusinessSummary | null> => {
      const response = await unwrap(flowApi.getMyBusinesses());
      const businesses = response.businesses ?? [];
      return businesses.find((business) => business.isActive) ?? businesses[0] ?? null;
    },
    staleTime: 30_000,
  };
}

export function useFlowBusinessName() {
  return useQuery({
    ...activeBusinessQueryOptions(),
    // TanStack Query v5 treats a queryFn resolving to `undefined` as a bug
    // and throws ("Query data cannot be undefined") - null is the valid way
    // to say "no business name available" for a caller with zero businesses.
    select: (business) => business?.name ?? null,
  });
}

/**
 * The active business id, for the `business:<id>` realtime topic.
 *
 * Shares the businesses query with useFlowBusinessName rather than fetching
 * again - same cache entry, derived with `select`.
 */
export function useFlowBusinessId() {
  return useQuery({
    ...activeBusinessQueryOptions(),
    select: (business) => (business ? String(business.id) : null),
  });
}
