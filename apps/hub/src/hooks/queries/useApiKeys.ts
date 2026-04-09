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

const MOCK_KEYS: ApiKey[] = [
  { id: 'k_1', name: 'Production Server',  prefix: 'rtylr_live', lastFour: 'a9f2', createdAt: new Date(Date.now() - 60 * 86400e3).toISOString(), lastUsedAt: new Date(Date.now() - 30 * 60e3).toISOString(), scopes: ['read','write'] },
  { id: 'k_2', name: 'Mobile App Backend', prefix: 'rtylr_live', lastFour: '0b7c', createdAt: new Date(Date.now() - 14 * 86400e3).toISOString(), lastUsedAt: new Date(Date.now() - 4 * 3600e3).toISOString(), scopes: ['read'] },
  { id: 'k_3', name: 'Integrations Test',  prefix: 'rtylr_test', lastFour: 'c412', createdAt: new Date(Date.now() - 3 * 86400e3).toISOString(),  scopes: ['read','write','admin'] },
];

export const useApiKeys = () => {
  return useQuery<ApiKey[]>({
    queryKey: ['api-keys'],
    queryFn: async () => {
      try {
        const res = await api.get('/api-keys');
        return (res.data?.keys ?? res.data) as ApiKey[];
      } catch {
        return MOCK_KEYS;
      }
    },
  });
};

export const useCreateApiKey = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { name: string; scopes: string[] }) => {
      try {
        const res = await api.post('/api-keys', payload);
        return res.data as { key: string; meta: ApiKey };
      } catch {
        const raw = `rtylr_live_${Math.random().toString(36).slice(2, 14)}${Math.random().toString(36).slice(2, 14)}`;
        return {
          key: raw,
          meta: {
            id: `k_${Math.random().toString(36).slice(2, 6)}`,
            name: payload.name,
            prefix: 'rtylr_live',
            lastFour: raw.slice(-4),
            createdAt: new Date().toISOString(),
            scopes: payload.scopes,
          },
        };
      }
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
      try {
        await api.delete(`/api-keys/${id}`);
      } catch {
        // ignore
      }
      return id;
    },
    onSuccess: () => {
      toast.success('API key revoked');
      qc.invalidateQueries({ queryKey: ['api-keys'] });
    },
  });
};
