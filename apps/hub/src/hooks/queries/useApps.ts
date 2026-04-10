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

const APP_META: Record<string, { href: string; color: string; bg: string; category: RtylrApp['category'] }> = {
  pos:     { href: 'https://pos.rtylr.com',     color: '#4583C4', bg: 'rgba(69,131,196,0.12)',  category: 'core' },
  erp:     { href: 'https://erp.rtylr.com',     color: '#10b981', bg: 'rgba(16,185,129,0.12)',  category: 'operations' },
  hr:      { href: 'https://hr.rtylr.com',      color: '#f59e0b', bg: 'rgba(245,158,11,0.12)',  category: 'operations' },
  crm:     { href: 'https://crm.rtylr.com',     color: '#8b5cf6', bg: 'rgba(139,92,246,0.12)',  category: 'growth' },
  finance: { href: 'https://finance.rtylr.com', color: '#ef4444', bg: 'rgba(239,68,68,0.12)',   category: 'finance' },
  flow:    { href: 'https://flow.rtylr.com',    color: '#64C5B9', bg: 'rgba(100,197,185,0.12)', category: 'core' },
};

export const useApps = () => {
  return useQuery<RtylrApp[]>({
    queryKey: ['apps'],
    queryFn: async () => {
      const res = await api.get('/apps');
      const remote = res.data?.apps ?? res.data;
      if (!Array.isArray(remote)) return [];
      return remote.map((a: any) => {
        const meta = APP_META[a.id] ?? { href: '#', color: '#888', bg: 'rgba(136,136,136,0.12)', category: 'core' as const };
        return {
          key: a.id,
          label: a.name,
          description: a.description ?? '',
          href: meta.href,
          color: meta.color,
          bg: meta.bg,
          icon: ICON_MAP[a.id] ?? Layers,
          category: meta.category,
          status: (a.isEnabled ? 'active' : 'coming_soon') as RtylrApp['status'],
        };
      });
    },
    staleTime: 60_000,
  });
};
