import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  LeaveTypePayload,
  LeaveRequestPayload,
  LeaveBalancePayload,
} from '@voxire/contracts/hr/leave/v1/types';
import type { ListLeaveTypesResponse } from '@voxire/contracts/hr/leave/v1/list_leave_types';
import type {
  ListLeaveRequestsRequest,
  ListLeaveRequestsResponse,
} from '@voxire/contracts/hr/leave/v1/list_leave_requests';
import type { CreateLeaveRequestRequest } from '@voxire/contracts/hr/leave/v1/create_leave_request';
import type { ApproveLeaveRequest } from '@voxire/contracts/hr/leave/v1/approve_leave';
import type { ListLeaveBalancesResponse } from '@voxire/contracts/hr/leave/v1/list_leave_balances';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface LeaveCalendarParams {
  month: number;
  year: number;
  departmentId?: number;
  locationId?: number;
}

interface LeaveCalendarResponse {
  requests: LeaveRequestPayload[];
}

export const useListLeaveTypes = () => {
  return useQuery<LeaveTypePayload[]>({
    queryKey: ['leave-types'],
    queryFn: async (): Promise<LeaveTypePayload[]> => {
      const res = await unwrap(safe.get<ListLeaveTypesResponse>('/hr/leave/types'));
      return (res as ListLeaveTypesResponse).leaveTypes;
    },
  });
};

export const useListLeaveRequests = (params?: ListLeaveRequestsRequest) => {
  return useQuery<LeaveRequestPayload[]>({
    queryKey: ['leave-requests', params],
    queryFn: async (): Promise<LeaveRequestPayload[]> => {
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
    mutationFn: async (data: CreateLeaveRequestRequest): Promise<void> => {
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
    mutationFn: async (data: ApproveLeaveRequest): Promise<void> => {
      await unwrap(safe.post(`/hr/leave/requests/${data.leaveRequestId}/approve`, data));
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
  return useQuery<LeaveBalancePayload[]>({
    queryKey: ['leave-balances'],
    queryFn: async (): Promise<LeaveBalancePayload[]> => {
      const res = await unwrap(safe.get<ListLeaveBalancesResponse>('/hr/leave/balances'));
      return (res as ListLeaveBalancesResponse).balances;
    },
  });
};

export const useGetLeaveCalendar = (params: LeaveCalendarParams) => {
  return useQuery<LeaveRequestPayload[]>({
    queryKey: ['leave-calendar', params],
    queryFn: async (): Promise<LeaveRequestPayload[]> => {
      const res = await unwrap(safe.get<LeaveCalendarResponse>('/hr/leave/calendar', { params }));
      return (res as LeaveCalendarResponse).requests;
    },
  });
};
