import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { bookService } from '../../shared/services/bookService';
import { bookCustomerSchema } from '../../shared/schemas/customers';

export interface CustomerFilters {
  search?: string;
  limit?: number;
  offset?: number;
}

export function useCustomers(filters?: CustomerFilters) {
  return useQuery({
    queryKey: ['customers', filters],
    queryFn: async () => {
      const res = await unwrap(bookService.getCustomers(filters));
      const raw = (res as Record<string, unknown>)?.customers ?? res;
      return z.array(bookCustomerSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCustomer(id: number | undefined) {
  return useQuery({
    queryKey: ['customer', id],
    enabled: !!id,
    queryFn: async () => {
      const res = await unwrap(bookService.getCustomer(id!));
      return bookCustomerSchema.parse(res);
    },
  });
}
