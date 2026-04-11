import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { hrApi } from '../services/hrApi';

export function useTimeOff() {
  return useQuery({
    queryKey: ['time-off'],
    queryFn: () => unwrap(hrApi.listTimeOff()),
  });
}

export function useSubmitTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { staffId: number; startDate: string; endDate: string; reason: string }) =>
      hrApi.submitTimeOff(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['time-off'] }),
  });
}

export function useApproveTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, approved }: { id: number; approved: boolean }) => unwrap(hrApi.approveTimeOff(id, approved)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['time-off'] }),
  });
}
