import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { developerClient } from '../../lib/clients';
import { apiKeySchema } from '../../schemas/apiKeys';
import { toast } from 'sonner';
import i18next from 'i18next';

export type { ApiKey } from '../../schemas/apiKeys';
type ApiKey = z.infer<typeof apiKeySchema>;

export const useApiKeys = () => {
  return useQuery<ApiKey[]>({
    queryKey: ['api-keys'],
    queryFn: async () => {
      const res = await developerClient.listApiKeys({});
      const raw = res.keys ?? [];
      return z.array(apiKeySchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
};

export const useCreateApiKey = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (payload: { name: string; scopes: string[] }) => {
      return developerClient.createApiKey(payload);
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.apiKeyCreated', 'API key created'));
      qc.invalidateQueries({ queryKey: ['api-keys'] });
    },
    onError: () =>
      toast.error(i18next.t('common.toasts.failedCreateApiKey', 'Failed to create API key')),
  });
};

export const useRevokeApiKey = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await developerClient.revokeApiKey({ id });
      return id;
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.apiKeyRevoked', 'API key revoked'));
      qc.invalidateQueries({ queryKey: ['api-keys'] });
    },
    onError: () =>
      toast.error(i18next.t('common.toasts.failedRevokeApiKey', 'Failed to revoke API key')),
  });
};
