import { unwrap } from '@voxire/api';
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { notificationSchema } from '../../schemas/notifications';

export type { Notification } from '../../schemas/notifications';
type Notification = z.infer<typeof notificationSchema>;

type NotificationPage = {
  notifications: Notification[];
  nextCursor: string;
};

const parseNotificationPage = (res: unknown): NotificationPage => {
  const source = res && typeof res === 'object' ? (res as Record<string, unknown>) : {};
  return {
    notifications: z.array(notificationSchema).parse(source.notifications ?? []),
    nextCursor: typeof source.next_cursor === 'string' ? source.next_cursor : '',
  };
};

export const useNotifications = () => {
  return useQuery<Notification[]>({
    queryKey: ['notifications'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/hub/notification/v1/notifications'));
      const raw = res?.notifications ?? res;
      return z.array(notificationSchema).parse(Array.isArray(raw) ? raw : []);
    },
    refetchOnWindowFocus: true,
  });
};

export const useNotificationPages = () => {
  return useInfiniteQuery<NotificationPage>({
    queryKey: ['notifications', 'pages'],
    initialPageParam: '',
    queryFn: async ({ pageParam }) => {
      const params = new URLSearchParams({ limit: '100' });
      if (typeof pageParam === 'string' && pageParam) params.set('cursor', pageParam);
      return parseNotificationPage(
        await unwrap(safe.get(`/hub/notification/v1/notifications?${params.toString()}`))
      );
    },
    getNextPageParam: (lastPage) => lastPage.nextCursor || undefined,
    refetchOnWindowFocus: true,
  });
};

export const useMarkNotificationRead = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await unwrap(safe.patch(`/hub/notification/v1/notifications/${encodeURIComponent(id)}/read`));
      return id;
    },
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: ['notifications'] });
      const prev = qc.getQueryData<Notification[]>(['notifications']);
      qc.setQueryData<Notification[]>(['notifications'], (old) =>
        (old ?? []).map((n) => (n.id === id ? { ...n, read: true } : n))
      );
      return { prev };
    },
    onError: (_err, _id, ctx) => {
      if (ctx?.prev) qc.setQueryData(['notifications'], ctx.prev);
    },
  });
};

export const useMarkAllNotificationsRead = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      await unwrap(safe.patch('/hub/notification/v1/notifications/read-all'));
    },
    onMutate: async () => {
      await qc.cancelQueries({ queryKey: ['notifications'] });
      const prev = qc.getQueryData<Notification[]>(['notifications']);
      qc.setQueryData<Notification[]>(['notifications'], (old) =>
        (old ?? []).map((n) => ({ ...n, read: true }))
      );
      return { prev };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) qc.setQueryData(['notifications'], ctx.prev);
    },
    onSettled: () => {
      void qc.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
};
