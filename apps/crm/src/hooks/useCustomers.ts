import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { crmApi, Customer, CustomersParams } from '../services/crmApi';

export function useCustomers(params?: CustomersParams) {
  return useQuery({
    queryKey: ['customers', params],
    queryFn: () => crmApi.getCustomers(params),
  });
}

export function useCustomerDetail(id: number | undefined) {
  return useQuery({
    queryKey: ['customer', id],
    queryFn: () => crmApi.getCustomerDetail(id!),
    enabled: !!id,
  });
}

export function useCreateCustomer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (data: Partial<Customer>) => crmApi.createCustomer(data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  });
}

export function useUpdateCustomer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, data }: { id: number; data: Partial<Customer> }) => crmApi.updateCustomer(id, data),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  });
}

export function useDeleteCustomer() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => crmApi.deleteCustomer(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  });
}

export function useRecordVisit() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (customerId: number) => crmApi.recordVisit(customerId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['customers'] }),
  });
}
