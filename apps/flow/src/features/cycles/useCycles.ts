import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiCycle } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';

export function useCycles(workspaceId?: string) {
  const { activeWorkspaceId } = useFlowStore();
  const wsId = workspaceId ?? activeWorkspaceId;
  return useQuery({
    queryKey: ['cycles', wsId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getCycles(wsId));
      // backend returns { cycles: [...] }
      return (res as any).cycles ?? (Array.isArray(res) ? res : []);
    },
  });
}

export function useCreateCycle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ApiCycle>) => {
      return unwrap(flowApi.createCycle(data));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
  });
}

export function useUpdateCycle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiCycle> }) => {
      return unwrap(flowApi.updateCycle(id, data)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
  });
}
