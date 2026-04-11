import { useQuery } from '@tanstack/react-query';
import { unwrap } from '@voxire/api';
import { insightsService } from '../services/insightsService';

export function useCustomerAnalytics() {
  return useQuery({
    queryKey: ['insights', 'customers'],
    queryFn: () => unwrap(insightsService.getCustomerAnalytics()),
  });
}

export function useTopCustomers(limit?: number) {
  return useQuery({
    queryKey: ['insights', 'customers-top', limit],
    queryFn: () => unwrap(insightsService.getTopCustomers({ limit })),
  });
}

export function useCustomerChurn() {
  return useQuery({
    queryKey: ['insights', 'customers-churn'],
    queryFn: () => unwrap(insightsService.getCustomerChurn()),
  });
}

export function useCustomerLtv() {
  return useQuery({
    queryKey: ['insights', 'customers-ltv'],
    queryFn: () => unwrap(insightsService.getCustomerLtv()),
  });
}
