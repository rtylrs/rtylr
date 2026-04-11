import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { storeApi, PaymentMethod } from '../services/storeService';

export function usePaymentMethods() {
  return useQuery({
    queryKey: ['store-payment-methods'],
    queryFn: async () => {
      const r = await storeApi.listPaymentMethods();
      return r.data ?? [];
    },
  });
}

export function useUpdatePaymentMethod() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<PaymentMethod> & { id: number }) => storeApi.updatePaymentMethod(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-payment-methods'] }),
  });
}
