import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi, ApiCycle } from '../services/flowApi';

export function useCycles(workspaceId?: string) {
  return useQuery({
    queryKey: ['cycles', workspaceId],
    queryFn: () => flowApi.getCycles(workspaceId),
  });
}

export function useCycleById(id: string | undefined) {
  const { data: cycles } = useCycles();
  return cycles?.find((c) => c.id === id) ?? null;
}

export function useCreateCycle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<ApiCycle>) => flowApi.createCycle(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
  });
}

export function useUpdateCycle() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<ApiCycle> }) =>
      flowApi.updateCycle(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['cycles'] }),
  });
}
