import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiCycle } from '../../shared/services/flowApi';

export function useCycles(workspaceId?: string) {
  return useQuery({
    queryKey: ['cycles', workspaceId],
    queryFn: () => unwrap(flowApi.getCycles(workspaceId)),
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
