import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { toast } from 'sonner';
import { hrApi, CreateShiftInput, UpdateShiftInput } from '../../shared/services/hrApi';
import { shiftSchema } from '../../shared/schemas/shifts';

export function useShifts(params?: { day?: string; staffId?: number }) {
  return useQuery({
    queryKey: ['shifts', params],
    queryFn: async () => {
      const res = await unwrap(hrApi.listShifts(params));
      const raw = res ?? [];
      return z.array(shiftSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCreateShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateShiftInput) => unwrap(hrApi.createShift(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['shifts'] }),
    onError: () => toast.error('Failed to create shift'),
  });
}

export function useUpdateShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateShiftInput) => unwrap(hrApi.updateShift(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['shifts'] }),
    onError: () => toast.error('Failed to update shift'),
  });
}
