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
  Utensils,
  BriefcaseBusiness,
} from 'lucide-react';
import { cn } from '../../lib/utils';
import { useTranslation } from 'react-i18next';
import { APP_NAMES } from '../../config/appNames';

const APP_META: Record<string, { labelKey: string; color: string; bg: string; icon: LucideIcon }> =
  {
    pos: {
      labelKey: APP_NAMES.pos.short,
      color: 'var(--app-pos)',
      bg: 'rgba(69,131,196,0.1)',
      icon: MonitorSpeaker,
    },
    erp: {
      labelKey: APP_NAMES.erp.short,
      color: 'var(--app-erp)',
      bg: 'rgba(16,185,129,0.1)',
      icon: Package,
    },
    hr: {
      labelKey: APP_NAMES.hr.short,
      color: 'var(--app-hr)',
      bg: 'rgba(245,158,11,0.1)',
      icon: UserCheck,
    },
    crm: {
      labelKey: APP_NAMES.crm.short,
      color: 'var(--app-crm)',
      bg: 'rgba(139,92,246,0.1)',
      icon: Users,
    },
    finance: {
      labelKey: APP_NAMES.finance.short,
      color: 'var(--app-finance)',
      bg: 'rgba(239,68,68,0.1)',
      icon: CreditCard,
    },
    flow: {
      labelKey: APP_NAMES.flow.short,
      color: 'var(--app-flow)',
      bg: 'rgba(100,197,185,0.1)',
      icon: Layers,
    },
    insights: {
      labelKey: APP_NAMES.insights.short,
      color: 'var(--app-insights)',
      bg: 'rgba(99,102,241,0.1)',
      icon: BarChart3,
    },
    menu: {
      labelKey: APP_NAMES.menu.short,
      color: 'var(--app-menu)',
      bg: 'rgba(249,115,22,0.1)',
      icon: Utensils,
    },
    recruit: {
      labelKey: APP_NAMES.recruit.short,
      color: 'var(--app-recruit)',
      bg: 'rgba(6,182,212,0.1)',
      icon: BriefcaseBusiness,
    },
    storefront: {
      labelKey: APP_NAMES.storefront.short,
      color: '#ec4899',
      bg: 'rgba(236,72,153,0.1)',
      icon: Store,
    },
    market: {
      labelKey: APP_NAMES.market.short,
      color: '#a855f7',
      bg: 'rgba(168,85,247,0.1)',
      icon: ShoppingBag,
    },
    web: {
      labelKey: APP_NAMES.web.short,
      color: '#0ea5e9',
      bg: 'rgba(14,165,233,0.1)',
      icon: Globe,
    },
    dash: {
      labelKey: APP_NAMES.dash.short,
      color: 'var(--dash-text-muted)',
      bg: 'var(--dash-hover)',
      icon: Circle,
    },
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
  const { t } = useTranslation();
  const meta = APP_META[app] ?? {
    labelKey: app,
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
      {showLabel && <span>{t(meta.labelKey)}</span>}
    </span>
  );
};

export default AppBadge;
