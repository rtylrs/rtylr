import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { apiKeySchema } from '../../schemas/apiKeys';
import { toast } from 'sonner';

export type { ApiKey } from '../../schemas/apiKeys';
type ApiKey = z.infer<typeof apiKeySchema>;

export const useApiKeys = () => {
  return useQuery<ApiKey[]>({
    queryKey: ['api-keys'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/hub/developer/v1/api-keys'));
      const raw = res?.keys ?? res;
      return z.array(apiKeySchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
};

export const useCreateApiKey = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { name: string; scopes: string[] }) => {
      const res = await unwrap(safe.post('/hub/developer/v1/api-keys', payload));
      return res as { id: number; key: string; message: string };
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
      await unwrap(safe.post('/hub/developer/v1/api-keys/revoke', { id: Number(id) }));
      return id;
    },
    onSuccess: () => {
      toast.success('API key revoked');
      qc.invalidateQueries({ queryKey: ['api-keys'] });
    },
    onError: () => toast.error('Failed to revoke API key'),
  });
};
