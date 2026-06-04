import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ListAttendanceRequest, ListAttendanceResponse } from '@voxire/contracts/hr/attendance/v1/list_attendance';
import type { AttendanceClockInRequest, AttendanceClockInResponse } from '@voxire/contracts/hr/attendance/v1/attendance_clock_in';
import type { AttendanceClockOutRequest, AttendanceClockOutResponse } from '@voxire/contracts/hr/attendance/v1/attendance_clock_out';
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

export const useSetPin = () => {
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (pin: string): Promise<{ success: boolean }> => {
      const res = await unwrap(safe.post<{ success: boolean }>('/hr/attendance/set-pin', { pin }));
      return res as { success: boolean };
    },
    onSuccess: () => {
      toast.success(t('attendance.pinModal.pinSaved'));
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
