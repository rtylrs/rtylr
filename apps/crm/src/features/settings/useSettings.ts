import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { crmApi, Settings } from '../../shared/services/crmApi';
import { settingsSchema } from '../../shared/schemas/settings';

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: async () => {
      const res = await unwrap(crmApi.getSettings());
      return settingsSchema.parse(res);
    },
  });
}

export function useUpdateSettings() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Settings>) => unwrap(crmApi.updateSettings(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['settings'] }),
  });
}
