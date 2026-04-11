import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { isFlowLocal } from '../../shared/lib/flowLocalMode';
import { useFlowStore } from '../../shared/stores/flowStore';
import type { ApiMilestone } from '../../shared/services/flowApi';

export function useMilestones(projectId?: string) {
  const milestones = useFlowStore((s) => s.milestones);
  const localList = useMemo(
    () => (projectId ? milestones.filter((m) => m.projectId === projectId) : milestones),
    [milestones, projectId]
  );
  const q = useQuery({
    queryKey: ['milestones', projectId],
    queryFn: () => unwrap(flowApi.getMilestones(projectId)),
    enabled: !isFlowLocal(),
  });
  return {
    ...q,
    data: isFlowLocal() ? localList : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}

export function useCreateMilestone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ApiMilestone>) => {
      if (isFlowLocal()) {
        const id = `mil-${Date.now()}`;
        const row = {
          id,
          projectId: data.projectId ?? '',
          name: data.name ?? 'Milestone',
          targetDate: data.targetDate ?? new Date().toISOString().slice(0, 10),
          progress: data.progress ?? 0,
          status: data.status ?? ('planning' as const),
          description: data.description,
        };
        useFlowStore.setState({ milestones: [...useFlowStore.getState().milestones, row] });
        return Promise.resolve(row);
      }
      return flowApi.createMilestone(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
  });
}

export function useUpdateMilestone() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiMilestone> }) => {
      if (isFlowLocal()) {
        const milestones = useFlowStore
          .getState()
          .milestones.map((m) => (m.id === id ? { ...m, ...data } : m));
        useFlowStore.setState({ milestones });
        return Promise.resolve();
      }
      return flowApi.updateMilestone(id, data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['milestones'] }),
  });
}
