import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { developerClient } from '../../lib/clients';
import { isSupportedWebhookEvent, normalizeWebhookUrl } from '../../lib/webhookUrl';
import { toast } from 'sonner';
import i18next from 'i18next';

const webhookSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  url: z.string(),
  events: z.array(z.string()),
  // The Go contract omits false, so an absent boolean means the webhook is inactive.
  isActive: z.boolean().default(false),
  createdAt: z.string().optional(),
  lastTriggeredAt: z.string().optional(),
});

export type Webhook = z.infer<typeof webhookSchema>;

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
      const url = normalizeWebhookUrl(payload.url);
      if (!url) throw new Error('Webhook URL must be a valid HTTP(S) URL');
      if (payload.events.length === 0) throw new Error('At least one webhook event is required');
      if (payload.events.some((event) => !isSupportedWebhookEvent(event))) {
        throw new Error('Unsupported webhook event');
      }
      return developerClient.createWebhook({ url, events: payload.events });
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.webhookCreated', 'Webhook created'));
      qc.invalidateQueries({ queryKey: ['webhooks'] });
    },
    onError: () => {
      toast.error(i18next.t('common.toasts.failedCreateWebhook', 'Failed to create webhook'));
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
      toast.success(i18next.t('common.toasts.webhookRemoved', 'Webhook removed'));
      qc.invalidateQueries({ queryKey: ['webhooks'] });
    },
    onError: () => {
      toast.error(i18next.t('common.toasts.failedRemoveWebhook', 'Failed to remove webhook'));
    },
  });
};
