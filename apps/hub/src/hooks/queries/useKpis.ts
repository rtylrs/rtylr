import { useQuery } from '@tanstack/react-query';
import api from '../../lib/api';

export interface HubKpis {
  revenueToday: number;
  revenueTrendPct: number;
  ordersToday: number;
  ordersTrendPct: number;
  headcount: number;
  activeStaff: number;
  lowStockItems: number;
  openInvoices: number;
  newCustomers: number;
  apiHealthPct: number;
  revenueSeries: number[];
  ordersSeries: number[];
}

export const useKpis = () => {
  return useQuery<HubKpis>({
    queryKey: ['hub-kpis'],
    queryFn: async () => {
      const res = await api.get('/hub/kpis');
      return res.data as HubKpis;
    },
    staleTime: 30_000,
  });
};
