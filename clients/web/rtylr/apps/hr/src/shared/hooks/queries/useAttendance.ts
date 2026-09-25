import { createOneKitClient, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  ListAttendanceRequest,
  ListAttendanceResponse,
  AttendanceClockInResponse,
  AttendanceClockOutResponse,
} from '@voxire/voxie/hr/attendance/v1/types';
import { AttendanceServiceClient } from '@voxire/voxie/hr/attendance/v1/client';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

const attendanceClient = createOneKitClient(AttendanceServiceClient, 'hr');

export interface ClockInPayload {
  method: string;
  locationId: string;
  lat?: number;
  lng?: number;
  qrToken?: string;
  pin?: string;
}

export interface ClockOutPayload {
  attendanceRecordId: string;
}

export interface SetPinPayload {
  employeeId: number;
  pin: string;
}

export interface AttendanceLocation {
  id: string;
  name: string;
  isDefault: boolean;
}

interface LocationResponse {
  locations?: Array<{
    id?: string | number;
    name?: string;
    isDefault?: boolean;
    is_default?: boolean;
  }>;
}

export const useGetMyAttendance = (params: ListAttendanceRequest) => {
  return useQuery<ListAttendanceResponse>({
    queryKey: ['attendance', 'filter', params],
    queryFn: () => attendanceClient.listAttendance(params),
  });
};

export const useAttendanceLocations = () => {
  return useQuery<AttendanceLocation[]>({
    queryKey: ['attendance', 'locations'],
    queryFn: async (): Promise<AttendanceLocation[]> => {
      const response = await unwrap(hrTransport.get<LocationResponse>('/hr/locations'));
      return (response.locations ?? []).flatMap((location) => {
        const id = location.id === undefined ? '' : String(location.id);
        if (!/^[1-9][0-9]*$/.test(id)) return [];
        return [
          {
            id,
            name: location.name ?? '',
            isDefault: location.isDefault ?? location.is_default ?? false,
          },
        ];
      });
    },
    staleTime: 5 * 60_000,
  });
};

export const useClockIn = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: ClockInPayload): Promise<AttendanceClockInResponse> =>
      attendanceClient.attendanceClockIn({
        method: data.method,
        locationId: data.locationId,
        lat: data.lat,
        lng: data.lng,
        qrToken: data.qrToken,
        pin: data.pin,
      }),
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
    mutationFn: (data: ClockOutPayload): Promise<AttendanceClockOutResponse> =>
      attendanceClient.attendanceClockOut({
        attendanceRecordId: data.attendanceRecordId,
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['attendance'] });
      toast.success(t('attendance.clockedOut'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
