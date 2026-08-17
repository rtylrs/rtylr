import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiProject, ProjectUpdate } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';
import { isApiWorkspaceId } from '../../shared/lib/personalWorkspace';

export function useProjects(workspaceId: string | undefined) {
  return useQuery<ApiProject[]>({
    queryKey: ['projects', workspaceId],
    queryFn: async () => {
      const res = await unwrap(flowApi.getProjects(workspaceId!));
      return unwrapListResponse<ApiProject>(res, 'projects');
    },
    // Not `!!workspaceId`: the Personal sentinel is a non-empty string, so bare
    // truthiness fired this query with `workspace_id=personal` and always 400d.
    enabled: isApiWorkspaceId(workspaceId),
  });
}

export function useCreateProject() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { workspaceId: string; name: string; color?: string }) => {
      return unwrap(flowApi.createProject(data));
    },
    // Invalidate by prefix - the caller's workspaceId may be the "personal"
    // sentinel while the fetched query keyed with the resolved numeric ID,
    // so keying the invalidation on variables.workspaceId would miss.
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
    onError: () => toast.error(t('toasts.project.createFailed')),
  });
}

export function useUpdateProject() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: ProjectUpdate }) => {
      return unwrap(flowApi.updateProject(id, data)).then(() => {});
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
    onError: () => toast.error(t('toasts.project.updateFailed')),
  });
}

export function useDeleteProject() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      return unwrap(flowApi.deleteProject(id));
    },
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
    onError: () => toast.error(t('toasts.project.deleteFailed')),
  });
}

export function useMoveProject() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { project_id: number; workspace_id: number }) =>
      unwrap(flowApi.moveProject(data)).then(() => {}),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['projects'] }),
    onError: () => toast.error(t('toasts.project.moveFailed')),
  });
}
