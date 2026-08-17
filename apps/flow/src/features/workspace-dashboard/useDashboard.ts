import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { isApiWorkspaceId } from '../../shared/lib/personalWorkspace';

export function useFlowDashboard(workspaceId?: string) {
  return useQuery({
    queryKey: ['flow-dashboard', workspaceId],
    queryFn: () => unwrap(flowApi.getDashboard(workspaceId)),
    // Not `!!workspaceId`: the Personal sentinel is a non-empty string, so bare
    // truthiness fired this query with `workspace_id=personal` and always 400d.
    enabled: isApiWorkspaceId(workspaceId),
  });
}
