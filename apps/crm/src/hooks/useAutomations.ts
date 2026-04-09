import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { crmApi, Automation } from '../services/crmApi';

export function useAutomations() {
  return useQuery({
    queryKey: ['automations'],
    queryFn: () => crmApi.getAutomations(),
  });
}

export function useCreateAutomation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Automation>) => crmApi.createAutomation(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['automations'] }),
  });
}

export function useUpdateAutomation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Automation> }) =>
      crmApi.updateAutomation(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['automations'] }),
  });
}
