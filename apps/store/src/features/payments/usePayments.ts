import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { storeApi, PaymentMethod } from '../../shared/services/storeService';
import { paymentMethodSchema } from '../../shared/schemas/payments';

export function usePaymentMethods() {
  return useQuery({
    queryKey: ['store-payment-methods'],
    queryFn: async () => {
      const res = await unwrap(storeApi.listPaymentMethods());
      const raw = (res as any)?.methods ?? res;
      return z.array(paymentMethodSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useUpdatePaymentMethod() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<PaymentMethod> & { id: number }) =>
      unwrap(storeApi.updatePaymentMethod(data)),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['store-payment-methods'] }),
  });
}
