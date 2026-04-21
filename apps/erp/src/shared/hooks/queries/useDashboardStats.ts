import { unwrap } from '@voxire/api';
import { useQuery } from '@tanstack/react-query';
import type { UseQueryResult } from '@tanstack/react-query';
import type { PurchaseOrder } from '@voxire/contracts/erp/v1/purchase_orders';
import { safe } from '../../lib/api';

// No proto type yet — define locally.
export interface LowStockDashboardItem {
  name: string;
  sku: string;
  stock: number;
  minStock: number;
  unit: string;
}

export interface DashboardStats {
  totalProducts: number;
  lowStockCount: number;
  pendingOrders: number;
  totalSuppliers: number;
  lowStockItems: LowStockDashboardItem[];
  recentOrders: PurchaseOrder[];
}

export const useDashboardStats = (): UseQueryResult<DashboardStats> => {
  return useQuery<DashboardStats>({
    queryKey: ['erp-dashboard-stats'],
    queryFn: async (): Promise<DashboardStats> => {
      const res = await unwrap(safe.get<DashboardStats>('/erp/dashboard'));
      return res as DashboardStats;
    },
  });
};
