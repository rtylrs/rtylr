import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { bookService } from '../services/bookService';

export interface CustomerFilters {
  search?: string;
  limit?: number;
  offset?: number;
}

export function useCustomers(filters?: CustomerFilters) {
  return useQuery({
    queryKey: ['customers', filters],
    queryFn: () => unwrap(bookService.getCustomers(filters)),
  });
}

export function useCustomer(id: number | undefined) {
  return useQuery({
    queryKey: ['customer', id],
    enabled: !!id,
    queryFn: () => unwrap(bookService.getCustomer(id!)),
  });
}
