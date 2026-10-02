import {
  BarChart2,
  BookOpen,
  Building2,
  Calculator,
  ClipboardList,
  FileCheck,
  FileText,
  LayoutDashboard,
  LineChart,
  Receipt,
  Repeat,
  Shuffle,
  Tag,
  TrendingUp,
  Wallet,
} from 'lucide-react';
import { canEffectiveAccess, hasEffectiveAppAccess } from '@voxire/permissions';
import type { StoredUser } from '../../shared/stores/userStore';
export const financeNavItems = [
  { key: 'dashboard', path: '/finance', icon: LayoutDashboard, permission: 'finance:reports:view' },
  {
    key: 'invoices',
    path: '/finance/invoices',
    icon: FileText,
    permission: 'finance:invoices:view',
  },
  {
    key: 'recurring',
    path: '/finance/invoices/recurring',
    icon: Repeat,
    permission: 'finance:recurring_invoices:view',
  },
  {
    key: 'quotes',
    path: '/finance/quotes',
    icon: ClipboardList,
    permission: 'finance:quotes:view',
  },
  {
    key: 'expenses',
    path: '/finance/expenses',
    icon: Receipt,
    permission: 'finance:expenses:view',
  },
  {
    key: 'categories',
    path: '/finance/categories',
    icon: Tag,
    permission: 'finance:expenses:view',
  },
  { key: 'bills', path: '/finance/bills', icon: Wallet, permission: 'finance:expenses:view' },
  {
    key: 'bankFeeds',
    path: '/finance/bank',
    icon: Building2,
    permission: 'finance:bank_feeds:view',
  },
  {
    key: 'reconcile',
    path: '/finance/reconcile',
    icon: Shuffle,
    permission: 'finance:reconciliation:manage',
  },
  { key: 'pl', path: '/finance/pl', icon: TrendingUp, permission: 'finance:reports:view' },
  {
    key: 'balanceSheet',
    path: '/finance/reports/balance-sheet',
    icon: BarChart2,
    permission: 'finance:reports:view',
  },
  {
    key: 'cashFlow',
    path: '/finance/reports/cash-flow',
    icon: LineChart,
    permission: 'finance:reports:view',
  },
  {
    key: 'taxSummary',
    path: '/finance/reports/tax',
    icon: FileCheck,
    permission: 'finance:reports:view',
  },
  { key: 'journal', path: '/finance/journal', icon: BookOpen, permission: 'finance:accounts:view' },
  {
    key: 'accounts',
    path: '/finance/accounts',
    icon: FileText,
    permission: 'finance:accounts:view',
  },
  { key: 'taxConfig', path: '/finance/tax', icon: Calculator, permission: 'finance:tax:configure' },
] as const;
export function canAccessFinancePath(user: StoredUser | null, path: string): boolean {
  const access = user?.effectiveAccess;
  if (!hasEffectiveAppAccess(access, 'finance')) return false;
  if (access?.isOwner || access?.isAdmin || path === '/finance') return true;
  if (path === '/settings/finance' || path === '/finance/settings') return false;
  const permission =
    path === '/finance/invoices/new'
      ? 'finance:invoices:create'
      : path === '/finance/expenses/new'
        ? 'finance:expenses:create'
        : financeNavItems
            .filter(
              (item) =>
                item.path !== '/finance' && (path === item.path || path.startsWith(`${item.path}/`))
            )
            .sort((a, b) => b.path.length - a.path.length)[0]?.permission;
  return !!permission && canEffectiveAccess(access, permission, 'finance');
}
export function getFinanceHomePath(user: StoredUser | null): string {
  return (
    financeNavItems.find((item) =>
      canEffectiveAccess(user?.effectiveAccess, item.permission, 'finance')
    )?.path ?? '/'
  );
}
