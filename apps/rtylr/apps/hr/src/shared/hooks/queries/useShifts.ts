import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ShiftPayload } from '@voxire/voxie/hr/shift/v1/types';
import type { CreateShiftRequest } from '@voxire/voxie/hr/shift/v1/types';
import type { ListShiftsResponse } from '@voxire/voxie/hr/shift/v1/types';
import type { CreateShiftAssignmentResponse } from '@voxire/voxie/hr/shift/v1/types';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

/**
 * Availability warnings are advisory and easy to miss, so they outlast the
 * default toast. A scheduler who assigned someone outside their stated hours
 * needs long enough to read who and when before deciding to undo it.
 */
const WARNING_TOAST_DURATION_MS = 10_000;

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
      const res = await unwrap(hrTransport.get<ListShiftsResponse>('/hr/shifts'));
      return (res as ListShiftsResponse).shifts ?? [];
    },
  });
};

export const useCreateShift = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateShiftRequest): Promise<void> => {
      await unwrap(
        hrTransport.post('/hr/shifts', {
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
      const res = await unwrap(hrTransport.get<RosterResponse>('/hr/shifts/roster', { params }));
      return (res as RosterResponse).entries;
    },
  });
};

export const useCreateShiftAssignment = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: CreateShiftAssignmentPayload): Promise<CreateShiftAssignmentResponse> =>
      unwrap(
        hrTransport.post<CreateShiftAssignmentResponse>('/hr/shifts/roster/assignments', {
          employeeId: data.staffId,
          shiftId: data.shiftId,
          date: data.date,
        })
      ),
    onSuccess: (response) => {
      queryClient.invalidateQueries({ queryKey: ['roster'] });
      toast.success(t('shifts.assignmentCreated'));

      // Advisory conflicts from the contract's non-blocking warnings field. The
      // assignment was created, so a scheduling clash with the assignee's stated
      // availability arrives here rather than as an error, and dropping it would
      // leave the roster silently holding a shift nobody agreed to work.
      //
      // Shown as given because the server composes these strings; they are not
      // translation keys, and there is no code alongside them to map to one.
      for (const warning of response.warnings ?? []) {
        toast.warning(warning, { duration: WARNING_TOAST_DURATION_MS });
      }
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
      await unwrap(hrTransport.post('/hr/shifts/roster/publish', { weekStart }));
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
