import { createOneKitClient, getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { OvertimeServiceClient } from '@voxire/voxie/hr/overtime/v1/client';
import type {
  ListOvertimeRequestsRequest,
  CreateOvertimeRequestRequest,
  ApproveOvertimeRequest,
  OvertimeRequestPayload,
} from '@voxire/voxie/hr/overtime/v1/types';
import { hrTransport } from '../../lib/clients';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

const overtimeClient = createOneKitClient(OvertimeServiceClient, 'hr');

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
      const res = await overtimeClient.listOvertimeRequests(
        params ?? { employeeId: undefined, status: undefined }
      );
      return res.requests ?? [];
    },
  });
};

export const useCreateOvertimeRequest = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: CreateOvertimeRequestRequest) => overtimeClient.createOvertimeRequest(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overtime-requests'] });
      toast.success(t('overtime.requestCreated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const useApproveOvertime = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: ApproveOvertimeRequest) => overtimeClient.approveOvertime(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['overtime-requests'] });
      toast.success(t('overtime.requestUpdated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

// No onekit contract exists for overtime config yet - stays on the legacy
// REST endpoint until hr_overtime grows a matching rpc.
export const useGetOvertimeConfig = () => {
  return useQuery<OvertimeConfigData>({
    queryKey: ['overtime-config'],
    queryFn: async (): Promise<OvertimeConfigData> => {
      const res = await unwrap(
        hrTransport.get<{ config: OvertimeConfigData }>('/hr/overtime/config')
      );
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
        hrTransport.put('/hr/overtime/config', {
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
