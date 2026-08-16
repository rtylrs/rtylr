import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createOneKitClient } from '@voxire/api';
import { LeaveServiceClient } from '@voxire/voxie/hr/leave/v1/client';
import type {
  CreateLeaveRequestRequest,
  LeaveBalancePayload,
  LeaveRequestPayload,
  LeaveTypePayload,
} from '@voxire/voxie/hr/leave/v1/types';
import { toast } from 'sonner';
import i18next from '../../shared/services/i18next';

const leaveClient = createOneKitClient(LeaveServiceClient, 'flow');

const t = (key: string): string => i18next.t(key) as string;

export type { LeaveRequestPayload, LeaveBalancePayload, LeaveTypePayload };

export function useListLeaveTypes() {
  return useQuery({
    queryKey: ['leave-types'],
    queryFn: async (): Promise<LeaveTypePayload[]> => {
      const res = await leaveClient.listLeaveTypes({});
      return res.leaveTypes ?? [];
    },
  });
}

export function useMyLeaveRequests(employeeId: string | undefined) {
  return useQuery({
    queryKey: ['leave-requests', employeeId],
    enabled: !!employeeId,
    queryFn: async (): Promise<LeaveRequestPayload[]> => {
      const res = await leaveClient.listLeaveRequests({ employeeId, status: undefined });
      return res.requests ?? [];
    },
  });
}

export function useMyLeaveBalances(employeeId: string | undefined) {
  return useQuery({
    queryKey: ['leave-balances', employeeId],
    enabled: !!employeeId,
    queryFn: async (): Promise<LeaveBalancePayload[]> => {
      const res = await leaveClient.listLeaveBalances({ employeeId, year: undefined });
      return res.balances ?? [];
    },
  });
}

export function useCreateLeaveRequest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateLeaveRequestRequest) => leaveClient.createLeaveRequest(input),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['leave-requests'] });
      void queryClient.invalidateQueries({ queryKey: ['leave-balances'] });
      toast.success(t('flow.leave.requestSubmitted'));
    },
    onError: () => toast.error(t('flow.leave.requestFailed')),
  });
}

export function useCancelLeaveRequest() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (requestId: string) => leaveClient.cancelLeaveRequest({ requestId }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['leave-requests'] });
      void queryClient.invalidateQueries({ queryKey: ['leave-balances'] });
      toast.success(t('flow.leave.requestCanceled'));
    },
    onError: () => toast.error(t('flow.leave.cancelFailed')),
  });
}
