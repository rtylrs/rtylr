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
import { templateSchema } from '../../schemas/templates';
import type { Template } from '../../schemas/templates';

export const useTemplates = (): UseQueryResult<Template[]> =>
  useQuery<Template[]>({
    queryKey: ['crm-templates'],
    queryFn: async (): Promise<Template[]> => {
      const res = await unwrap(safe.get<unknown>('/crm/templates'));
      return z.array(templateSchema).parse(unwrapList(res, 'templates'));
    },
  });

export const useCreateTemplate = (): UseMutationResult<Template, unknown, Partial<Template>> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Template, unknown, Partial<Template>>({
    mutationFn: async (data: Partial<Template>): Promise<Template> => {
      const res = await unwrap(safe.post<Template>('/crm/templates', data));
      return res as Template;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-templates'] });
      toast.success(t('crm.templates.created'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useUpdateTemplate = (): UseMutationResult<
  Template,
  unknown,
  { id: number; data: Partial<Template> }
> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Template, unknown, { id: number; data: Partial<Template> }>({
    mutationFn: async (payload): Promise<Template> => {
      const res = await unwrap(
        safe.put<Template>('/crm/templates', { id: payload.id, ...payload.data })
      );
      return res as Template;
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-templates'] });
      toast.success(t('crm.templates.updated'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useDeleteTemplate = (): UseMutationResult<unknown, unknown, number> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<unknown, unknown, number>({
    mutationFn: (id: number): Promise<unknown> =>
      unwrap(safe.delete('/crm/templates', { params: { id } })),
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-templates'] });
      toast.success(t('crm.templates.deleted'));
    },
    onError: (): void => {
      toast.error(t('errors.generic'));
    },
  });
};

export type { Template } from '../../schemas/templates';
