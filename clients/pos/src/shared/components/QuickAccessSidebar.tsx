import { usePOSPortal } from './PortalContext';
import React from 'react';
import type { LucideIcon } from 'lucide-react';
import {
  Armchair,
  Boxes,
  Calculator,
  ChartNoAxesCombined,
  ChefHat,
  CirclePause,
  ClipboardList,
  Gift,
  Monitor,
  ReceiptText,
  Settings,
  User,
  WifiOff,
} from 'lucide-react';
import { NavLink } from 'react-router-dom';
import { useSettingsStore } from '../stores/settingsStore';
import type { POSBusinessProfile, POSFeatureKey } from '../stores/settingsStore';

type QuickLink = {
  to: string;
  label: string | ((profile: POSBusinessProfile) => string);
  icon: LucideIcon;
  feature?: POSFeatureKey;
};

const LINKS: QuickLink[] = [
  { to: '/', label: 'Terminal', icon: Monitor },
  {
    to: '/menu',
    label: (profile) => (profile === 'retail' ? 'Products' : 'Menu'),
    icon: ClipboardList,
  },
  { to: '/inventory', label: 'Inventory', icon: Boxes, feature: 'inventory' },
  { to: '/orders', label: 'Orders', icon: ReceiptText },
  { to: '/tables', label: 'Tables', icon: Armchair, feature: 'tables' },
  { to: '/kds', label: 'KDS', icon: ChefHat, feature: 'kitchenDisplay' },
  { to: '/customer-lookup', label: 'Customers', icon: User, feature: 'customers' },
  { to: '/gift-cards', label: 'Gift Cards', icon: Gift, feature: 'giftCards' },
  { to: '/receipts', label: 'Receipts', icon: Calculator },
  { to: '/day-summary', label: 'Day', icon: ChartNoAxesCombined },
  { to: '/held-orders', label: 'Held', icon: CirclePause },
  { to: '/offline', label: 'Offline', icon: WifiOff },
  { to: '/settings', label: 'Settings', icon: Settings },
];

const QuickAccessSidebar: React.FC = () => {
  const embedded = usePOSPortal();
  const enabled = useSettingsStore((s) => s.quickAccessSidebar);
  const features = useSettingsStore((s) => s.features);
  const businessProfile = useSettingsStore((s) => s.businessProfile);
  if (embedded || !enabled) return null;
  return (
    <aside className="hidden md:flex shrink-0 w-[4.5rem] flex-col border-r border-pos-border bg-pos-card overflow-y-auto">
      {LINKS.filter((l) => !l.feature || features[l.feature]).map((l) => {
        const Icon = l.icon;
        const label = typeof l.label === 'function' ? l.label(businessProfile) : l.label;
        return (
          <NavLink
            key={l.to}
            to={l.to}
            end={l.to === '/'}
            className={({ isActive }) =>
              `flex flex-col items-center justify-center gap-0.5 py-3 text-[10px] font-medium transition-colors ${
                isActive
                  ? 'text-pos-text bg-white/8 border-l-2 border-pos-text'
                  : 'text-pos-muted hover:text-pos-text hover:bg-white/5'
              }`
            }
          >
            <Icon className="h-5 w-5" aria-hidden="true" strokeWidth={2} />
            <span>{label}</span>
          </NavLink>
        );
      })}
    </aside>
  );
};

export default QuickAccessSidebar;
