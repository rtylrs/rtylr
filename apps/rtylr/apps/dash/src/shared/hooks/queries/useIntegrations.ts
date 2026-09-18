import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { integrationClient } from '../../lib/clients';
import { integrationSchema } from '../../schemas/integrations';
import { toast } from 'sonner';
import i18next from 'i18next';

export type { Integration } from '../../schemas/integrations';

export const useIntegrations = () => {
  return useQuery({
    queryKey: ['integrations'],
    queryFn: async () => {
      const res = await integrationClient.listIntegrations({});
      const raw = res.integrations ?? [];
      return z.array(integrationSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
};

export const useConnectIntegration = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await integrationClient.connectIntegration({ provider: id });
      return id;
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.integrationConnected', 'Integration connected'));
      qc.invalidateQueries({ queryKey: ['integrations'] });
    },
    onError: () =>
      toast.error(
        i18next.t('common.toasts.failedConnectIntegration', 'Failed to connect integration')
      ),
  });
};

export const useDisconnectIntegration = () => {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async (id: string) => {
      await integrationClient.disconnectIntegration({ id });
      return id;
    },
    onSuccess: () => {
      toast.success(i18next.t('common.toasts.integrationDisconnected', 'Integration disconnected'));
      qc.invalidateQueries({ queryKey: ['integrations'] });
    },
    onError: () =>
      toast.error(
        i18next.t('common.toasts.failedDisconnectIntegration', 'Failed to disconnect integration')
      ),
  });
};
