import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { supportService } from '../services/supportService';

export function useSupportSettings() {
  return useQuery({
    queryKey: ['support-settings'],
    queryFn: () => unwrap(supportService.getSettings()),
  });
}

export function useUpdateSupportSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: any) => unwrap(supportService.updateSettings(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['support-settings'] });
    },
  });
}
