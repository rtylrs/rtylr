import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { supportService, SupportSettingsPayload } from '../../shared/services/supportService';
import { supportSettingsSchema } from '../../shared/schemas/settings';

export function useSupportSettings() {
  return useQuery({
    queryKey: ['support-settings'],
    queryFn: async () => {
      const res = await unwrap(supportService.getSettings());
      return supportSettingsSchema.parse(res);
    },
  });
}

export function useUpdateSupportSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: SupportSettingsPayload) => unwrap(supportService.updateSettings(data)),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['support-settings'] });
    },
  });
}
