import { unwrap } from '@voxire/api';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { projectV2 } from '../../shared/lib/clients';

export function useRestoreProject() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async (id: string) => {
      await projectV2.restoreProject({ id });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['archived-projects'] });
      qc.invalidateQueries({ queryKey: ['projects-v2'] });
      toast.success(t('flow.bin.restored'));
    },
    onError: () => toast.error(t('toasts.project.restoreFailed')),
  });
}

export function usePurgeProject() {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation({
    mutationFn: async (id: string) => {
      await projectV2.purgeProject({ id });
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['archived-projects'] });
      toast.success(t('flow.bin.purged'));
    },
    onError: () => toast.error(t('toasts.project.purgeFailed')),
  });
}
