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

type QuickLinkSection = {
  id: string;
  title: string;
  links: QuickLink[];
};

const SECTIONS: QuickLinkSection[] = [
  {
    id: 'sales',
    title: 'Sales',
    links: [
      { to: '/', label: 'Terminal', icon: Monitor },
      { to: '/orders', label: 'Orders', icon: ReceiptText },
      { to: '/held-orders', label: 'Held', icon: CirclePause },
      { to: '/tables', label: 'Tables', icon: Armchair, feature: 'tables' },
      { to: '/kds', label: 'KDS', icon: ChefHat, feature: 'kitchenDisplay' },
    ],
  },
  {
    id: 'catalog',
    title: 'Catalog',
    links: [
      {
        to: '/menu',
        label: (profile) => (profile === 'retail' ? 'Products' : 'Menu'),
        icon: ClipboardList,
      },
      { to: '/inventory', label: 'Inventory', icon: Boxes, feature: 'inventory' },
    ],
  },
  {
    id: 'guests',
    title: 'Guests',
    links: [
      { to: '/customer-lookup', label: 'Customers', icon: User, feature: 'customers' },
      { to: '/gift-cards', label: 'Gift Cards', icon: Gift, feature: 'giftCards' },
    ],
  },
  {
    id: 'shift',
    title: 'Shift',
    links: [
      { to: '/receipts', label: 'Receipts', icon: Calculator },
      { to: '/day-summary', label: 'Day', icon: ChartNoAxesCombined },
      { to: '/offline', label: 'Offline', icon: WifiOff },
      { to: '/settings', label: 'Settings', icon: Settings },
    ],
  },
];

const QuickAccessSidebar: React.FC = () => {
  const enabled = useSettingsStore((s) => s.quickAccessSidebar);
  const features = useSettingsStore((s) => s.features);
  const businessProfile = useSettingsStore((s) => s.businessProfile);
  if (!enabled) return null;

  const visibleSections = SECTIONS.map((section) => ({
    ...section,
    links: section.links.filter((l) => !l.feature || features[l.feature]),
  })).filter((section) => section.links.length > 0);

  return (
    <aside
      aria-label="POS quick navigation"
      className="hidden md:flex shrink-0 w-[4.75rem] flex-col border-r border-pos-border bg-pos-card overflow-y-auto py-1"
    >
      {visibleSections.map((section, sectionIndex) => (
        <div
          key={section.id}
          role="group"
          aria-label={section.title}
          className={sectionIndex > 0 ? 'mt-1 border-t border-pos-border/70 pt-1.5' : ''}
        >
          <span className="block px-1 pb-1 text-center text-[9px] font-semibold uppercase tracking-wider text-pos-muted/70">
            {section.title}
          </span>
          {section.links.map((l) => {
            const Icon = l.icon;
            const label = typeof l.label === 'function' ? l.label(businessProfile) : l.label;
            return (
              <NavLink
                key={l.to}
                to={l.to}
                end={l.to === '/'}
                className={({ isActive }) =>
                  `flex flex-col items-center justify-center gap-0.5 py-2.5 text-[10px] font-medium transition-colors ${
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
        </div>
      ))}
    </aside>
  );
};

export default QuickAccessSidebar;
