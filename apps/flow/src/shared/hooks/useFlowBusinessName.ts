import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';

export function useFlowBusinessName() {
  return useQuery<string | null>({
    queryKey: ['flow-business-name'],
    queryFn: async () => {
      const response = await unwrap(flowApi.getMyBusinesses());
      const businesses = response.businesses ?? [];
      // TanStack Query v5 treats a queryFn resolving to `undefined` as a bug
      // and throws ("Query data cannot be undefined") - null is the valid way
      // to say "no business name available" for a caller with zero businesses.
      return businesses.find((business) => business.isActive)?.name ?? businesses[0]?.name ?? null;
    },
    staleTime: 30_000,
  });
}

/**
 * The active business id, for the `business:<id>` realtime topic.
 *
 * Shares the businesses query with useFlowBusinessName rather than fetching
 * again - same queryKey, different projection.
 */
export function useFlowBusinessId() {
  return useQuery<string | null>({
    queryKey: ['flow-business-name'],
    queryFn: async () => {
      const response = await unwrap(flowApi.getMyBusinesses());
      const businesses = response.businesses ?? [];
      const active = businesses.find((business) => business.isActive) ?? businesses[0];
      return active ? String(active.id) : null;
    },
    staleTime: 30_000,
  });
}
