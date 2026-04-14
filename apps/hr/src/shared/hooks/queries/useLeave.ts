import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import type { LeaveType, LeaveRequest, LeaveBalance, LeaveCalendarEntry } from '../../schemas';

interface ListLeaveTypesResponse {
  leaveTypes: LeaveType[];
}

interface ListLeaveRequestsParams {
  status?: string;
  staffId?: number;
}

interface ListLeaveRequestsResponse {
  requests: LeaveRequest[];
}

interface CreateLeaveRequestPayload {
  leaveTypeId: number;
  startDate: string;
  endDate: string;
  reason: string;
  notes?: string;
}

interface ApproveLeavePayload {
  id: number;
  status: string;
  notes?: string;
}

interface ListLeaveBalancesResponse {
  balances: LeaveBalance[];
}

interface LeaveCalendarParams {
  month: number;
  year: number;
  departmentId?: number;
  locationId?: number;
}

interface LeaveCalendarResponse {
  entries: LeaveCalendarEntry[];
}

export const useListLeaveTypes = () => {
  return useQuery<LeaveType[]>({
    queryKey: ['leave-types'],
    queryFn: async (): Promise<LeaveType[]> => {
      const res = await unwrap(safe.get<ListLeaveTypesResponse>('/hr/leave/types'));
      return (res as ListLeaveTypesResponse).leaveTypes;
    },
  });
};

export const useListLeaveRequests = (params?: ListLeaveRequestsParams) => {
  return useQuery<LeaveRequest[]>({
    queryKey: ['leave-requests', params],
    queryFn: async (): Promise<LeaveRequest[]> => {
      const res = await unwrap(
        safe.get<ListLeaveRequestsResponse>('/hr/leave/requests', { params })
      );
      return (res as ListLeaveRequestsResponse).requests;
    },
  });
};

export const useCreateLeaveRequest = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateLeaveRequestPayload): Promise<void> => {
      await unwrap(safe.post('/hr/leave/requests', data));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['leave-requests'] });
      queryClient.invalidateQueries({ queryKey: ['leave-balances'] });
      toast.success(t('leave.requestCreated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useApproveLeave = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: ApproveLeavePayload): Promise<void> => {
      await unwrap(safe.put(`/hr/leave/requests/${data.id}/approve`, data));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['leave-requests'] });
      queryClient.invalidateQueries({ queryKey: ['leave-balances'] });
      queryClient.invalidateQueries({ queryKey: ['leave-calendar'] });
      toast.success(t('leave.requestUpdated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useListLeaveBalances = () => {
  return useQuery<LeaveBalance[]>({
    queryKey: ['leave-balances'],
    queryFn: async (): Promise<LeaveBalance[]> => {
      const res = await unwrap(safe.get<ListLeaveBalancesResponse>('/hr/leave/balances'));
      return (res as ListLeaveBalancesResponse).balances;
    },
  });
};

export const useGetLeaveCalendar = (params: LeaveCalendarParams) => {
  return useQuery<LeaveCalendarEntry[]>({
    queryKey: ['leave-calendar', params],
    queryFn: async (): Promise<LeaveCalendarEntry[]> => {
      const res = await unwrap(safe.get<LeaveCalendarResponse>('/hr/leave/calendar', { params }));
      return (res as LeaveCalendarResponse).entries;
    },
  });
};
