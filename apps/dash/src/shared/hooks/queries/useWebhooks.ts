import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { webhookSchema } from '../../schemas/webhooks';
import { toast } from 'sonner';
import i18next from 'i18next';

export type { Webhook } from '../../schemas/webhooks';
type Webhook = z.infer<typeof webhookSchema>;

export const useWebhooks = () => {
  return useQuery<Webhook[]>({
    queryKey: ['webhooks'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/hub/developer/v1/webhooks'));
      const raw = res?.webhooks ?? res;
      return z.array(webhookSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
};

export const useCreateWebhook = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { url: string; events: string[] }) => {
      const res = await unwrap(safe.post('/hub/developer/v1/webhooks', payload));
      return res as { id: number; secret?: string; message?: string };
    },
    onSuccess: () => {
      toast.success(i18next.t('toasts.webhookCreated', 'Webhook created'));
      qc.invalidateQueries({ queryKey: ['webhooks'] });
    },
    onError: () => {
      toast.error(i18next.t('toasts.failedCreateWebhook', 'Failed to create webhook'));
    },
  });
};

export const useDeleteWebhook = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await unwrap(safe.delete(`/hub/developer/v1/webhooks/${encodeURIComponent(id)}`));
      return id;
    },
    onSuccess: () => {
      toast.success(i18next.t('toasts.webhookRemoved', 'Webhook removed'));
      qc.invalidateQueries({ queryKey: ['webhooks'] });
    },
    onError: () => {
      toast.error(i18next.t('toasts.failedRemoveWebhook', 'Failed to remove webhook'));
    },
  });
};
