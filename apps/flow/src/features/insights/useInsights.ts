import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { isFlowLocal } from '../../shared/lib/flowLocalMode';
import { useFlowStore } from '../../shared/stores/flowStore';
import { buildInsights } from '../../shared/lib/localFlowComputed';

export function useInsights(workspaceId?: string) {
  const tasks = useFlowStore((s) => s.tasks);
  const computed = useMemo(() => {
    if (!workspaceId) return undefined;
    return buildInsights(workspaceId, tasks);
  }, [workspaceId, tasks]);

  const q = useQuery({
    queryKey: ['insights', workspaceId],
    queryFn: () => unwrap(flowApi.getInsights(workspaceId)),
    enabled: !isFlowLocal() && !!workspaceId,
  });
  return {
    ...q,
    data: isFlowLocal() ? computed : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}
