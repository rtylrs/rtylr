import { unwrap } from '@voxire/api';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { toast } from 'sonner';
import { hrApi, CreateShiftInput } from '../../shared/services/hrApi';
import { shiftSchema } from '../../shared/schemas/shifts';

export function useShifts(params?: { day?: string; staffId?: number }) {
  return useQuery({
    queryKey: ['hr', 'shifts', params],
    queryFn: async () => {
      const res = await unwrap(hrApi.listShifts(params));
      const raw = res ?? [];
      return z.array(shiftSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateShift() {
  const { t } = useTranslation('hrCommon');
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateShiftInput) => unwrap(hrApi.createShift(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['hr', 'shifts'] }),
    onError: () => toast.error(t('hrErrors.createShift')),
  });
}
