import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { EquipmentServiceClient } from '@voxire/voxie/hr/equipment/v1/client';
import type {
  ListEquipmentRequestsRequest,
  ListEquipmentRequestsResponse,
  ApproveEquipmentRequest,
  ListAssetsResponse,
} from '@voxire/voxie/hr/equipment/v1/types';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

const equipmentClient = createOneKitClient(EquipmentServiceClient, 'hr');

export interface CreateEquipmentRequestInput {
  employeeId: string;
  equipmentType: string;
  description: string;
  neededByDate?: string;
}

export const useListEquipmentRequests = (params?: ListEquipmentRequestsRequest) => {
  return useQuery<ListEquipmentRequestsResponse>({
    queryKey: ['hr', 'equipment-requests', params],
    queryFn: () =>
      equipmentClient.listEquipmentRequests(params ?? { employeeId: undefined, status: undefined }),
  });
};

export const useCreateEquipmentRequest = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: (data: CreateEquipmentRequestInput) =>
      equipmentClient.createEquipmentRequest({
        ...data,
        // The generated request requires urgency, but the live HR equipment
        // adapter currently discards it. Keep the transport contract valid
        // without presenting a choice the server cannot persist.
        urgency: 'medium',
      }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'equipment-requests'] });
      toast.success(t('hrEquipment.requestCreated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('hrErrors.generic')));
    },
  });
};

export const useApproveEquipment = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('hrCommon');

  return useMutation({
    mutationFn: (data: ApproveEquipmentRequest) => equipmentClient.approveEquipment(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['hr', 'equipment-requests'] });
      toast.success(t('hrEquipment.requestUpdated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('hrErrors.generic')));
    },
  });
};

export const useListAssets = (employeeId?: string) => {
  return useQuery<ListAssetsResponse>({
    queryKey: ['hr', 'assets', employeeId],
    queryFn: () =>
      equipmentClient.listAssets({ employeeId: employeeId ?? undefined, condition: undefined }),
  });
};
