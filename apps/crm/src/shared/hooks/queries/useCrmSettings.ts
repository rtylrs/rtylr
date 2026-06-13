import { unwrap } from '@voxire/api';
import {
  useMutation,
  useQuery,
  useQueryClient,
  type UseMutationResult,
  type UseQueryResult,
} from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { safe } from '../../lib/api';
import { settingsSchema } from '../../schemas/settings';
import type { Settings } from '../../schemas/settings';

export const useCrmSettings = (): UseQueryResult<Settings> =>
  useQuery<Settings>({
    queryKey: ['crm-settings'],
    queryFn: async (): Promise<Settings> => {
      const res = await unwrap(safe.get<unknown>('/crm/settings'));
      const envelope = res as { settings?: unknown };
      return settingsSchema.parse(envelope?.settings ?? res);
    },
  });

export const useUpdateCrmSettings = (): UseMutationResult<Settings, unknown, Partial<Settings>> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Settings, unknown, Partial<Settings>>({
    mutationFn: async (data: Partial<Settings>): Promise<Settings> => {
      const res = await unwrap(safe.put<Settings>('/crm/settings', data));
      return res as Settings;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-settings'] });
      toast.success(t('crm.settings.updated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export type { Settings } from '../../schemas/settings';
