import { createOneKitClient, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { ShiftAssignmentPayload, ShiftPayload } from '@voxire/voxie/hr/shift/v1/types';
import type { CreateShiftRequest } from '@voxire/voxie/hr/shift/v1/types';
import type { ListShiftsResponse } from '@voxire/voxie/hr/shift/v1/types';
import type { CreateShiftAssignmentResponse } from '@voxire/voxie/hr/shift/v1/types';
import { StaffServiceClient } from '@voxire/voxie/hr/staff/v1/client';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

/**
 * Availability warnings are advisory and easy to miss, so they outlast the
 * default toast. A scheduler who assigned someone outside their stated hours
 * needs long enough to read who and when before deciding to undo it.
 */
const WARNING_TOAST_DURATION_MS = 10_000;
const staffClient = createOneKitClient(StaffServiceClient, 'hr');

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

interface RosterConflict {
  employeeId?: number | string;
  employee_id?: number | string;
  date?: string;
}

interface RosterResponse {
  assignments?: ShiftAssignmentPayload[];
  conflicts?: Array<RosterConflict | string>;
}

interface CreateShiftAssignmentPayload {
  staffId: number;
  shiftId: number;
  date: string;
}

export const useListShifts = () => {
  return useQuery<ShiftPayload[]>({
    queryKey: ['hr', 'shifts'],
    queryFn: async (): Promise<ShiftPayload[]> => {
      const res = await unwrap(hrTransport.get<ListShiftsResponse>('/hr/shifts'));
      return (res as ListShiftsResponse).shifts ?? [];
    },
  });
};

export const useCreateShift = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

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
      queryClient.invalidateQueries({ queryKey: ['hr', 'shifts'] });
      toast.success(t('hrShifts.shiftCreated'));
    },
    onError: () => {
      toast.error(t('hrErrors.generic'));
    },
  });
};

export const useGetRoster = (params: RosterParams) => {
  const queryClient = useQueryClient();

  return useQuery<RosterEntry[]>({
    queryKey: ['hr', 'roster', params],
    queryFn: async (): Promise<RosterEntry[]> => {
      const [res, shifts, staff] = await Promise.all([
        unwrap(hrTransport.get<RosterResponse>('/hr/shifts/roster', { params })),
        queryClient.fetchQuery({
          queryKey: ['hr', 'shifts'],
          queryFn: async (): Promise<ShiftPayload[]> => {
            const response = await unwrap(hrTransport.get<ListShiftsResponse>('/hr/shifts'));
            return (response as ListShiftsResponse).shifts ?? [];
          },
        }),
        staffClient.listStaff({ search: '', departmentId: '', status: '' }),
      ]);

      const assignments = res.assignments ?? [];
      const shiftsById = new Map(shifts.map((shift) => [Number(shift.id), shift]));
      const staffById = new Map((staff.staff ?? []).map((member) => [Number(member.id), member]));
      const conflictKeys = new Set<string>();
      let hasUnscopedConflict = false;

      for (const conflict of res.conflicts ?? []) {
        if (typeof conflict === 'string') {
          hasUnscopedConflict = true;
          continue;
        }
        const employeeId = Number(conflict.employeeId ?? conflict.employee_id);
        if (!Number.isFinite(employeeId) || !conflict.date) {
          hasUnscopedConflict = true;
          continue;
        }
        conflictKeys.add(`${employeeId}:${conflict.date}`);
      }

      const employeeIds = new Set([
        ...staffById.keys(),
        ...assignments.map((assignment) => Number(assignment.employeeId)),
      ]);

      return Array.from(employeeIds, (staffId) => {
        const member = staffById.get(staffId);
        return {
          staffId,
          staffName: member
            ? `${member.firstName} ${member.lastName}`.trim()
            : `Employee #${staffId}`,
          assignments: assignments
            .filter((assignment) => Number(assignment.employeeId) === staffId)
            .map((assignment) => {
              const shiftId = Number(assignment.shiftId);
              const shift = shiftsById.get(shiftId);
              return {
                id: Number(assignment.id),
                shiftId,
                shiftName: shift?.name ?? '',
                date: assignment.date ?? '',
                startTime: shift?.startTime ?? '',
                endTime: shift?.endTime ?? '',
                color: shift?.color ?? '',
                conflict: hasUnscopedConflict || conflictKeys.has(`${staffId}:${assignment.date}`),
              };
            }),
        };
      });
    },
  });
};

export const useCreateShiftAssignment = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

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
      queryClient.invalidateQueries({ queryKey: ['hr', 'roster'] });
      toast.success(t('hrShifts.assignmentCreated'));

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
      toast.error(t('hrErrors.generic'));
    },
  });
};

export const usePublishRoster = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: async (weekStart: string): Promise<void> => {
      await unwrap(hrTransport.post('/hr/shifts/roster/publish', { weekStart }));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'roster'] });
      toast.success(t('hrShifts.rosterPublished'));
    },
    onError: () => {
      toast.error(t('hrErrors.generic'));
    },
  });
};

export type { RosterEntry, ShiftAssignment };
