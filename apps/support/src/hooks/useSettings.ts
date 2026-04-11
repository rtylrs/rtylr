import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { supportService } from '../services/supportService';

export function useSupportSettings() {
  return useQuery({
    queryKey: ['support-settings'],
    queryFn: () => supportService.getSettings().then((r) => r.data),
  });
}

export function useUpdateSupportSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => supportService.updateSettings(data).then((r) => r.data),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['support-settings'] });
    },
  });
}
