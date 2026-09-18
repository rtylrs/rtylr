import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  GetSetupStatusResponse,
  HRCompanyConfigPayload,
} from '@voxire/voxie/hr/setup/v1/types';
import { hrTransport } from '../../lib/clients';
import { useHRConfigStore } from '../../stores/hrConfigStore';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

export const useGetSetupStatus = () => {
  return useQuery<GetSetupStatusResponse>({
    queryKey: ['hr-setup-status'],
    queryFn: async (): Promise<GetSetupStatusResponse> => {
      const res = await unwrap(hrTransport.get<GetSetupStatusResponse>('/hr/setup/status'));
      return res as GetSetupStatusResponse;
    },
  });
};

export type CompleteStepInput = {
  stepKey: string;
  data?: Record<string, unknown>;
};

export const useCompleteSetupStep = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async ({ stepKey, data }: CompleteStepInput): Promise<void> => {
      await unwrap(hrTransport.post(`/hr/setup/steps/${stepKey}/complete`, data ?? {}));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr-setup-status'] });
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useGetHRConfig = () => {
  return useQuery<HRCompanyConfigPayload>({
    queryKey: ['hr-config'],
    queryFn: async (): Promise<HRCompanyConfigPayload> => {
      const res = await unwrap(hrTransport.get<{ config: HRCompanyConfigPayload }>('/hr/config'));
      return (res as { config: HRCompanyConfigPayload }).config;
    },
  });
};

export const useUpsertHRConfig = () => {
  const queryClient = useQueryClient();
  const setConfig = useHRConfigStore((s) => s.setConfig);
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: Partial<HRCompanyConfigPayload>): Promise<HRCompanyConfigPayload> => {
      const res = await unwrap(
        hrTransport.put<{ config: HRCompanyConfigPayload }>('/hr/config', data)
      );
      return (res as { config: HRCompanyConfigPayload }).config;
    },
    onSuccess: (config: HRCompanyConfigPayload) => {
      setConfig(config);
      queryClient.setQueryData(['hr-config'], config);
      queryClient.invalidateQueries({ queryKey: ['hr-config'] });
      toast.success(t('setup.configSaved'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
