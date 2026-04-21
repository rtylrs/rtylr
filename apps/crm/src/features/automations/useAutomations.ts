import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { crmApi, Automation } from '../../shared/services/crmApi';
import { automationSchema } from '../../shared/schemas/automations';
import { unwrapList } from '../../shared/lib/unwrapList';

export function useAutomations() {
  return useQuery({
    queryKey: ['automations'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getAutomations());
      return z.array(automationSchema).parse(unwrapList(res, 'automations'));
    },
  });
}

export function useCreateAutomation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Automation>) => unwrap(crmApi.createAutomation(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['automations'] }),
  });
}

export function useUpdateAutomation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Automation> }) =>
      unwrap(crmApi.updateAutomation(id, data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['automations'] }),
  });
}

export function useDeleteAutomation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => unwrap(crmApi.deleteAutomation(id)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['automations'] }),
  });
}
