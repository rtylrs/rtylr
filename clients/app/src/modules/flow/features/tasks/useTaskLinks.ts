import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import type { TaskLink } from '@voxire/voxie/flow/task/v2/types';
import { taskV2 } from '../../shared/lib/clients';

export type { TaskLink };

export function useTaskLinks(taskId: string | undefined) {
  return useQuery<TaskLink[]>({
    queryKey: ['task-links', taskId],
    queryFn: async () => {
      const res = await taskV2.listTaskLinks({ taskId: taskId! });
      return (res as { links?: TaskLink[] }).links ?? [];
    },
    enabled: !!taskId,
  });
}

export function useAddTaskLink() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async (vars: { taskId: string; url: string }) => {
      const res = await taskV2.createTaskLink({ taskId: vars.taskId, url: vars.url });
      return (res as { link?: TaskLink }).link;
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['task-links', vars.taskId] });
      qc.invalidateQueries({ queryKey: ['task', vars.taskId] });
    },
    onError: () => toast.error(t('flowToasts.link.addFailed')),
  });
}

export function useRemoveTaskLink() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async (vars: { taskId: string; linkId: string }) => {
      await taskV2.deleteTaskLink({ id: vars.linkId });
    },
    onSuccess: (_data, vars) => {
      qc.invalidateQueries({ queryKey: ['task-links', vars.taskId] });
      qc.invalidateQueries({ queryKey: ['task', vars.taskId] });
    },
    onError: () => toast.error(t('flowToasts.link.removeFailed')),
  });
}

/** Mirrors the server's rule so an obviously bad paste fails before a round trip. */
export function isWebUrl(value: string): boolean {
  try {
    const url = new URL(value.trim());
    return (url.protocol === 'http:' || url.protocol === 'https:') && url.hostname !== '';
  } catch {
    return false;
  }
}
