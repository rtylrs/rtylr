import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hrApi, CreateShiftInput, UpdateShiftInput } from '../services/hrApi';

export function useShifts(params?: { day?: string; staffId?: number }) {
  return useQuery({
    queryKey: ['shifts', params],
    queryFn: () => hrApi.listShifts(params).then((r) => r.data || []),
  });
}

export function useCreateShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateShiftInput) => hrApi.createShift(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['shifts'] }),
  });
}

export function useUpdateShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateShiftInput) => hrApi.updateShift(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['shifts'] }),
  });
}
