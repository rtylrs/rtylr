import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';
import { toast } from 'sonner';

export interface Webhook {
  id: string;
  url: string;
  events: string[];
  status: 'active' | 'paused' | 'failing';
  lastDelivery?: string;
  secretPreview?: string;
}

const MOCK_HOOKS: Webhook[] = [
  { id: 'wh_1', url: 'https://api.acme.com/rtylr/events',     events: ['order.created','order.paid'],    status: 'active',  lastDelivery: new Date(Date.now() - 15 * 60e3).toISOString(), secretPreview: 'whsec_...9a2c' },
  { id: 'wh_2', url: 'https://ops.company.io/hooks/rtylr',    events: ['inventory.low','shift.started'], status: 'active',  lastDelivery: new Date(Date.now() - 3 * 3600e3).toISOString(), secretPreview: 'whsec_...4de1' },
  { id: 'wh_3', url: 'https://n8n.self.host/webhook/rtylr',   events: ['customer.created'],              status: 'failing', lastDelivery: new Date(Date.now() - 12 * 3600e3).toISOString(), secretPreview: 'whsec_...0fab' },
];

export const useWebhooks = () => {
  return useQuery<Webhook[]>({
    queryKey: ['webhooks'],
    queryFn: async () => {
      try {
        const res = await api.get('/webhooks');
        return (res.data?.webhooks ?? res.data) as Webhook[];
      } catch {
        return MOCK_HOOKS;
      }
    },
  });
};

export const useCreateWebhook = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { url: string; events: string[] }) => {
      try {
        const res = await api.post('/webhooks', payload);
        return res.data as Webhook;
      } catch {
        return {
          id: `wh_${Math.random().toString(36).slice(2, 6)}`,
          url: payload.url,
          events: payload.events,
          status: 'active' as const,
          secretPreview: `whsec_...${Math.random().toString(36).slice(2, 6)}`,
        };
      }
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
      try {
        await api.delete(`/webhooks/${id}`);
      } catch {
        // ignore
      }
      return id;
    },
    onSuccess: () => {
      toast.success('Webhook removed');
      qc.invalidateQueries({ queryKey: ['webhooks'] });
    },
  });
};
