import { useQuery } from '@tanstack/react-query';
import { insightsService } from '../services/insightsService';

export function useCustomerAnalytics() {
  return useQuery({
    queryKey: ['insights', 'customers'],
    queryFn: () => insightsService.getCustomerAnalytics().then((r) => r.data),
  });
}

export function useTopCustomers(limit?: number) {
  return useQuery({
    queryKey: ['insights', 'customers-top', limit],
    queryFn: () => insightsService.getTopCustomers({ limit }).then((r) => r.data),
  });
}

export function useCustomerChurn() {
  return useQuery({
    queryKey: ['insights', 'customers-churn'],
    queryFn: () => insightsService.getCustomerChurn().then((r) => r.data),
  });
}

export function useCustomerLtv() {
  return useQuery({
    queryKey: ['insights', 'customers-ltv'],
    queryFn: () => insightsService.getCustomerLtv().then((r) => r.data),
  });
}
