import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  ListTimesheetsRequest,
  ListTimesheetsResponse,
} from '@voxire/voxie/hr/time_entry/v1/types';
import type { TimesheetPayload } from '@voxire/voxie/hr/time_entry/v1/types';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface SubmitTimesheetRequest {
  timesheetId: number;
}

interface SubmitTimesheetResponse {
  success: boolean;
}

interface ApproveTimesheetRequest {
  timesheetId: number;
  approved: boolean;
}

interface ApproveTimesheetResponse {
  success: boolean;
}

interface RejectTimesheetRequest {
  timesheetId: number;
  rejectionReason?: string;
}

export const useListTimesheets = (params: ListTimesheetsRequest) => {
  return useQuery<ListTimesheetsResponse>({
    queryKey: ['timesheets', 'filter', params],
    queryFn: async (): Promise<ListTimesheetsResponse> => {
      const res = await unwrap(
        hrTransport.get<ListTimesheetsResponse>('/hr/timesheets', { params })
      );
      return res as ListTimesheetsResponse;
    },
  });
};

export const useGetTimesheet = (id: number | undefined) => {
  return useQuery<TimesheetPayload>({
    queryKey: ['timesheet', id],
    queryFn: async (): Promise<TimesheetPayload> => {
      const res = await unwrap(
        hrTransport.get<{ timesheet: TimesheetPayload }>(`/hr/timesheets/${id}`)
      );
      return (res as { timesheet: TimesheetPayload }).timesheet;
    },
    enabled: !!id,
  });
};

export const useSubmitTimesheet = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: SubmitTimesheetRequest): Promise<SubmitTimesheetResponse> => {
      const res = await unwrap(
        hrTransport.post<SubmitTimesheetResponse>(`/hr/timesheets/${data.timesheetId}/submit`, {})
      );
      return res as SubmitTimesheetResponse;
    },
    onSuccess: (_data: SubmitTimesheetResponse, variables: SubmitTimesheetRequest) => {
      queryClient.invalidateQueries({ queryKey: ['timesheets'] });
      queryClient.invalidateQueries({ queryKey: ['timesheet', variables.timesheetId] });
      toast.success(t('timesheets.submitted'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useApproveTimesheet = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: ApproveTimesheetRequest): Promise<ApproveTimesheetResponse> => {
      const res = await unwrap(
        hrTransport.post<ApproveTimesheetResponse>(`/hr/timesheets/${data.timesheetId}/approve`, {})
      );
      return res as ApproveTimesheetResponse;
    },
    onSuccess: (_data: ApproveTimesheetResponse, variables: ApproveTimesheetRequest) => {
      queryClient.invalidateQueries({ queryKey: ['timesheets'] });
      queryClient.invalidateQueries({ queryKey: ['timesheet', variables.timesheetId] });
      toast.success(t('timesheets.approved'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useRejectTimesheet = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: RejectTimesheetRequest): Promise<ApproveTimesheetResponse> => {
      const res = await unwrap(
        hrTransport.post<ApproveTimesheetResponse>(`/hr/timesheets/${data.timesheetId}/reject`, {
          rejectionReason: data.rejectionReason,
        })
      );
      return res as ApproveTimesheetResponse;
    },
    onSuccess: (_data: ApproveTimesheetResponse, variables: RejectTimesheetRequest) => {
      queryClient.invalidateQueries({ queryKey: ['timesheets'] });
      queryClient.invalidateQueries({ queryKey: ['timesheet', variables.timesheetId] });
      toast.success(t('timesheets.rejected'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
