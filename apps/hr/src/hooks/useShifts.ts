import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hrApi, CreateShiftInput, UpdateShiftInput } from '../services/hrApi';

export function useShifts(params?: { day?: string; staffId?: number }) {
  return useQuery({
    queryKey: ['shifts', params],
    queryFn: () => unwrap(hrApi.listShifts(params)),
  });
}

export function useCreateShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: CreateShiftInput) => unwrap(hrApi.createShift(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['shifts'] }),
  });
}

export function useUpdateShift() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: UpdateShiftInput) => unwrap(hrApi.updateShift(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['shifts'] }),
  });
}
