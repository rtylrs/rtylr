/** Human label + accent color per dash-app category, mirrored across the
 *  Caret quick-action surfaces. */
import { Building2, Users, Package, Wallet, ListChecks, CreditCard } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';

export interface CategoryInfo {
  id: 'kpi' | 'crm' | 'inventory' | 'team' | 'orders' | 'finance';
  label: string;
  icon: LucideIcon;
  color: string;
}

export const CATEGORY_INFO: Record<CategoryInfo['id'], CategoryInfo> = {
  kpi: { id: 'kpi', label: 'KPIs', icon: Building2, color: '#64c5b9' },
  crm: { id: 'crm', label: 'Customers', icon: Users, color: '#a855f7' },
  inventory: { id: 'inventory', label: 'Inventory', icon: Package, color: '#10b981' },
  team: { id: 'team', label: 'Team', icon: ListChecks, color: '#0ea5e9' },
  orders: { id: 'orders', label: 'Orders', icon: Wallet, color: '#f59e0b' },
  finance: { id: 'finance', label: 'Finance', icon: CreditCard, color: '#f43f5e' },
};
