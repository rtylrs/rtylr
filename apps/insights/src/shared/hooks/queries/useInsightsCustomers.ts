import { unwrap } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { safe } from '../../lib/api';
import type {
  CustomerKPIResponse,
  TopCustomersResponse,
  CustomerLTVResponse,
  CustomerChurnResponse,
} from './types';

export const useCustomerKPIs = (): UseQueryResult<CustomerKPIResponse> => {
  return useQuery<CustomerKPIResponse>({
    queryKey: ['insights-customers'],
    queryFn: async (): Promise<CustomerKPIResponse> => {
      const res = await unwrap(safe.get<CustomerKPIResponse>('/insights/customers'));
      return res as CustomerKPIResponse;
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
        safe.get<TopCustomersResponse>('/insights/customers/top', { params: { limit } })
      );
      return res as TopCustomersResponse;
    },
  });
};

export const useCustomerLTV = (): UseQueryResult<CustomerLTVResponse> => {
  return useQuery<CustomerLTVResponse>({
    queryKey: ['insights-customers-ltv'],
    queryFn: async (): Promise<CustomerLTVResponse> => {
      const res = await unwrap(safe.get<CustomerLTVResponse>('/insights/customers/ltv'));
      return res as CustomerLTVResponse;
    },
  });
};

export const useCustomerChurn = (): UseQueryResult<CustomerChurnResponse> => {
  return useQuery<CustomerChurnResponse>({
    queryKey: ['insights-customers-churn'],
    queryFn: async (): Promise<CustomerChurnResponse> => {
      const res = await unwrap(safe.get<CustomerChurnResponse>('/insights/customers/churn'));
      return res as CustomerChurnResponse;
    },
  });
};
