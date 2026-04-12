import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
import { unwrap } from '@voxire/api';
import { insightsService } from '../../shared/services/insightsService';
import {
  customerAnalyticsSchema,
  topCustomerSchema,
  customerChurnSchema,
  customerLtvSchema,
} from '../../shared/schemas/customers';

export function useCustomerAnalytics() {
  return useQuery({
    queryKey: ['insights', 'customers'],
    queryFn: async () => {
      const res = await unwrap(insightsService.getCustomerAnalytics());
      return customerAnalyticsSchema.parse(res);
    },
  });
}

export function useTopCustomers(limit?: number) {
  return useQuery({
    queryKey: ['insights', 'customers-top', limit],
    queryFn: async () => {
      const res = await unwrap(insightsService.getTopCustomers({ limit }));
      const raw = (res as Record<string, unknown>)?.customers ?? res;
      return z.array(topCustomerSchema).parse(Array.isArray(raw) ? raw : []);
    },
  });
}

export function useCustomerChurn() {
  return useQuery({
    queryKey: ['insights', 'customers-churn'],
    queryFn: async () => {
      const res = await unwrap(insightsService.getCustomerChurn());
      return customerChurnSchema.parse(res);
    },
  });
}

export function useCustomerLtv() {
  return useQuery({
    queryKey: ['insights', 'customers-ltv'],
    queryFn: async () => {
      const res = await unwrap(insightsService.getCustomerLtv());
      return customerLtvSchema.parse(res);
    },
  });
}
