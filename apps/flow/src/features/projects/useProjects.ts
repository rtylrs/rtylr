import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiProject } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

export function useProjects(workspaceId: string | undefined) {
  return useQuery<ApiProject[]>({
    queryKey: ['projects', workspaceId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getProjects(workspaceId!));
      return unwrapListResponse<ApiProject>(res, 'projects');
    },
    enabled: !!workspaceId,
  });
}

export function useCreateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspaceId: string; name: string; color?: string }) => {
      return unwrap(flowApi.createProject(data));
    },
    // Invalidate by prefix - the caller's workspaceId may be the "personal"
    // sentinel while the fetched query keyed with the resolved numeric ID,
    // so keying the invalidation on variables.workspaceId would miss.
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
    onError: () => toast.error('Failed to create project'),
  });
}

export function useUpdateProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      data,
    }: {
      id: string;
      data: Partial<{ name: string; color: string; description: string }>;
    }) => {
      return unwrap(flowApi.updateProject(id, data)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
    onError: () => toast.error('Failed to update project'),
  });
}

export function useDeleteProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      return unwrap(flowApi.deleteProject(id));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
    onError: () => toast.error('Failed to delete project'),
  });
}

export function useMoveProject() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { project_id: number; workspace_id: number }) =>
      unwrap(flowApi.moveProject(data)).then(() => {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
    onError: () => toast.error('Failed to move project'),
  });
}
