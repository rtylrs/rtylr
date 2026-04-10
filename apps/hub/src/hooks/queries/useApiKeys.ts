import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import api from '../../lib/api';
import { toast } from 'sonner';

export interface ApiKey {
  id: string;
  name: string;
  prefix: string;
  lastFour: string;
  createdAt: string;
  lastUsedAt?: string;
  scopes: string[];
}

export const useApiKeys = () => {
  return useQuery<ApiKey[]>({
    queryKey: ['api-keys'],
    queryFn: async () => {
      const res = await api.get('/api-keys');
      return (res.data?.keys ?? res.data) as ApiKey[];
    },
  });
};

export const useCreateApiKey = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { name: string; scopes: string[] }) => {
      const res = await api.post('/api-keys', payload);
      return res.data as { key: string; meta: ApiKey };
    },
    onSuccess: () => {
      toast.success('API key created');
      qc.invalidateQueries({ queryKey: ['api-keys'] });
    },
    onError: () => toast.error('Failed to create API key'),
  });
};

export const useRevokeApiKey = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/api-keys/${id}`);
      return id;
    },
    onSuccess: () => {
      toast.success('API key revoked');
      qc.invalidateQueries({ queryKey: ['api-keys'] });
    },
  });
};
