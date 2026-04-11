import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';
import { isFlowLocal } from '../lib/flowLocalMode';
import { useFlowStore } from '../stores/flowStore';
import { buildInsights } from '../lib/localFlowComputed';

export function useInsights(workspaceId?: string) {
  const tasks = useFlowStore((s) => s.tasks);
  const computed = useMemo(() => {
    if (!workspaceId) return undefined;
    return buildInsights(workspaceId, tasks);
  }, [workspaceId, tasks]);

  const q = useQuery({
    queryKey: ['insights', workspaceId],
    queryFn: () => flowApi.getInsights(workspaceId),
    enabled: !isFlowLocal() && !!workspaceId,
  });
  return {
    ...q,
    data: isFlowLocal() ? computed : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}
