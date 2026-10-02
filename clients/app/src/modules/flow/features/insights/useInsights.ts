import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiInsights } from '../../shared/services/flowApi';

/**
 * Business-wide Flow insights.
 *
 * This used to take a workspace id and carry `enabled: !!workspaceId`, so once
 * the workspace layer was stubbed to an empty string the query stopped firing
 * altogether and the screen sat empty. There is nothing to gate on now.
 */
export function useInsights() {
  return useQuery({
    queryKey: ['insights'],
    queryFn: async (): Promise<ApiInsights> => {
      const response = await unwrap(flowApi.getInsights());
      // The v2 overview can omit a whole series (its registered handler
      // returns an empty response, and the generated decoder leaves those
      // fields undefined despite the contract). Fill them in at the boundary
      // so the screen renders its empty states instead of throwing on
      // `data.velocity.reduce(...)`.
      return {
        ...response,
        velocity: response.velocity ?? [],
        burndown: response.burndown ?? [],
        throughput: response.throughput ?? [],
        timeByMember: response.timeByMember ?? [],
      };
    },
  });
}
