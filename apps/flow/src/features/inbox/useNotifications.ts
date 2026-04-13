import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { flowApi } from '../../shared/services/flowApi';

export function useNotifications() {
  return useQuery({
    queryKey: ['notifications'],
    queryFn: async () => {
      const res = await unwrap(flowApi.getNotifications());
      return res.notifications ?? [];
    },
    refetchInterval: 30_000, // poll every 30s
  });
}

export function useMarkNotificationRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => unwrap(flowApi.markNotificationRead(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notifications'] }),
  });
}

export function useMarkAllNotificationsRead() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: () => unwrap(flowApi.markAllNotificationsRead()),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['notifications'] }),
  });
}
