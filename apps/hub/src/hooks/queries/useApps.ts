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

export const ALL_APPS: RtylrApp[] = [
  { key: 'pos',       label: 'Point of Sale',  description: 'Floor & counter sales, table management, kitchen display.', icon: MonitorSpeaker, color: '#4583C4', bg: 'rgba(69,131,196,0.1)',  href: 'https://pos.rtylr.com',     category: 'sales',      status: 'active' },
  { key: 'erp',       label: 'ERP',            description: 'Inventory, purchasing, suppliers, cost of goods.',           icon: Package,        color: '#10b981', bg: 'rgba(16,185,129,0.1)',  href: 'https://erp.rtylr.com',     category: 'operations', status: 'active' },
  { key: 'hr',        label: 'HR',             description: 'Staff, shifts, payroll, timesheets.',                        icon: UserCheck,      color: '#f59e0b', bg: 'rgba(245,158,11,0.1)',  href: 'https://hr.rtylr.com',      category: 'operations', status: 'active' },
  { key: 'crm',       label: 'CRM',            description: 'Customer profiles, loyalty programs, campaigns.',            icon: Users,          color: '#8b5cf6', bg: 'rgba(139,92,246,0.1)',  href: 'https://crm.rtylr.com',     category: 'growth',     status: 'active' },
  { key: 'finance',   label: 'Finance',        description: 'Invoicing, expenses, P&L, tax reporting.',                   icon: CreditCard,     color: '#ef4444', bg: 'rgba(239,68,68,0.1)',   href: 'https://finance.rtylr.com', category: 'finance',    status: 'active' },
  { key: 'tracker',   label: 'Tracker',        description: 'Projects, tasks, subtasks, team workspaces.',                icon: Layers,         color: '#64C5B9', bg: 'rgba(100,197,185,0.1)', href: 'https://tracker.rtylr.com', category: 'core',       status: 'active' },
  { key: 'storefront',label: 'Storefront',     description: 'Your public online storefront and menu.',                    icon: Store,          color: '#ec4899', bg: 'rgba(236,72,153,0.1)',  href: 'https://storefront.rtylr.com', category: 'sales',   status: 'beta' },
  { key: 'analytics', label: 'Analytics',      description: 'Business-wide reports and dashboards.',                      icon: BarChart3,      color: '#06b6d4', bg: 'rgba(6,182,212,0.1)',   href: '/analytics',                category: 'core',       status: 'active' },
  { key: 'market',    label: 'Marketplace',    description: 'Install third-party apps and extensions.',                   icon: ShoppingBag,    color: '#a855f7', bg: 'rgba(168,85,247,0.1)',  href: '/marketplace',              category: 'growth',     status: 'coming_soon' },
  { key: 'web',       label: 'Website',        description: 'Drag-and-drop website builder.',                              icon: Globe,          color: '#0ea5e9', bg: 'rgba(14,165,233,0.1)',  href: 'https://web.rtylr.com',     category: 'growth',     status: 'coming_soon' },
];

export const useApps = () => {
  return useQuery<RtylrApp[]>({
    queryKey: ['apps'],
    queryFn: async () => {
      try {
        const res = await api.get('/apps');
        const remote = (res.data?.apps ?? res.data) as RtylrApp[] | undefined;
        if (remote && Array.isArray(remote) && remote.length > 0) {
          // merge remote status with local definitions
          return ALL_APPS.map((a) => {
            const r = remote.find((x) => x.key === a.key);
            return r ? { ...a, ...r, icon: a.icon } : a;
          });
        }
        return ALL_APPS;
      } catch {
        return ALL_APPS;
      }
    },
    staleTime: 60_000,
  });
};
