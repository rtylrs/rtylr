import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { OvertimeServiceClient } from '@voxire/voxie/hr/overtime/v1/client';
import type {
  ListOvertimeRequestsRequest,
  CreateOvertimeRequestRequest,
  ApproveOvertimeRequest,
  OvertimeRequestPayload,
} from '@voxire/voxie/hr/overtime/v1/types';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

const overtimeClient = createOneKitClient(OvertimeServiceClient, 'hr');

export const useListOvertimeRequests = (params?: ListOvertimeRequestsRequest) => {
  return useQuery<OvertimeRequestPayload[]>({
    queryKey: ['hr', 'overtime-requests', params],
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
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: (data: CreateOvertimeRequestRequest) => overtimeClient.createOvertimeRequest(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'overtime-requests'] });
      toast.success(t('hrOvertime.requestCreated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('hrErrors.generic')));
    },
  });
};

export const useApproveOvertime = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: (data: ApproveOvertimeRequest) => overtimeClient.approveOvertime(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'overtime-requests'] });
      toast.success(t('hrOvertime.requestUpdated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('hrErrors.generic')));
    },
  });
};
