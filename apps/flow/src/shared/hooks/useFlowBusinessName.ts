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
