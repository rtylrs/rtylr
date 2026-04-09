import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { crmApi, Settings } from '../services/crmApi';

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: () => crmApi.getSettings(),
  });
}

export function useUpdateSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Settings>) => crmApi.updateSettings(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  });
}
