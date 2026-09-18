import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { z } from 'zod';
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
  CalendarCheck,
  Headphones,
} from 'lucide-react';
import { APP_URLS } from '@voxire/config';
import type { AppId } from '@voxire/auth';
import { appClient } from '../../lib/clients';

function appLaunchUrl(appId: string): string {
  if (appId in APP_URLS) {
    return APP_URLS[appId as AppId];
  }
  return '#';
}
import { appContractSchema } from '../../schemas/apps';

export interface RtylrApp {
  key: string;
  label: string;
  description: string;
  href: string;
  color: string;
  bg: string;
  icon: React.ComponentType<{ size?: number; style?: React.CSSProperties }>;
  category: 'core' | 'sales' | 'operations' | 'finance' | 'growth';
  status: 'active' | 'beta' | 'coming_soon';
  /** True when the app is available but not included in the owner's current plan. */
  locked?: boolean;
}

/** Icon lookup used to hydrate server data with client-side icons */
const ICON_MAP: Record<
  string,
  React.ComponentType<{ size?: number; style?: React.CSSProperties }>
> = {
  pos: MonitorSpeaker,
  erp: Package,
  hr: UserCheck,
  crm: Users,
  finance: CreditCard,
  flow: Layers,
  store: Store,
  book: CalendarCheck,
  support: Headphones,
  insights: BarChart3,

  storefront: Store,
  analytics: BarChart3,
  market: ShoppingBag,
  web: Globe,
};

const APP_META: Record<
  string,
  { href: string; color: string; bg: string; category: RtylrApp['category'] }
> = {
  pos: {
    href: appLaunchUrl('pos'),
    color: '#4583C4',
    bg: 'rgba(69,131,196,0.12)',
    category: 'core',
  },
  erp: {
    href: appLaunchUrl('erp'),
    color: '#10b981',
    bg: 'rgba(16,185,129,0.12)',
    category: 'operations',
  },
  hr: {
    href: appLaunchUrl('hr'),
    color: '#f59e0b',
    bg: 'rgba(245,158,11,0.12)',
    category: 'operations',
  },
  crm: {
    href: appLaunchUrl('crm'),
    color: '#8b5cf6',
    bg: 'rgba(139,92,246,0.12)',
    category: 'growth',
  },
  finance: {
    href: appLaunchUrl('finance'),
    color: '#ef4444',
    bg: 'rgba(239,68,68,0.12)',
    category: 'finance',
  },
  flow: {
    href: appLaunchUrl('flow'),
    color: '#64C5B9',
    bg: 'rgba(100,197,185,0.12)',
    category: 'core',
  },
  store: {
    href: 'https://store.rtylr.com',
    color: '#14b8a6',
    bg: 'rgba(20,184,166,0.12)',
    category: 'sales',
  },
  book: {
    href: 'https://book.rtylr.com',
    color: '#0ea5e9',
    bg: 'rgba(14,165,233,0.12)',
    category: 'operations',
  },
  support: {
    href: 'https://support.rtylr.com',
    color: '#4583C4',
    bg: 'rgba(69,131,196,0.12)',
    category: 'growth',
  },
  insights: {
    href: appLaunchUrl('insights'),
    color: '#6366f1',
    bg: 'rgba(99,102,241,0.12)',
    category: 'core',
  },
};

export const useApps = () => {
  return useQuery<RtylrApp[]>({
    queryKey: ['apps'],
    queryFn: async () => {
      const res = await appClient.listApps({});
      const remote = Array.isArray(res) ? res : (res?.apps ?? []);
      if (!Array.isArray(remote)) return [];
      const validated = z.array(appContractSchema).parse(remote);
      return validated.map((a) => {
        const meta = APP_META[a.id] ?? {
          href: '#',
          color: '#888',
          bg: 'rgba(136,136,136,0.12)',
          category: 'core' as const,
        };
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
