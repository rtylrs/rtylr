import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';

export interface Notification {
  id: string;
  title: string;
  body: string;
  app: 'hub' | 'pos' | 'erp' | 'hr' | 'crm' | 'finance' | 'tracker';
  type: 'info' | 'success' | 'warning' | 'error';
  read: boolean;
  createdAt: string;
  href?: string;
}

const MOCK_NOTIFS: Notification[] = [
  { id: 'n1', title: 'Welcome to rtylr Hub',         body: 'Take a tour to see everything your Hub can do.',        app: 'hub',     type: 'info',    read: false, createdAt: new Date(Date.now() - 10 * 60e3).toISOString(), href: '/apps' },
  { id: 'n2', title: '5 new orders came in',         body: 'POS registered 5 orders in the last hour.',             app: 'pos',     type: 'success', read: false, createdAt: new Date(Date.now() - 45 * 60e3).toISOString() },
  { id: 'n3', title: 'Low stock alert',              body: 'Tomatoes 5kg is below reorder point.',                  app: 'erp',     type: 'warning', read: false, createdAt: new Date(Date.now() - 2 * 3600e3).toISOString() },
  { id: 'n4', title: 'Timesheet submitted',          body: 'Jorge submitted hours for week of Apr 6.',              app: 'hr',      type: 'info',    read: true,  createdAt: new Date(Date.now() - 6 * 3600e3).toISOString() },
  { id: 'n5', title: 'Invoice overdue',              body: 'INV-2026-0301 is 4 days overdue.',                      app: 'finance', type: 'error',   read: false, createdAt: new Date(Date.now() - 12 * 3600e3).toISOString() },
  { id: 'n6', title: 'New customer signed up',       body: 'Jamie R. created a loyalty account.',                    app: 'crm',     type: 'info',    read: true,  createdAt: new Date(Date.now() - 20 * 3600e3).toISOString() },
  { id: 'n7', title: 'Weekly summary ready',         body: 'Your weekly performance report is available.',          app: 'hub',     type: 'info',    read: true,  createdAt: new Date(Date.now() - 2 * 86400e3).toISOString() },
];

export const useNotifications = () => {
  return useQuery<Notification[]>({
    queryKey: ['notifications'],
    queryFn: async () => {
      try {
        const res = await api.get('/notifications');
        return (res.data?.notifications ?? res.data) as Notification[];
      } catch {
        return MOCK_NOTIFS;
      }
    },
    staleTime: 15_000,
  });
};

export const useMarkNotificationRead = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      try {
        await api.post(`/notifications/${id}/read`);
      } catch {
        // ignore
      }
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
      try {
        await api.post('/notifications/read-all');
      } catch {
        // ignore
      }
    },
    onMutate: async () => {
      await qc.cancelQueries({ queryKey: ['notifications'] });
      qc.setQueryData<Notification[]>(['notifications'], (old) =>
        (old ?? []).map((n) => ({ ...n, read: true })),
      );
    },
  });
};
