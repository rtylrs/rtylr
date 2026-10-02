import {
  LayoutDashboard,
  Package,
  Folder,
  Tag,
  Layers,
  Tags,
  Barcode,
  Warehouse,
  Truck,
  ClipboardList,
  PackageCheck,
  CalendarX,
  AlertTriangle,
  Sparkles,
  DollarSign,
  ShoppingCart,
  Users,
} from 'lucide-react';
import { canEffectiveAccess, hasEffectiveAppAccess } from '@voxire/permissions';
import type { StoredUser } from '../../shared/stores/userStore';

export const erpNavItems = [
  { key: 'dashboard', path: '/erp', icon: LayoutDashboard, permission: 'erp:inventory:view' },
  { key: 'inventory', path: '/erp/inventory', icon: Package, permission: 'erp:inventory:view' },
  { key: 'categories', path: '/erp/categories', icon: Folder, permission: 'erp:products:view' },
  { key: 'brands', path: '/erp/brands', icon: Tag, permission: 'erp:products:view' },
  { key: 'variants', path: '/erp/variants', icon: Layers, permission: 'erp:products:view' },
  { key: 'priceLists', path: '/erp/price-lists', icon: Tags, permission: 'erp:products:view' },
  { key: 'barcodes', path: '/erp/barcodes', icon: Barcode, permission: 'erp:products:view' },
  {
    key: 'warehouses',
    path: '/erp/warehouses',
    icon: Warehouse,
    permission: 'erp:warehouses:manage',
  },
  { key: 'stockTransfers', path: '/erp/transfers', icon: Truck, permission: 'erp:inventory:view' },
  {
    key: 'stockTakes',
    path: '/erp/stock-takes',
    icon: ClipboardList,
    permission: 'erp:inventory:view',
  },
  {
    key: 'receiving',
    path: '/erp/receiving',
    icon: PackageCheck,
    permission: 'erp:purchasing:view',
  },
  { key: 'expiry', path: '/erp/expiry', icon: CalendarX, permission: 'erp:inventory:view' },
  {
    key: 'lowStockAlerts',
    path: '/erp/alerts',
    icon: AlertTriangle,
    permission: 'erp:inventory:view',
  },
  { key: 'reorder', path: '/erp/reorder', icon: Sparkles, permission: 'erp:inventory:view' },
  { key: 'valuation', path: '/erp/valuation', icon: DollarSign, permission: 'erp:inventory:view' },
  {
    key: 'purchaseOrders',
    path: '/erp/purchasing',
    icon: ShoppingCart,
    permission: 'erp:purchasing:view',
  },
  { key: 'suppliers', path: '/erp/suppliers', icon: Users, permission: 'erp:suppliers:view' },
] as const;

export function canAccessErpPath(user: StoredUser | null, pathname: string): boolean {
  const access = user?.effectiveAccess;
  if (!hasEffectiveAppAccess(access, 'erp')) return false;
  if (access?.isOwner || access?.isAdmin || pathname === '/erp') return true;
  if (pathname === '/erp/settings' || pathname === '/settings/erp')
    return canEffectiveAccess(access, 'erp:settings:view', 'erp');
  const permission =
    pathname === '/erp/products' || pathname.startsWith('/erp/products/')
      ? 'erp:products:view'
      : erpNavItems.find(
          (item) =>
            pathname === item.path || (item.path !== '/erp' && pathname.startsWith(`${item.path}/`))
        )?.permission;
  return !!permission && canEffectiveAccess(access, permission, 'erp');
}
export function getErpHomePath(user: StoredUser | null): string {
  if (canEffectiveAccess(user?.effectiveAccess, 'erp:inventory:view', 'erp')) return '/erp';
  if (canEffectiveAccess(user?.effectiveAccess, 'erp:products:view', 'erp')) return '/erp/products';
  return (
    erpNavItems.slice(1).find((item) => canAccessErpPath(user, item.path))?.path ??
    (canAccessErpPath(user, '/settings/erp') ? '/settings/erp' : '/')
  );
}
