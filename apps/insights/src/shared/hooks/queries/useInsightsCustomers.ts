import { createOneKitClient, toFiniteNumber } from '@voxire/api';
import { useQuery, type UseQueryResult } from '@tanstack/react-query';
import { CustomerServiceClient } from '@voxire/voxie/insights/customer/v1/client';
import { resolveDateRange, type DateRangeParams } from '../../utils/dateRange';
import type {
  CustomerKPIResponse,
  TopCustomersResponse,
  CustomerLTVResponse,
  CustomerChurnResponse,
} from './types';

const customerClient = createOneKitClient(CustomerServiceClient, 'insights');

const majorToMinor = (value: unknown): number => Math.round(toFiniteNumber(value) * 100);

export const useCustomerKPIs = (params?: DateRangeParams): UseQueryResult<CustomerKPIResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<CustomerKPIResponse>({
    queryKey: ['insights-customers', { from, to }],
    queryFn: async (): Promise<CustomerKPIResponse> => {
      const res = await customerClient.getCustomerAnalytics({ period: `${from},${to}` });
      const analytics = res.analytics;
      const total = toFiniteNumber(analytics?.total);
      const retentionRate = toFiniteNumber(analytics?.retentionRate);
      const churnRate = toFiniteNumber(analytics?.churnRate);
      return {
        total,
        newThisMonth: toFiniteNumber(analytics?.newThisPeriod),
        active: Math.round((total * retentionRate) / 100),
        churned: Math.round((total * churnRate) / 100),
      };
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
  const { from, to } = resolveDateRange(undefined);
  return useQuery<TopCustomersResponse>({
    queryKey: ['insights-customers-top', { limit, from, to }],
    queryFn: async (): Promise<TopCustomersResponse> => {
      const res = await customerClient.getTopCustomers({
        page: 1,
        limit,
        sortBy: 'totalSpent',
        sortDir: 'desc',
        period: `${from},${to}`,
      });
      return {
        customers: (res.customers ?? []).map((entry) => ({
          customerId: toFiniteNumber(entry.id),
          name: entry.name ?? '',
          email: '',
          lifetimeSpendMinorUnits: majorToMinor(entry.totalSpent),
          orderCount: toFiniteNumber(entry.orderCount),
        })),
      };
    },
  });
};

export const useCustomerLTV = (): UseQueryResult<CustomerLTVResponse> => {
  return useQuery<CustomerLTVResponse>({
    queryKey: ['insights-customers-ltv'],
    queryFn: async (): Promise<CustomerLTVResponse> => {
      const res = await customerClient.getLTVAnalysis({});
      const segments = res.analysis?.segments ?? [];
      return {
        buckets: segments.map((entry) => {
          const avg = majorToMinor(entry.avgLtv);
          return {
            label: entry.segment ?? '',
            lowerMinorUnits: avg,
            upperMinorUnits: avg,
            customerCount: toFiniteNumber(entry.customerCount),
          };
        }),
      };
    },
  });
};

export const useCustomerChurn = (
  params?: DateRangeParams
): UseQueryResult<CustomerChurnResponse> => {
  const { from, to } = resolveDateRange(params);
  return useQuery<CustomerChurnResponse>({
    queryKey: ['insights-customers-churn', { from, to }],
    queryFn: async (): Promise<CustomerChurnResponse> => {
      const res = await customerClient.getChurnAnalysis({ period: `${from},${to}` });
      const analysis = res.analysis;
      return {
        windows: [
          {
            windowDays: 30,
            churnedCount: toFiniteNumber(analysis?.churnedCount),
            activeCount: 0,
            churnPct: toFiniteNumber(analysis?.churnRate),
          },
        ],
      };
    },
  });
};
