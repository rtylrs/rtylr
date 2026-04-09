import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';

export function useWorkspaces() {
  return useQuery({
    queryKey: ['workspaces'],
    queryFn: () => flowApi.getWorkspaces(),
  });
}

export function useCreateWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color: string }) => flowApi.createWorkspace(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workspaces'] }),
  });
}

export function useUpdateWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<{ name: string; color: string }> }) =>
      flowApi.updateWorkspace(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workspaces'] }),
  });
}

export function useDeleteWorkspace() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => flowApi.deleteWorkspace(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['workspaces'] }),
  });
}
