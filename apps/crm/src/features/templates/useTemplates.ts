import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { crmApi, Template } from '../../shared/services/crmApi';
import { templateSchema } from '../../shared/schemas/templates';

export function useTemplates() {
  return useQuery({
    queryKey: ['templates'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getTemplates());
      const raw = res ?? [];
      return z.array(templateSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Template>) => unwrap(crmApi.createTemplate(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['templates'] }),
  });
}
