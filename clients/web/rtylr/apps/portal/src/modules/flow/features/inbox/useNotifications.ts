import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { flowApi } from '../../shared/services/flowApi';
import type { ApiNotification } from '../../shared/services/flowApi';

const NOTIFICATIONS_PAGE_SIZE = 500;

async function fetchAllNotifications(): Promise<ApiNotification[]> {
  const all: ApiNotification[] = [];
  let page = 1;

  for (;;) {
    const res = await unwrap(flowApi.getNotifications({ page, limit: NOTIFICATIONS_PAGE_SIZE }));
    const rows = res.notifications ?? [];
    all.push(...rows);

    const pagination = res.pagination;
    const hasNext =
      pagination?.hasNext ??
      (typeof pagination?.total === 'number' && pagination.total > all.length);
    if (!hasNext || rows.length === 0) return all;

    page = (pagination?.page ?? page) + 1;
  }
}

export function useNotifications() {
  return useQuery({
    queryKey: ['flow-notifications'],
    queryFn: fetchAllNotifications,
    // Polling removed - the backend pushes `notification.*` events on
    // the caller's `user:<id>` websocket topic; FlowLayout invalidates
    // this query on receipt, so the list stays fresh without a 30s
    // refetch cycle. The shell disables realtime on /settings, though, so
    // mount the inbox with a refetch even when the cached data is fresh.
    staleTime: 60_000,
    refetchOnMount: 'always',
    refetchOnWindowFocus: true,
  });
}

export function useMarkNotificationRead() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(flowApi.markNotificationRead(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['flow-notifications'] }),
    onError: () => toast.error(t('flowToasts.notification.markReadFailed')),
  });
}

export function useMarkAllNotificationsRead() {
  const { t } = useTranslation();
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => unwrap(flowApi.markAllNotificationsRead()),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['flow-notifications'] }),
    onError: () => toast.error(t('flowToasts.notification.markAllReadFailed')),
  });
}
