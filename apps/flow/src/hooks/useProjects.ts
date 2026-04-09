import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../services/flowApi';

export function useProjects(workspaceId: string | undefined) {
  return useQuery({
    queryKey: ['projects', workspaceId],
    queryFn: () => flowApi.getProjects(workspaceId!),
    enabled: !!workspaceId,
  });
}

export function useCreateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspaceId: string; name: string; color?: string }) => flowApi.createProject(data),
    onSuccess: (_data, variables) => {
      qc.invalidateQueries({ queryKey: ['projects', variables.workspaceId] });
    },
  });
}

export function useUpdateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: Partial<{ name: string; color: string }> }) =>
      flowApi.updateProject(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
  });
}

export function useDeleteProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => flowApi.deleteProject(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
  });
}
