import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { crmApi, Template } from '../services/crmApi';

export function useTemplates() {
  return useQuery({
    queryKey: ['templates'],
    queryFn: () => crmApi.getTemplates(),
  });
}

export function useCreateTemplate() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Template>) => crmApi.createTemplate(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['templates'] }),
  });
}
