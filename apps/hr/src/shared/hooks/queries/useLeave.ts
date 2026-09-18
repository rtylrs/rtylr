import { createOneKitClient, getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { LeaveServiceClient } from '@voxire/voxie/hr/leave/v1/client';
import type {
  ListLeaveRequestsRequest,
  CreateLeaveRequestRequest,
  ApproveLeaveRequest,
  LeaveRequestPayload,
  LeaveTypePayload,
  LeaveBalancePayload,
  CreateLeaveTypeRequest,
} from '@voxire/voxie/hr/leave/v1/types';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

const leaveClient = createOneKitClient(LeaveServiceClient, 'hr');

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
      const res = await leaveClient.listLeaveTypes({});
      return res.leaveTypes ?? [];
    },
  });
};

export const useCreateLeaveType = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: CreateLeaveTypeRequest) => leaveClient.createLeaveType(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['leave-types'] });
      queryClient.invalidateQueries({ queryKey: ['leave-balances'] });
      toast.success(t('leave.policyCreated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const useListLeaveRequests = (params?: ListLeaveRequestsRequest) => {
  return useQuery<LeaveRequestPayload[]>({
    queryKey: ['leave-requests', params],
    queryFn: async (): Promise<LeaveRequestPayload[]> => {
      const res = await leaveClient.listLeaveRequests(
        params ?? { employeeId: undefined, status: undefined }
      );
      return res.requests ?? [];
    },
  });
};

export const useCreateLeaveRequest = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: CreateLeaveRequestRequest) => leaveClient.createLeaveRequest(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['leave-requests'] });
      queryClient.invalidateQueries({ queryKey: ['leave-balances'] });
      toast.success(t('leave.requestCreated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const useApproveLeave = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: ApproveLeaveRequest) => leaveClient.approveLeave(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['leave-requests'] });
      queryClient.invalidateQueries({ queryKey: ['leave-balances'] });
      queryClient.invalidateQueries({ queryKey: ['leave-calendar'] });
      toast.success(t('leave.requestUpdated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const useListLeaveBalances = (employeeId?: string) => {
  return useQuery<LeaveBalancePayload[]>({
    queryKey: ['leave-balances', employeeId],
    enabled: !!employeeId,
    queryFn: async (): Promise<LeaveBalancePayload[]> => {
      const res = await leaveClient.listLeaveBalances({ employeeId, year: undefined });
      return res.balances ?? [];
    },
  });
};

// No onekit contract exists for the leave calendar yet - stays on the legacy
// REST endpoint until hr_leave grows a matching rpc.
export const useGetLeaveCalendar = (params: LeaveCalendarParams) => {
  return useQuery<LeaveRequestPayload[]>({
    queryKey: ['leave-calendar', params],
    queryFn: async (): Promise<LeaveRequestPayload[]> => {
      const res = await unwrap(
        hrTransport.get<LeaveCalendarResponse>('/hr/leave/calendar', { params })
      );
      return (res as LeaveCalendarResponse).requests;
    },
  });
};
