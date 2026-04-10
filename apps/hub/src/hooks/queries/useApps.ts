import { useQuery } from '@tanstack/react-query';
import {
  MonitorSpeaker,
  Package,
  UserCheck,
  Users,
  CreditCard,
  Layers,
  Store,
  BarChart3,
  ShoppingBag,
  Globe,
} from 'lucide-react';
import api from '../../lib/api';

export interface RtylrApp {
  key: string;
  label: string;
  description: string;
  href: string;
  color: string;
  bg: string;
  icon: any;
  category: 'core' | 'sales' | 'operations' | 'finance' | 'growth';
  status: 'active' | 'beta' | 'coming_soon';
}

/** Icon lookup used to hydrate server data with client-side icons */
const ICON_MAP: Record<string, any> = {
  pos: MonitorSpeaker,
  erp: Package,
  hr: UserCheck,
  crm: Users,
  finance: CreditCard,
  flow: Layers,
  storefront: Store,
  analytics: BarChart3,
  market: ShoppingBag,
  web: Globe,
};

export const useApps = () => {
  return useQuery<RtylrApp[]>({
    queryKey: ['apps'],
    queryFn: async () => {
      const res = await api.get('/apps');
      const remote = (res.data?.apps ?? res.data) as RtylrApp[];
      if (Array.isArray(remote)) {
        return remote.map((a) => ({
          ...a,
          icon: ICON_MAP[a.key] ?? Layers,
        }));
      }
      return [];
    },
    staleTime: 60_000,
  });
};
