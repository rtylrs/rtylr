import { unwrap } from '@voxire/api';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { toast } from 'sonner';
import { hrApi } from '../../shared/services/hrApi';
import { timeOffRequestSchema } from '../../shared/schemas/timeOff';

export function useTimeOff() {
  return useQuery({
    queryKey: ['time-off'],
    queryFn: async () => {
      const res = await unwrap(hrApi.listTimeOff());
      const raw = res ?? [];
      return z.array(timeOffRequestSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useSubmitTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: { staffId: number; startDate: string; endDate: string; reason: string }) =>
      unwrap(hrApi.submitTimeOff(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['time-off'] }),
    onError: () => toast.error('Failed to submit time off request'),
  });
}

export function useApproveTimeOff() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, approved }: { id: number; approved: boolean }) =>
      unwrap(hrApi.approveTimeOff(id, approved)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['time-off'] }),
    onError: () => toast.error('Failed to update time off request'),
  });
}
