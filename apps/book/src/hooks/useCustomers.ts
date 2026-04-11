import { useQuery } from '@tanstack/react-query';
import { bookService } from '../services/bookService';

export interface CustomerFilters {
  search?: string;
  limit?: number;
  offset?: number;
}

export function useCustomers(filters?: CustomerFilters) {
  return useQuery({
    queryKey: ['customers', filters],
    queryFn: async () => {
      const res = await bookService.getCustomers(filters);
      return res.data?.data ?? res.data ?? [];
    },
  });
}

export function useCustomer(id: number | undefined) {
  return useQuery({
    queryKey: ['customer', id],
    enabled: !!id,
    queryFn: async () => {
      if (!id) return null;
      const res = await bookService.getCustomer(id);
      return res.data?.data ?? res.data ?? null;
    },
  });
}
