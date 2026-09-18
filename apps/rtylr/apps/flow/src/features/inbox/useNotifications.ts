import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';

export function useNotifications() {
  return useQuery({
    queryKey: ['notifications'],
    queryFn: async () => {
      const res = await unwrap(flowApi.getNotifications());
      return res.notifications ?? [];
    },
    // Polling removed - the backend pushes `notification.*` events on
    // the caller's `user:<id>` websocket topic; FlowLayout invalidates
    // this query on receipt, so the list stays fresh without a 30s
    // refetch cycle. Kept a long `staleTime` as a backstop - if the
    // socket is down for a while, a tab refocus still re-fetches.
    staleTime: 60_000,
  });
}

export function useMarkNotificationRead() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(flowApi.markNotificationRead(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notifications'] }),
    onError: () => toast.error(t('toasts.notification.markReadFailed')),
  });
}

export function useMarkAllNotificationsRead() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => unwrap(flowApi.markAllNotificationsRead()),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notifications'] }),
    onError: () => toast.error(t('toasts.notification.markAllReadFailed')),
  });
}
