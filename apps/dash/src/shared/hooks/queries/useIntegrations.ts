import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { safe } from '../../lib/api';
import { integrationSchema } from '../../schemas/integrations';
import { toast } from 'sonner';
import i18next from 'i18next';

export type { Integration } from '../../schemas/integrations';

export const useIntegrations = () => {
  return useQuery({
    queryKey: ['integrations'],
    queryFn: async () => {
      const res = await unwrap(safe.get('/hub/integration/v1/integrations'));
      const raw = res?.integrations ?? res;
      return z.array(integrationSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
};

export const useConnectIntegration = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await unwrap(safe.post('/hub/integration/v1/integrations/connect', { provider: id }));
      return id;
    },
    onSuccess: () => {
      toast.success(i18next.t('toasts.integrationConnected', 'Integration connected'));
      qc.invalidateQueries({ queryKey: ['integrations'] });
    },
    onError: () =>
      toast.error(i18next.t('toasts.failedConnectIntegration', 'Failed to connect integration')),
  });
};

export const useDisconnectIntegration = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await unwrap(safe.post('/hub/integration/v1/integrations/disconnect', { id: Number(id) }));
      return id;
    },
    onSuccess: () => {
      toast.success(i18next.t('toasts.integrationDisconnected', 'Integration disconnected'));
      qc.invalidateQueries({ queryKey: ['integrations'] });
    },
    onError: () =>
      toast.error(
        i18next.t('toasts.failedDisconnectIntegration', 'Failed to disconnect integration')
      ),
  });
};
