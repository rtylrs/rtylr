import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiProject, ProjectUpdate } from '../../shared/services/flowApi';
import { unwrapListResponse } from '../../shared/lib/unwrapListResponse';

export function useProjects({ enabled = true } = {}) {
  return useQuery<ApiProject[]>({
    queryKey: ['projects'],
    enabled,
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
    mutationFn: (data: {
      name: string;
      color?: string;
      description?: string;
      visibility?: string;
    }) => {
      return unwrap(flowApi.createProject(data));
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['projects'] });
      qc.invalidateQueries({ queryKey: ['projects-v2'] });
    },
    onError: () => toast.error(t('flowToasts.project.createFailed')),
  });
}

export function useUpdateProject() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: string; data: ProjectUpdate }) => {
      return unwrap(flowApi.updateProject(id, data)).then(() => {});
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['projects'] });
      qc.invalidateQueries({ queryKey: ['projects-v2'] });
    },
    onError: () => toast.error(t('flowToasts.project.updateFailed')),
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
    onSuccess: (res) => {
      void qc.invalidateQueries({ queryKey: ['projects'] });
      void qc.invalidateQueries({ queryKey: ['projects-v2'] });
      void qc.invalidateQueries({ queryKey: ['archived-projects'] });

      // Ruling P3: state the scheduled permanent-deletion date explicitly. The
      // date is only knowable after the archive, because the server stamps
      // purge_at from the business's retention setting, so the confirmation is
      // where it belongs. Falls back to the undated message if the server did
      // not send one, rather than printing "undefined".
      const purgeAt = res?.project?.purgeAt;
      if (purgeAt) {
        toast.success(t('flowToasts.project.archivedUntil', { date: purgeAt.slice(0, 10) }));
      } else {
        toast.success(t('flowToasts.project.archived'));
      }
    },
    onError: () => toast.error(t('flowToasts.project.archiveFailed')),
  });
}
