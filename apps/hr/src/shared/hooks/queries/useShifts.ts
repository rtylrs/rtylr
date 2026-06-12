import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ShiftPayload } from '@voxire/contracts/hr/shift/v1/types';
import type { CreateShiftRequest } from '@voxire/contracts/hr/shift/v1/create_shift';
import type { ListShiftsResponse } from '@voxire/contracts/hr/shift/v1/list_shifts';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface RosterParams {
  weekStart: string;
  locationId?: number;
}

interface RosterEntry {
  staffId: number;
  staffName: string;
  assignments: ShiftAssignment[];
}

interface ShiftAssignment {
  id: number;
  shiftId: number;
  shiftName: string;
  date: string;
  startTime: string;
  endTime: string;
  color: string;
  conflict: boolean;
}

interface RosterResponse {
  entries: RosterEntry[];
}

interface CreateShiftAssignmentPayload {
  staffId: number;
  shiftId: number;
  date: string;
}

export const useListShifts = () => {
  return useQuery<ShiftPayload[]>({
    queryKey: ['shifts'],
    queryFn: async (): Promise<ShiftPayload[]> => {
      const res = await unwrap(safe.get<ListShiftsResponse>('/hr/shifts'));
      return (res as ListShiftsResponse).shifts;
    },
  });
};

export const useCreateShift = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateShiftRequest): Promise<void> => {
      await unwrap(
        safe.post('/hr/shifts', {
          name: data.name,
          locationId: data.locationId,
          startTime: data.startTime,
          endTime: data.endTime,
          color: data.color,
        })
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shifts'] });
      toast.success(t('shifts.shiftCreated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useGetRoster = (params: RosterParams) => {
  return useQuery<RosterEntry[]>({
    queryKey: ['roster', params],
    queryFn: async (): Promise<RosterEntry[]> => {
      const res = await unwrap(safe.get<RosterResponse>('/hr/shifts/roster', { params }));
      return (res as RosterResponse).entries;
    },
  });
};

export const useCreateShiftAssignment = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateShiftAssignmentPayload): Promise<void> => {
      await unwrap(
        safe.post('/hr/shifts/roster/assignments', {
          employeeId: data.staffId,
          shiftId: data.shiftId,
          date: data.date,
        })
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roster'] });
      toast.success(t('shifts.assignmentCreated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const usePublishRoster = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (weekStart: string): Promise<void> => {
      await unwrap(safe.post('/hr/shifts/roster/publish', { weekStart }));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roster'] });
      toast.success(t('shifts.rosterPublished'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export type { RosterEntry, ShiftAssignment };
