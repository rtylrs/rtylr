import { createOneKitClient } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import { ProductServiceClient } from '@voxire/voxie/erp/product/v1/client';
import type { GetDashboardResponse } from '@voxire/voxie/erp/product/v1/types';

const productClient = createOneKitClient(ProductServiceClient, 'erp');

export type DashboardStats = GetDashboardResponse;

export const useDashboardStats = (): UseQueryResult<GetDashboardResponse> => {
  return useQuery<GetDashboardResponse>({
    queryKey: ['erp-dashboard-stats'],
    queryFn: () => productClient.getDashboard({}),
  });
};
