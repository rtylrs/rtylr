import React from 'react';
import {
  MonitorSpeaker,
  Package,
  UserCheck,
  Users,
  CreditCard,
  Layers,
  Store,
  BarChart3,
  Globe,
  ShoppingBag,
  LucideIcon,
  Circle,
} from 'lucide-react';
import { cn } from '../../lib/utils';

const APP_META: Record<string, { label: string; color: string; bg: string; icon: LucideIcon }> = {
  pos: { label: 'POS', color: 'var(--app-pos)', bg: 'rgba(69,131,196,0.1)', icon: MonitorSpeaker },
  erp: { label: 'ERP', color: 'var(--app-erp)', bg: 'rgba(16,185,129,0.1)', icon: Package },
  hr: { label: 'HR', color: 'var(--app-hr)', bg: 'rgba(245,158,11,0.1)', icon: UserCheck },
  crm: { label: 'CRM', color: 'var(--app-crm)', bg: 'rgba(139,92,246,0.1)', icon: Users },
  finance: {
    label: 'Finance',
    color: 'var(--app-finance)',
    bg: 'rgba(239,68,68,0.1)',
    icon: CreditCard,
  },
  flow: { label: 'Flow', color: 'var(--app-flow)', bg: 'rgba(100,197,185,0.1)', icon: Layers },
  storefront: { label: 'Storefront', color: '#ec4899', bg: 'rgba(236,72,153,0.1)', icon: Store },
  analytics: {
    label: 'Analytics',
    color: 'var(--app-insights)',
    bg: 'rgba(99,102,241,0.1)',
    icon: BarChart3,
  },
  market: { label: 'Marketplace', color: '#a855f7', bg: 'rgba(168,85,247,0.1)', icon: ShoppingBag },
  web: { label: 'Website', color: '#0ea5e9', bg: 'rgba(14,165,233,0.1)', icon: Globe },
  dash: { label: 'Dash', color: 'var(--dash-text-muted)', bg: 'var(--dash-hover)', icon: Circle },
};

export interface AppBadgeProps {
  app: string;
  size?: 'xs' | 'sm' | 'md';
  showLabel?: boolean;
  className?: string;
}

export const AppBadge: React.FC<AppBadgeProps> = ({
  app,
  size = 'sm',
  showLabel = true,
  className,
}) => {
  const meta = APP_META[app] ?? {
    label: app,
    color: '#6b7280',
    bg: 'rgba(107,114,128,0.1)',
    icon: Circle,
  };
  const Icon = meta.icon;
  const iconSize = size === 'xs' ? 9 : size === 'sm' ? 11 : 13;
  const padding = size === 'xs' ? 'px-1.5 py-0.5' : size === 'sm' ? 'px-2 py-1' : 'px-2.5 py-1';
  const textSize = size === 'xs' ? 'text-[9px]' : size === 'sm' ? 'text-[10px]' : 'text-xs';

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-md font-semibold',
        padding,
        textSize,
        className
      )}
      style={{ background: meta.bg, color: meta.color }}
    >
      <Icon size={iconSize} />
      {showLabel && <span>{meta.label}</span>}
    </span>
  );
};

export default AppBadge;
