import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { z } from 'zod';
import { supportService } from '../../shared/services/supportService';
import { supportCustomerSchema } from '../../shared/schemas/customers';
import { ticketSchema } from '../../shared/schemas/tickets';

export function useCustomers(params?: { search?: string; limit?: number; offset?: number }) {
  return useQuery({
    queryKey: ['customers', params],
    queryFn: async () => {
      const res = await unwrap(supportService.getCustomers(params));
      const raw = res ?? [];
      return z.array(supportCustomerSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCustomer(id: number) {
  return useQuery({
    queryKey: ['customer', id],
    queryFn: async () => {
      const res = await unwrap(supportService.getCustomer(id));
      return supportCustomerSchema.parse(res);
    },
    enabled: !!id,
  });
}

export function useCustomerTickets(customerId: number) {
  return useQuery({
    queryKey: ['customer-tickets', customerId],
    queryFn: async () => {
      const res = await unwrap(supportService.getCustomerTickets(customerId));
      const raw = res ?? [];
      return z.array(ticketSchema).parse(Array.isArray(raw) ? raw : []);
    },
    enabled: !!customerId,
  });
}
