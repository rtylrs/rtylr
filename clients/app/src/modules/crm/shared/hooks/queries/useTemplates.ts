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
import { z } from 'zod';
import { TemplateServiceClient } from '@voxire/voxie/crm/template/v1/client';
import type { Template as ApiTemplate } from '@voxire/voxie/crm/template/v1/types';
import { templateSchema } from '../../schemas/templates';
import type { Template } from '../../schemas/templates';
import { unwrapList } from '../../lib/unwrapList';

const templateClient = createOneKitClient(TemplateServiceClient, 'crm');

// crm.template.v1.Template uses `type` for what this app's UI calls `channel`,
// and has no category/thumbnail fields (UI-only concepts, don't persist).
function fromApiTemplate(t: ApiTemplate): Template {
  return templateSchema.parse({
    id: t.id,
    name: t.name,
    channel: t.type,
    subject: t.subject,
    body: t.body,
    updatedAt: t.updatedAt,
  });
}

export const useTemplates = (): UseQueryResult<Template[]> =>
  useQuery<Template[]>({
    queryKey: ['crm-templates'],
    queryFn: async (): Promise<Template[]> => {
      const res = await templateClient.listTemplates({});
      return z
        .array(templateSchema)
        .parse(unwrapList<ApiTemplate>(res, 'templates').map(fromApiTemplate));
    },
  });

export const useCreateTemplate = (): UseMutationResult<Template, unknown, Partial<Template>> => {
  const qc = useQueryClient();
  const { t } = useTranslation();
  return useMutation<Template, unknown, Partial<Template>>({
    mutationFn: async (data: Partial<Template>): Promise<Template> => {
      const res = await templateClient.createTemplate({
        name: data.name ?? '',
        type: data.channel ?? 'email',
        subject: data.subject ?? '',
        body: data.body ?? '',
        isActive: true,
      });
      return templateSchema.parse({
        id: res.templateId,
        name: data.name ?? '',
        channel: data.channel ?? 'email',
        category: data.category,
        subject: data.subject,
        body: data.body ?? '',
        thumbnail: data.thumbnail,
        updatedAt: new Date().toISOString(),
      });
    },
    onSuccess: (): void => {
      void qc.invalidateQueries({ queryKey: ['crm-templates'] });
      toast.success(t('crm.templates.created'));
    },
    onError: (error): void => {
      toast.error(getOneKitErrorMessage(error, t('crmErrors.generic')));
    },
  });
};

export type { Template } from '../../schemas/templates';
