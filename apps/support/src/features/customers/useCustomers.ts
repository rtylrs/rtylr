import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { supportService } from '../../shared/services/supportService';

export function useCustomers(params?: { search?: string; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ['customers', params],
    queryFn: () => unwrap(supportService.getCustomers(params)),
  });
}

export function useCustomer(id: number) {
  return useQuery({
    queryKey: ['customer', id],
    queryFn: () => unwrap(supportService.getCustomer(id)),
    enabled: !!id,
  });
}

export function useCustomerTickets(customerId: number) {
  return useQuery({
    queryKey: ['customer-tickets', customerId],
    queryFn: () => unwrap(supportService.getCustomerTickets(customerId)),
    enabled: !!customerId,
  });
}
