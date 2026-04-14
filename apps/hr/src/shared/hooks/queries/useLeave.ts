import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  LeaveTypePayload,
  LeaveRequestPayload,
  LeaveBalancePayload,
  LeaveCalendarEntryPayload,
  ListLeaveTypesResponse,
  ListLeaveRequestsRequest,
  ListLeaveRequestsResponse,
  CreateLeaveRequestRequest,
  ApproveLeaveRequestRequest,
  ListLeaveBalancesResponse,
  GetLeaveCalendarRequest,
  GetLeaveCalendarResponse,
} from '@voxire/contracts/hr/v1/leave';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

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
    mutationFn: async (data: ApproveLeaveRequestRequest): Promise<void> => {
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
  return useQuery<LeaveBalancePayload[]>({
    queryKey: ['leave-balances'],
    queryFn: async (): Promise<LeaveBalancePayload[]> => {
      const res = await unwrap(safe.get<ListLeaveBalancesResponse>('/hr/leave/balances'));
      return (res as ListLeaveBalancesResponse).balances;
    },
  });
};

export const useGetLeaveCalendar = (params: GetLeaveCalendarRequest) => {
  return useQuery<LeaveCalendarEntryPayload[]>({
    queryKey: ['leave-calendar', params],
    queryFn: async (): Promise<LeaveCalendarEntryPayload[]> => {
      const res = await unwrap(
        safe.get<GetLeaveCalendarResponse>('/hr/leave/calendar', { params })
      );
      return (res as GetLeaveCalendarResponse).entries;
    },
  });
};
