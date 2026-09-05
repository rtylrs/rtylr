import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { notificationClient } from '../../lib/clients';
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
    nextCursor:
      typeof source.nextCursor === 'string'
        ? source.nextCursor
        : typeof source.next_cursor === 'string'
          ? source.next_cursor
          : '',
  };
};

type PagesData = { pages: NotificationPage[]; pageParams: unknown[] };

type MutationContext = {
  prev: Notification[] | undefined;
  prevPages: [readonly unknown[], PagesData | undefined][];
};

const snapshotPages = (
  qc: ReturnType<typeof useQueryClient>
): [readonly unknown[], PagesData | undefined][] =>
  qc.getQueriesData<PagesData>({ queryKey: ['notifications', 'pages'] });

const restorePages = (
  qc: ReturnType<typeof useQueryClient>,
  prevPages: [readonly unknown[], PagesData | undefined][]
) => {
  for (const [key, data] of prevPages) {
    qc.setQueryData(key, data);
  }
};

export const useNotifications = () => {
  return useQuery<Notification[]>({
    queryKey: ['notifications'],
    queryFn: async () => {
      const res = await notificationClient.listNotifications({});
      const notifications = Array.isArray(res) ? res : (res?.notifications ?? []);
      return z.array(notificationSchema).parse(notifications);
    },
    refetchOnWindowFocus: true,
  });
};

export const useNotificationPages = () => {
  return useInfiniteQuery<NotificationPage>({
    queryKey: ['notifications', 'pages'],
    initialPageParam: '',
    queryFn: async ({ pageParam }) => {
      return parseNotificationPage(
        await notificationClient.listNotifications({
          limit: 100,
          cursor: typeof pageParam === 'string' && pageParam ? pageParam : undefined,
        })
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
      await notificationClient.markNotificationRead({ id });
      return id;
    },
    onMutate: async (id): Promise<MutationContext> => {
      await qc.cancelQueries({ queryKey: ['notifications'] });
      const prev = qc.getQueryData<Notification[]>(['notifications']);
      const prevPages = snapshotPages(qc);
      qc.setQueryData<Notification[]>(['notifications'], (old) =>
        (old ?? []).map((n) => (n.id === id ? { ...n, read: true } : n))
      );
      qc.setQueriesData<PagesData>({ queryKey: ['notifications', 'pages'] }, (old) => {
        if (!old) return old;
        return {
          ...old,
          pages: old.pages.map((page) => ({
            ...page,
            notifications: page.notifications.map((n) => (n.id === id ? { ...n, read: true } : n)),
          })),
        };
      });
      return { prev, prevPages };
    },
    onError: (_err, _id, ctx) => {
      if (ctx?.prev) qc.setQueryData(['notifications'], ctx.prev);
      if (ctx?.prevPages) restorePages(qc, ctx.prevPages);
    },
    onSettled: () => {
      void qc.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
};

export const useMarkAllNotificationsRead = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      await notificationClient.markAllRead({});
    },
    onMutate: async (): Promise<MutationContext> => {
      await qc.cancelQueries({ queryKey: ['notifications'] });
      const prev = qc.getQueryData<Notification[]>(['notifications']);
      const prevPages = snapshotPages(qc);
      qc.setQueryData<Notification[]>(['notifications'], (old) =>
        (old ?? []).map((n) => ({ ...n, read: true }))
      );
      qc.setQueriesData<PagesData>({ queryKey: ['notifications', 'pages'] }, (old) => {
        if (!old) return old;
        return {
          ...old,
          pages: old.pages.map((page) => ({
            ...page,
            notifications: page.notifications.map((n) => ({ ...n, read: true })),
          })),
        };
      });
      return { prev, prevPages };
    },
    onError: (_err, _vars, ctx) => {
      if (ctx?.prev) qc.setQueryData(['notifications'], ctx.prev);
      if (ctx?.prevPages) restorePages(qc, ctx.prevPages);
    },
    onSettled: () => {
      void qc.invalidateQueries({ queryKey: ['notifications'] });
    },
  });
};
