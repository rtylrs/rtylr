import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';
import { toast } from 'sonner';

export interface Integration {
  id: string;
  name: string;
  category: 'payments' | 'messaging' | 'accounting' | 'analytics' | 'delivery' | 'data';
  logo: string;
  status: 'connected' | 'available' | 'coming_soon';
  description: string;
  connectedAt?: string;
}

export const useIntegrations = () => {
  return useQuery<Integration[]>({
    queryKey: ['integrations'],
    queryFn: async () => {
      const res = await api.get('/integrations');
      return (res.data?.integrations ?? res.data) as Integration[];
    },
  });
};

export const useConnectIntegration = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/integrations/${id}/connect`);
      return id;
    },
    onSuccess: () => {
      toast.success('Integration connected');
      qc.invalidateQueries({ queryKey: ['integrations'] });
    },
  });
};

export const useDisconnectIntegration = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await api.post(`/integrations/${id}/disconnect`);
      return id;
    },
    onSuccess: () => {
      toast.success('Integration disconnected');
      qc.invalidateQueries({ queryKey: ['integrations'] });
    },
  });
};
