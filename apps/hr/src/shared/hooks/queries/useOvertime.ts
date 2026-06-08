import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { OvertimeRequestPayload } from '@voxire/contracts/hr/overtime/v1/types';
import type {
  ListOvertimeRequestsRequest,
  ListOvertimeRequestsResponse,
} from '@voxire/contracts/hr/overtime/v1/list_overtime_requests';
import type { CreateOvertimeRequestRequest } from '@voxire/contracts/hr/overtime/v1/create_overtime_request';
import type { ApproveOvertimeRequest } from '@voxire/contracts/hr/overtime/v1/approve_overtime';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface OvertimeConfigData {
  dailyOtThresholdMinutes: number;
  weeklyOtThresholdMinutes: number;
  defaultMultiplier: number;
}

interface UpdateOvertimeConfigData {
  dailyOtThresholdMinutes?: number;
  weeklyOtThresholdMinutes?: number;
  defaultMultiplier?: number;
}

export const useListOvertimeRequests = (params?: ListOvertimeRequestsRequest) => {
  return useQuery<OvertimeRequestPayload[]>({
    queryKey: ['overtime-requests', params],
    queryFn: async (): Promise<OvertimeRequestPayload[]> => {
      const res = await unwrap(
        safe.get<ListOvertimeRequestsResponse>('/hr/overtime/requests', { params })
      );
      return (res as ListOvertimeRequestsResponse).requests;
    },
  });
};

export const useCreateOvertimeRequest = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: CreateOvertimeRequestRequest): Promise<void> => {
      await unwrap(safe.post('/hr/overtime/requests', data));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overtime-requests'] });
      toast.success(t('overtime.requestCreated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useApproveOvertime = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: ApproveOvertimeRequest): Promise<void> => {
      await unwrap(
        safe.post(`/hr/overtime/requests/${data.overtimeRequestId}/approve`, {
          requestId: data.overtimeRequestId,
          approved: data.approved,
          multiplierApplied: data.multiplierApplied,
        })
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overtime-requests'] });
      toast.success(t('overtime.requestUpdated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useGetOvertimeConfig = () => {
  return useQuery<OvertimeConfigData>({
    queryKey: ['overtime-config'],
    queryFn: async (): Promise<OvertimeConfigData> => {
      const res = await unwrap(safe.get<{ config: OvertimeConfigData }>('/hr/overtime/config'));
      return (res as { config: OvertimeConfigData }).config;
    },
  });
};

export const useUpdateOvertimeConfig = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: UpdateOvertimeConfigData): Promise<void> => {
      await unwrap(
        safe.put('/hr/overtime/config', {
          dailyThresholdMinutes: data.dailyOtThresholdMinutes,
          defaultMultiplier: data.defaultMultiplier,
          autoDetectEnabled: true,
        })
      );
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overtime-config'] });
      toast.success(t('overtime.configUpdated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};
