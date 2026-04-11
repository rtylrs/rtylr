import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';

export interface Webhook {
  id: string;
  url: string;
  events: string[];
  status: 'active' | 'paused' | 'failing';
  lastDelivery?: string;
  secretPreview?: string;
}

export const useWebhooks = () => {
  return useQuery<Webhook[]>({
    queryKey: ['webhooks'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/webhooks'));
      return (res?.webhooks ?? res) as Webhook[];
    },
  });
};

export const useCreateWebhook = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { url: string; events: string[] }) => {
      const res = await unwrap(safe.post('/webhooks', payload));
      return res as Webhook;
    },
    onSuccess: () => {
      toast.success('Webhook created');
      qc.invalidateQueries({ queryKey: ['webhooks'] });
    },
  });
};

export const useDeleteWebhook = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await unwrap(safe.delete(`/webhooks/${id}`));
      return id;
    },
    onSuccess: () => {
      toast.success('Webhook removed');
      qc.invalidateQueries({ queryKey: ['webhooks'] });
    },
  });
};
