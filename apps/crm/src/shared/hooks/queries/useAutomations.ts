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
import { z } from 'zod';
import { safe } from '../../lib/api';
import { unwrapList } from '../../lib/unwrapList';
import { automationSchema } from '../../schemas/automations';
import type { Automation } from '../../schemas/automations';

interface UpdateAutomationPayload {
  id: number;
  data: Partial<Automation>;
}

export const useListAutomations = (): UseQueryResult<Automation[]> =>
  useQuery<Automation[]>({
    queryKey: ['crm-automations'],
    queryFn: async (): Promise<Automation[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/automations'));
      return z.array(automationSchema).parse(unwrapList(res, 'automations'));
    },
  });

export const useCreateAutomation = (): UseMutationResult<
  Automation,
  unknown,
  Partial<Automation>
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Automation, unknown, Partial<Automation>>({
    mutationFn: async (data: Partial<Automation>): Promise<Automation> => {
      const res = await unwrap(safe.post<Automation>('/crm/automations', data));
      return res as Automation;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-automations'] });
      toast.success(t('crm.automations.created'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateAutomation = (): UseMutationResult<
  Automation,
  unknown,
  UpdateAutomationPayload
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Automation, unknown, UpdateAutomationPayload>({
    mutationFn: async (payload: UpdateAutomationPayload): Promise<Automation> => {
      const res = await unwrap(
        safe.put<Automation>(`/crm/automations/${payload.id}`, payload.data)
      );
      return res as Automation;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-automations'] });
      toast.success(t('crm.automations.updated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useDeleteAutomation = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> => unwrap(safe.delete(`/crm/automations/${id}`)),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-automations'] });
      toast.success(t('crm.automations.deleted'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export type { Automation } from '../../schemas/automations';
