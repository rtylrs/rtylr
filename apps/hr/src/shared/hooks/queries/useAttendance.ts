import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  ListAttendanceRequest,
  ListAttendanceResponse,
  AttendanceClockInRequest,
  AttendanceClockInResponse,
  AttendanceClockOutRequest,
  AttendanceClockOutResponse,
} from '@voxire/contracts/hr/v1/attendance';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

export const useGetMyAttendance = (params: ListAttendanceRequest) => {
  return useQuery<ListAttendanceResponse>({
    queryKey: ['attendance', 'filter', params],
    queryFn: async (): Promise<ListAttendanceResponse> => {
      const res = await unwrap(safe.get<ListAttendanceResponse>('/hr/attendance', { params }));
      return res as ListAttendanceResponse;
    },
  });
};

export const useClockIn = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: AttendanceClockInRequest): Promise<AttendanceClockInResponse> => {
      const res = await unwrap(
        safe.post<AttendanceClockInResponse>('/hr/attendance/clock-in', data)
      );
      return res as AttendanceClockInResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['attendance'] });
      toast.success(t('attendance.clockedIn'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useClockOut = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: AttendanceClockOutRequest): Promise<AttendanceClockOutResponse> => {
      const res = await unwrap(
        safe.post<AttendanceClockOutResponse>('/hr/attendance/clock-out', data)
      );
      return res as AttendanceClockOutResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['attendance'] });
      toast.success(t('attendance.clockedOut'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
