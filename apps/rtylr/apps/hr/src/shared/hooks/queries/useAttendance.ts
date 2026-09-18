import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  ListAttendanceRequest,
  ListAttendanceResponse,
} from '@voxire/voxie/hr/attendance/v1/types';
import type { AttendanceClockInResponse } from '@voxire/voxie/hr/attendance/v1/types';
import type { AttendanceClockOutResponse } from '@voxire/voxie/hr/attendance/v1/types';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

export interface ClockInPayload {
  ownerId: number;
  employeeId: number;
  method: string;
  locationId: number;
  lat?: number;
  lng?: number;
  qrToken?: string;
  pin?: string;
}

interface BackendClockInPayload {
  ownerId: number;
  employeeId: number;
  locationId: number;
  clockInMethod: string;
  clockInLat?: number;
  clockInLng?: number;
  qrToken?: string;
  pin?: string;
}

export interface ClockOutPayload {
  employeeId: number;
}

export interface SetPinPayload {
  employeeId: number;
  pin: string;
}

const toBackendClockInPayload = (data: ClockInPayload): BackendClockInPayload => ({
  ownerId: data.ownerId,
  employeeId: data.employeeId,
  locationId: data.locationId,
  clockInMethod: data.method,
  clockInLat: data.lat,
  clockInLng: data.lng,
  qrToken: data.qrToken,
  pin: data.pin,
});

export const useGetMyAttendance = (params: ListAttendanceRequest) => {
  return useQuery<ListAttendanceResponse>({
    queryKey: ['attendance', 'filter', params],
    queryFn: async (): Promise<ListAttendanceResponse> => {
      const res = await unwrap(
        hrTransport.get<ListAttendanceResponse>('/hr/attendance', { params })
      );
      return res as ListAttendanceResponse;
    },
  });
};

export const useClockIn = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: ClockInPayload): Promise<AttendanceClockInResponse> => {
      const res = await unwrap(
        hrTransport.post<AttendanceClockInResponse>(
          '/hr/attendance/clock-in',
          toBackendClockInPayload(data)
        )
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
    mutationFn: async (data: SetPinPayload): Promise<{ success: boolean }> => {
      const res = await unwrap(hrTransport.put<{ success: boolean }>('/hr/employees/me/pin', data));
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
    mutationFn: async (data: ClockOutPayload): Promise<AttendanceClockOutResponse> => {
      const res = await unwrap(
        hrTransport.post<AttendanceClockOutResponse>('/hr/attendance/clock-out', data)
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
