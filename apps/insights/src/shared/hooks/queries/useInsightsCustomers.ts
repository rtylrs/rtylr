import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import type {
  CustomerKPIResponse,
  TopCustomersResponse,
  CustomerLTVResponse,
  CustomerChurnResponse,
} from './types';
import {
  mapCustomerAnalytics,
  mapCustomerChurn,
  mapCustomerLTV,
  mapTopCustomers,
} from '../../utils/onekitMappers';

export const useCustomerKPIs = (): UseQueryResult<CustomerKPIResponse> => {
  return useQuery<CustomerKPIResponse>({
    queryKey: ['insights-customers'],
    queryFn: async (): Promise<CustomerKPIResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/customer/v1/analytics')
      );
      return mapCustomerAnalytics(res as Record<string, unknown>);
    },
  });
};

interface TopCustomersParams {
  limit?: number;
}

export const useTopCustomers = (
  params?: TopCustomersParams
): UseQueryResult<TopCustomersResponse> => {
  const limit = params?.limit ?? 10;
  return useQuery<TopCustomersResponse>({
    queryKey: ['insights-customers-top', { limit }],
    queryFn: async (): Promise<TopCustomersResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/customer/v1/top', { params: { limit } })
      );
      return mapTopCustomers(res as Record<string, unknown>);
    },
  });
};

export const useCustomerLTV = (): UseQueryResult<CustomerLTVResponse> => {
  return useQuery<CustomerLTVResponse>({
    queryKey: ['insights-customers-ltv'],
    queryFn: async (): Promise<CustomerLTVResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/customer/v1/ltv')
      );
      return mapCustomerLTV(res as Record<string, unknown>);
    },
  });
};

export const useCustomerChurn = (): UseQueryResult<CustomerChurnResponse> => {
  return useQuery<CustomerChurnResponse>({
    queryKey: ['insights-customers-churn'],
    queryFn: async (): Promise<CustomerChurnResponse> => {
      const res = await unwrap(
        safe.get<Record<string, unknown>>('/api/insights/customer/v1/churn')
      );
      return mapCustomerChurn(res as Record<string, unknown>);
    },
  });
};
