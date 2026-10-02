import {
  LayoutDashboard,
  DollarSign,
  Users,
  Package,
  Activity,
  Wallet,
  FileText,
} from 'lucide-react';
import type { PortalNavItem } from '../../shell/navigation';

export const insightsNavItems: PortalNavItem[] = [
  { key: 'overview', path: '/insights', icon: LayoutDashboard },
  { key: 'revenue', path: '/insights/revenue', icon: DollarSign },
  { key: 'customers', path: '/insights/customers', icon: Users },
  { key: 'products', path: '/insights/products', icon: Package },
  { key: 'operations', path: '/insights/operations', icon: Activity },
  { key: 'financial', path: '/insights/financial', icon: Wallet },
  { key: 'reports', path: '/insights/reports', icon: FileText },
];
