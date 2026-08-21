import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { developerClient } from '../../lib/clients';
import { webhookSchema } from '../../schemas/webhooks';
import { toast } from 'sonner';
import i18next from 'i18next';

export type { Webhook } from '../../schemas/webhooks';
type Webhook = z.infer<typeof webhookSchema>;

export const useWebhooks = () => {
  return useQuery<Webhook[]>({
    queryKey: ['webhooks'],
    queryFn: async () => {
      const res = await developerClient.listWebhooks({});
      const raw = res.webhooks ?? [];
      return z.array(webhookSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
};

export const useCreateWebhook = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { url: string; events: string[] }) => {
      return developerClient.createWebhook(payload);
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
      await developerClient.deleteWebhook({ id });
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
