import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';

export interface Notification {
  id: string;
  title: string;
  body: string;
  app: 'hub' | 'pos' | 'erp' | 'hr' | 'crm' | 'finance' | 'flow';
  type: 'info' | 'success' | 'warning' | 'error';
  read: boolean;
  createdAt: string;
  href?: string;
}

export const useNotifications = () => {
  return useQuery<Notification[]>({
    queryKey: ['notifications'],
    queryFn: async () => {
      const res = await api.get('/notifications');
      return (res.data?.notifications ?? res.data) as Notification[];
    },
    staleTime: 15_000,
  });
};

export const useMarkNotificationRead = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/notifications/${id}/read`);
      return id;
    },
    onMutate: async (id) => {
      await qc.cancelQueries({ queryKey: ['notifications'] });
      const prev = qc.getQueryData<Notification[]>(['notifications']);
      qc.setQueryData<Notification[]>(['notifications'], (old) =>
        (old ?? []).map((n) => (n.id === id ? { ...n, read: true } : n)),
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
      await api.post('/notifications/read-all');
    },
    onMutate: async () => {
      await qc.cancelQueries({ queryKey: ['notifications'] });
      qc.setQueryData<Notification[]>(['notifications'], (old) =>
        (old ?? []).map((n) => ({ ...n, read: true })),
      );
    },
  });
};
