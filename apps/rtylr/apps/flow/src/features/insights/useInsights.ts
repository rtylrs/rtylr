import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';

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
    queryFn: () => unwrap(flowApi.getInsights()),
  });
}
