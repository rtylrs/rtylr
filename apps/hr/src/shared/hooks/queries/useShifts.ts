import { createOneKitClient, getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { ShiftServiceClient } from '@voxire/voxie/hr/shift/v1/client';
import type {
  CreateShiftRequest,
  GetRosterResponse,
  ListShiftsRequest,
  ShiftAssignmentPayload,
  ShiftPayload,
} from '@voxire/voxie/hr/shift/v1/types';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { safe } from '../../lib/api';

const shiftClient = createOneKitClient(ShiftServiceClient, 'hr');

interface RosterParams {
  weekStart: string;
  locationId?: number;
}

interface RosterConflict {
  employeeId: number;
  date: string;
  conflictType: string;
  description: string;
}

interface RosterData {
  assignments: ShiftAssignmentPayload[];
  conflicts: RosterConflict[];
}

const decodeRosterConflicts = (values: string[] | undefined): RosterConflict[] =>
  (values ?? []).flatMap((value) => {
    try {
      const parsed = JSON.parse(value) as Partial<RosterConflict>;
      const employeeId = Number(parsed.employeeId);
      if (!Number.isInteger(employeeId) || !parsed.date) return [];
      return [
        {
          employeeId,
          date: parsed.date,
          conflictType: parsed.conflictType ?? 'unknown',
          description: parsed.description ?? '',
        },
      ];
    } catch {
      return [];
    }
  });

interface CreateShiftAssignmentPayload {
  staffId: number;
  shiftId: number;
  date: string;
}

export const useListShifts = () => {
  return useQuery<ShiftPayload[]>({
    queryKey: ['shifts'],
    queryFn: async (): Promise<ShiftPayload[]> => {
      const res = await shiftClient.listShifts({} satisfies ListShiftsRequest);
      return res.shifts ?? [];
    },
  });
};

export const useCreateShift = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateShiftRequest): Promise<ShiftPayload | undefined> => {
      const res = await shiftClient.createShift(data);
      return res.shift;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['shifts'] });
      toast.success(t('shifts.shiftCreated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const useDeleteShiftAssignment = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (assignmentId: number): Promise<void> => {
      await unwrap(safe.delete<{ success: boolean }>(`/hr/shift/v1/assignments/${assignmentId}`));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roster'] });
      toast.success(t('shifts.assignmentDeleted', { defaultValue: 'Shift removed' }));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const useGetRoster = (params: RosterParams) => {
  return useQuery<RosterData>({
    queryKey: ['roster', params],
    enabled: Boolean(params.locationId && params.weekStart),
    queryFn: async (): Promise<RosterData> => {
      if (!params.locationId) return { assignments: [], conflicts: [] };
      const res: GetRosterResponse = await shiftClient.getRoster({
        locationId: String(params.locationId),
        weekStart: params.weekStart,
      });
      return {
        assignments: res.assignments ?? [],
        conflicts: decodeRosterConflicts(res.conflicts),
      };
    },
  });
};

export const useCreateShiftAssignment = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateShiftAssignmentPayload): Promise<void> => {
      await shiftClient.createShiftAssignment({
        employeeId: String(data.staffId),
        shiftId: String(data.shiftId),
        date: data.date,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roster'] });
      toast.success(t('shifts.assignmentCreated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const usePublishRoster = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (
      data: string | { weekStart: string; locationId?: number }
    ): Promise<void> => {
      const request = typeof data === 'string' ? { weekStart: data } : data;
      await shiftClient.publishRoster({
        weekStart: request.weekStart,
        locationId: request.locationId !== undefined ? String(request.locationId) : undefined,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['roster'] });
      toast.success(t('shifts.rosterPublished'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export type { RosterConflict, RosterData };
