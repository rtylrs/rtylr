import { useQuery } from '@tanstack/react-query';
import { supportService } from '../services/supportService';

export function useCustomers(params?: { search?: string; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ['customers', params],
    queryFn: () => supportService.getCustomers(params).then((r) => r.data),
  });
}

export function useCustomer(id: number) {
  return useQuery({
    queryKey: ['customer', id],
    queryFn: () => supportService.getCustomer(id).then((r) => r.data),
    enabled: !!id,
  });
}

export function useCustomerTickets(customerId: number) {
  return useQuery({
    queryKey: ['customer-tickets', customerId],
    queryFn: () => supportService.getCustomerTickets(customerId).then((r) => r.data),
    enabled: !!customerId,
  });
}
