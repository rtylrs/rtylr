import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiProject, ProjectUpdate } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

export function useProjects() {
  return useQuery<ApiProject[]>({
    queryKey: ['projects'],
    queryFn: async () => {
      const res = await unwrap(flowApi.getProjects());
      return unwrapListResponse<ApiProject>(res, 'projects');
    },
  });
}

export function useCreateProject() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { name: string; color?: string; visibility?: string }) => {
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

// Named for what it does. It archives: the project moves to the recycle bin and
// stays restorable until purge_at. Calling it useDeleteProject while it archives
// is how a Delete button ends up not deleting.
export function useArchiveProject() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => {
      return unwrap(flowApi.archiveProject(id));
    },
    onSuccess: () => {
      void qc.invalidateQueries({ queryKey: ['projects'] });
      void qc.invalidateQueries({ queryKey: ['archived-projects'] });
    },
    onError: () => toast.error(t('toasts.project.archiveFailed')),
  });
}
