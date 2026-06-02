import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiCycle } from '../../shared/services/flowApi';
import { useFlowStore } from '../../shared/stores/flowStore';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

export function useCycles(workspaceId?: string): ReturnType<typeof useQuery<ApiCycle[]>> {
  const { activeWorkspaceId } = useFlowStore();
  const wsId = workspaceId ?? activeWorkspaceId;
  return useQuery<ApiCycle[]>({
    queryKey: ['cycles', wsId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getCycles(wsId));
      return unwrapListResponse<ApiCycle>(res, 'cycles');
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
    onError: () => toast.error('Failed to create cycle'),
  });
}

export function useUpdateCycle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiCycle> }) => {
      return unwrap(flowApi.updateCycle(id, data)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
    onError: () => toast.error('Failed to update cycle'),
  });
}
