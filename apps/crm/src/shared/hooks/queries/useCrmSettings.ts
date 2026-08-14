import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { SettingsServiceClient } from '@voxire/voxie/crm/settings/v1/client';
import { settingsSchema } from '../../schemas/settings';
import type { Settings } from '../../schemas/settings';

const settingsClient = createOneKitClient(SettingsServiceClient, 'crm');

export const useCrmSettings = (): UseQueryResult<Settings> =>
  useQuery<Settings>({
    queryKey: ['crm-settings'],
    queryFn: async (): Promise<Settings> => {
      const res = await settingsClient.getSettings({});
      return settingsSchema.parse(res.settings ?? {});
    },
  });

export const useUpdateCrmSettings = (): UseMutationResult<Settings, unknown, Partial<Settings>> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Settings, unknown, Partial<Settings>>({
    mutationFn: async (data: Partial<Settings>): Promise<Settings> => {
      const current = qc.getQueryData<Settings>(['crm-settings']);
      const merged = settingsSchema.parse({ ...current, ...data });
      // UpdateSettingsResponse only returns a confirmation message, not the
      // updated record, so merge the confirmed request into the local cache.
      await settingsClient.updateSettings({
        defaultConsent: merged.defaultConsent,
        fromEmail: merged.fromEmail,
        fromName: merged.fromName,
        smsSenderId: merged.smsSenderId,
        timezone: merged.timezone,
      });
      return merged;
    },
    onSuccess: (settings): void => {
      qc.setQueryData(['crm-settings'], settings);
      toast.success(t('crm.settings.updated'));
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export type { Settings } from '../../schemas/settings';
