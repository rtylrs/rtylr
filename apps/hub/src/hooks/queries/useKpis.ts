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

const MOCK_KPIS: HubKpis = {
  revenueToday: 4820,
  revenueTrendPct: 12.4,
  ordersToday: 127,
  ordersTrendPct: 4.1,
  headcount: 18,
  activeStaff: 11,
  lowStockItems: 4,
  openInvoices: 7,
  newCustomers: 23,
  apiHealthPct: 99.98,
  revenueSeries: [3100, 3400, 3850, 2900, 4100, 4600, 4820],
  ordersSeries: [80, 92, 110, 75, 120, 134, 127],
};

export const useKpis = () => {
  return useQuery<HubKpis>({
    queryKey: ['hub-kpis'],
    queryFn: async () => {
      try {
        const res = await api.get('/hub/kpis');
        return res.data as HubKpis;
      } catch {
        return MOCK_KPIS;
      }
    },
    staleTime: 30_000,
  });
};
