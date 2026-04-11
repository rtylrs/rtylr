import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { storeApi, PaymentMethod } from '../services/storeService';

export function usePaymentMethods() {
  return useQuery({
    queryKey: ['store-payment-methods'],
    queryFn: () => unwrap(storeApi.listPaymentMethods()),
  });
}

export function useUpdatePaymentMethod() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<PaymentMethod> & { id: number }) => unwrap(storeApi.updatePaymentMethod(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-payment-methods'] }),
  });
}
