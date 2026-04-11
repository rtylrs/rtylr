import { unwrap } from '@voxire/api';
import { useMemo } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import { isFlowLocal } from '../../shared/lib/flowLocalMode';
import { useFlowStore } from '../../shared/stores/flowStore';
import { PERSONAL_WORKSPACE_ID } from '../../shared/lib/personalWorkspace';
import type { ApiCycle } from '../../shared/services/flowApi';

export function useCycles(workspaceId?: string) {
  const cycles = useFlowStore((s) => s.cycles);
  const localList = useMemo(
    () => (workspaceId ? cycles.filter((c) => c.workspaceId === workspaceId) : cycles),
    [cycles, workspaceId]
  );
  const q = useQuery({
    queryKey: ['cycles', workspaceId],
    queryFn: () => unwrap(flowApi.getCycles(workspaceId)),
    enabled: !isFlowLocal(),
  });
  return {
    ...q,
    data: isFlowLocal() ? localList : q.data,
    isLoading: isFlowLocal() ? false : q.isLoading,
  };
}

export function useCreateCycle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ApiCycle>) => {
      if (isFlowLocal()) {
        const id = `cyc-${Date.now()}`;
        const row = {
          id,
          workspaceId: data.workspaceId ?? PERSONAL_WORKSPACE_ID,
          name: data.name ?? 'New cycle',
          startDate: data.startDate ?? new Date().toISOString().slice(0, 10),
          endDate: data.endDate ?? new Date().toISOString().slice(0, 10),
          status: data.status ?? ('upcoming' as const),
          scope: data.scope ?? 0,
          completed: data.completed ?? 0,
        };
        useFlowStore.setState({ cycles: [...useFlowStore.getState().cycles, row] });
        return Promise.resolve(row);
      }
      return flowApi.createCycle(data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
  });
}

export function useUpdateCycle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiCycle> }) => {
      if (isFlowLocal()) {
        const cycles = useFlowStore
          .getState()
          .cycles.map((c) => (c.id === id ? { ...c, ...data } : c));
        useFlowStore.setState({ cycles });
        return Promise.resolve();
      }
      return flowApi.updateCycle(id, data);
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
  });
}
