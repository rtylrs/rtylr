import { createOneKitClient, getOneKitErrorMessage, unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { OvertimeServiceClient } from '@voxire/voxie/hr/overtime/v1/client';
import type {
  ListOvertimeRequestsRequest,
  CreateOvertimeRequestRequest,
  ApproveOvertimeRequest,
  OvertimeRequestPayload,
} from '@voxire/voxie/hr/overtime/v1/types';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

const overtimeClient = createOneKitClient(OvertimeServiceClient, 'hr');

interface OvertimeConfigData {
  dailyOtThresholdMinutes: number;
  weeklyOtThresholdMinutes: number;
  defaultMultiplier: number;
  autoDetectEnabled: boolean;
}

interface UpdateOvertimeConfigData {
  dailyOtThresholdMinutes?: number;
  weeklyOtThresholdMinutes?: number;
  defaultMultiplier?: number;
  autoDetectEnabled?: boolean;
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
        safe.get<OvertimeConfigData & { config?: OvertimeConfigData }>('/hr/overtime/config')
      );
      const raw = (res as { config?: OvertimeConfigData }).config ?? res;
      return {
        dailyOtThresholdMinutes: raw.dailyOtThresholdMinutes ?? 480,
        weeklyOtThresholdMinutes: raw.weeklyOtThresholdMinutes ?? 2400,
        defaultMultiplier: raw.defaultMultiplier ?? 1.5,
        autoDetectEnabled: raw.autoDetectEnabled ?? false,
      };
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
          weeklyThresholdMinutes: data.weeklyOtThresholdMinutes ?? 2400,
          defaultMultiplier: data.defaultMultiplier,
          autoDetectEnabled: data.autoDetectEnabled ?? true,
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
