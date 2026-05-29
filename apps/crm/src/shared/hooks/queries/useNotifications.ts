import { unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { safe } from '../../lib/api';
import type { Notification } from '../../lib/notificationUtils';

export const useNotifications = (): UseQueryResult<Notification[]> =>
  useQuery<Notification[]>({
    queryKey: ['crm-notifications'],
    queryFn: async (): Promise<Notification[]> => {
      const res = await unwrap(safe.get<{ notifications: Notification[] }>('/crm/notifications'));
      return (res as { notifications: Notification[] }).notifications ?? [];
    },
  });

export const useMarkNotificationRead = (): UseMutationResult<unknown, unknown, string> => {
  const qc = useQueryClient();
  return useMutation<unknown, unknown, string>({
    mutationFn: (id: string): Promise<unknown> =>
      unwrap(safe.put('/crm/notifications/read', { id })),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-notifications'] });
    },
  });
};

export const useMarkAllNotificationsRead = (): UseMutationResult<unknown, unknown, void> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, void>({
    mutationFn: (): Promise<unknown> => unwrap(safe.put('/crm/notifications/read-all', {})),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-notifications'] });
      toast.success(t('crm.notifications.allRead'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};
