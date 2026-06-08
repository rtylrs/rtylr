import { unwrap } from '@voxire/api';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type {
  ListEquipmentRequestsResponse,
  ListEquipmentRequestsRequest,
} from '@voxire/contracts/hr/equipment/v1/list_equipment_requests';
import type {
  CreateEquipmentRequestRequest,
  CreateEquipmentRequestResponse,
} from '@voxire/contracts/hr/equipment/v1/create_equipment_request';
import type {
  ApproveEquipmentRequest,
  ApproveEquipmentResponse,
} from '@voxire/contracts/hr/equipment/v1/approve_equipment';
import type { ListAssetsResponse } from '@voxire/contracts/hr/equipment/v1/list_assets';
import { safe } from '../../lib/api';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';

interface ApproveEquipmentPayload {
  approve: boolean;
  reason: string;
}

const toApproveEquipmentPayload = (data: ApproveEquipmentRequest): ApproveEquipmentPayload => ({
  approve: data.approved,
  reason: '',
});

export const useListEquipmentRequests = (params?: ListEquipmentRequestsRequest) => {
  return useQuery<ListEquipmentRequestsResponse>({
    queryKey: ['equipment-requests', params],
    queryFn: async (): Promise<ListEquipmentRequestsResponse> => {
      const res = await unwrap(
        safe.get<ListEquipmentRequestsResponse>('/hr/equipment/requests', { params })
      );
      return res as ListEquipmentRequestsResponse;
    },
  });
};

export const useCreateEquipmentRequest = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (
      data: CreateEquipmentRequestRequest
    ): Promise<CreateEquipmentRequestResponse> => {
      const res = await unwrap(
        safe.post<CreateEquipmentRequestResponse>('/hr/equipment/requests', data)
      );
      return res as CreateEquipmentRequestResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['equipment-requests'] });
      toast.success(t('equipment.requestCreated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useApproveEquipment = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation();

  return useMutation({
    mutationFn: async (data: ApproveEquipmentRequest): Promise<ApproveEquipmentResponse> => {
      const res = await unwrap(
        safe.post<ApproveEquipmentResponse>(
          `/hr/equipment/requests/${data.equipmentRequestId}/approve`,
          toApproveEquipmentPayload(data)
        )
      );
      return res as ApproveEquipmentResponse;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['equipment-requests'] });
      toast.success(t('equipment.requestUpdated'));
    },
    onError: () => {
      toast.error(t('errors.generic'));
    },
  });
};

export const useListAssets = (employeeId?: number) => {
  return useQuery<ListAssetsResponse>({
    queryKey: ['assets', employeeId],
    queryFn: async (): Promise<ListAssetsResponse> => {
      const res = await unwrap(
        safe.get<ListAssetsResponse>('/hr/equipment/assets', {
          params: employeeId ? { employeeId } : undefined,
        })
      );
      return res as ListAssetsResponse;
    },
  });
};
