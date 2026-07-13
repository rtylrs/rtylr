import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { EquipmentServiceClient } from '@voxire/voxie/hr/equipment/v1/client';
import type {
  ListEquipmentRequestsRequest,
  ListEquipmentRequestsResponse,
  CreateEquipmentRequestRequest,
  ApproveEquipmentRequest,
  ListAssetsResponse,
} from '@voxire/voxie/hr/equipment/v1/types';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

const equipmentClient = createOneKitClient(EquipmentServiceClient, 'hr');

export const useListEquipmentRequests = (params?: ListEquipmentRequestsRequest) => {
  return useQuery<ListEquipmentRequestsResponse>({
    queryKey: ['equipment-requests', params],
    queryFn: () =>
      equipmentClient.listEquipmentRequests(params ?? { employeeId: undefined, status: undefined }),
  });
};

export const useCreateEquipmentRequest = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: CreateEquipmentRequestRequest) =>
      equipmentClient.createEquipmentRequest(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['equipment-requests'] });
      toast.success(t('equipment.requestCreated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const useApproveEquipment = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: (data: ApproveEquipmentRequest) => equipmentClient.approveEquipment(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['equipment-requests'] });
      toast.success(t('equipment.requestUpdated'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('errors.generic')));
    },
  });
};

export const useListAssets = (employeeId?: string) => {
  return useQuery<ListAssetsResponse>({
    queryKey: ['assets', employeeId],
    queryFn: () =>
      equipmentClient.listAssets({ employeeId: employeeId ?? undefined, condition: undefined }),
  });
};
