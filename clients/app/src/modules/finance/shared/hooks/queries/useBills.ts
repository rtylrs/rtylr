import { createOneKitClient, getOneKitErrorMessage } from '@voxire/api';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { useTranslation } from 'react-i18next';
import { BillServiceClient } from '@voxire/voxie/finance/bill/v1/client';
import type {
  CreateBillRequest,
  CreateBillResponse,
  UpdateBillRequest,
  UpdateBillResponse,
} from '@voxire/voxie/finance/bill/v1/types';

const billClient = createOneKitClient(BillServiceClient, 'finance');

export const useCreateBill = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('financeCommon');

  return useMutation({
    mutationFn: (data: CreateBillRequest): Promise<CreateBillResponse> =>
      billClient.createBill(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['finance', 'bills'] });
      toast.success(t('financeBills.created'));
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    },
  });
};

export const useUpdateBill = () => {
  const queryClient = useQueryClient();
  const { t } = useTranslation('financeCommon');

  return useMutation({
    mutationFn: (data: UpdateBillRequest): Promise<UpdateBillResponse> =>
      billClient.updateBill(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['finance', 'bills'] });
    },
    onError: (error) => {
      toast.error(getOneKitErrorMessage(error, t('financeErrors.generic')));
    },
  });
};
