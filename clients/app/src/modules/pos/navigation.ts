import {
  Monitor,
  ClipboardList,
  Receipt,
  Armchair,
  ChefHat,
  Users,
  Gift,
  BarChart3,
  WifiOff,
  Package,
  Layers,
  Sliders,
  Pause,
  Banknote,
  LogOut,
} from 'lucide-react';
import { hasEffectiveAppAccess, canEffectiveAccess } from '@voxire/permissions';
import type { StoredUser } from '../../shared/stores/userStore';
export const posNavItems = [
  { key: 'terminal', path: '/pos', icon: Monitor },
  { key: 'products', path: '/pos/menu', icon: ClipboardList },
  { key: 'orders', path: '/pos/orders', icon: Receipt },
  { key: 'tables', path: '/pos/tables', icon: Armchair, feature: 'tables' },
  { key: 'kitchen', path: '/pos/kds', icon: ChefHat, feature: 'kitchenDisplay' },
  { key: 'customers', path: '/pos/customer-lookup', icon: Users, feature: 'customers' },
  { key: 'giftCards', path: '/pos/gift-cards', icon: Gift, feature: 'giftCards' },
  { key: 'receipts', path: '/pos/receipts', icon: Receipt },
  { key: 'day', path: '/pos/day-summary', icon: BarChart3 },
  { key: 'modifiers', path: '/pos/modifiers', icon: Sliders, feature: 'modifiers' },
  { key: 'combos', path: '/pos/combos', icon: Layers, feature: 'combos' },
  { key: 'inventory', path: '/pos/inventory', icon: Package, feature: 'inventory' },
  { key: 'staff', path: '/pos/staff', icon: Users },
  { key: 'heldOrders', path: '/pos/held-orders', icon: Pause },
  { key: 'cashDrop', path: '/pos/cash-drop', icon: Banknote, feature: 'cashDrawer' },
  { key: 'closeSession', path: '/pos/close-session', icon: LogOut },
  { key: 'offline', path: '/pos/offline', icon: WifiOff },
] as const;
export function canAccessPosPath(user: StoredUser | null, path: string) {
  const access = user?.effectiveAccess;
  if (!hasEffectiveAppAccess(access, 'pos')) return false;
  if (path === '/settings/pos') return canEffectiveAccess(access, 'pos:terminals:manage', 'pos');
  return true;
}
