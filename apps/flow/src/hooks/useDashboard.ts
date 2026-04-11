import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';
import { isFlowLocal } from '../lib/flowLocalMode';
import { useFlowStore } from '../stores/flowStore';
import { buildDashboardStats } from '../lib/localFlowComputed';

export function useFlowDashboard(workspaceId?: string) {
  const tasks = useFlowStore((s) => s.tasks);
  const projects = useFlowStore((s) => s.projects);
  const computed = useMemo(() => {
    if (!workspaceId) return undefined;
    return buildDashboardStats(workspaceId, tasks, projects);
  }, [workspaceId, tasks, projects]);

  const q = useQuery({
    queryKey: ['flow-dashboard', workspaceId],
    queryFn: () => unwrap(flowApi.getDashboard(workspaceId)),
    enabled: !isFlowLocal() && !!workspaceId,
  });
  return {
    ...q,
    data: isFlowLocal() ? computed : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}
