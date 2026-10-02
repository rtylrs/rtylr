import React from 'react';
import { Store, Globe, ShoppingBag, Circle } from 'lucide-react';
import { cn } from '../../lib/utils';
import { useTranslation } from 'react-i18next';
import { APP_ICONS } from '../../config/appIcons';
import { APP_NAMES } from '../../config/appNames';

const APP_META: Record<
  string,
  {
    labelKey: string;
    icon: React.ComponentType<import('@voxire/ui').AppIconProps>;
  }
> = {
  pos: {
    labelKey: APP_NAMES.pos.short,
    icon: APP_ICONS.pos,
  },
  erp: {
    labelKey: APP_NAMES.erp.short,
    icon: APP_ICONS.erp,
  },
  hr: {
    labelKey: APP_NAMES.hr.short,
    icon: APP_ICONS.hr,
  },
  crm: {
    labelKey: APP_NAMES.crm.short,
    icon: APP_ICONS.crm,
  },
  finance: {
    labelKey: APP_NAMES.finance.short,
    icon: APP_ICONS.finance,
  },
  flow: {
    labelKey: APP_NAMES.flow.short,
    icon: APP_ICONS.flow,
  },
  insights: {
    labelKey: APP_NAMES.insights.short,
    icon: APP_ICONS.insights,
  },
  menu: {
    labelKey: APP_NAMES.menu.short,
    icon: APP_ICONS.menu,
  },
  recruit: {
    labelKey: APP_NAMES.recruit.short,
    icon: APP_ICONS.recruit,
  },
  storefront: {
    labelKey: APP_NAMES.storefront.short,
    icon: Store,
  },
  market: {
    labelKey: APP_NAMES.market.short,
    icon: ShoppingBag,
  },
  web: {
    labelKey: APP_NAMES.web.short,
    icon: Globe,
  },
  dash: {
    labelKey: APP_NAMES.dash.short,
    icon: APP_ICONS.dash,
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
      style={{ background: 'var(--dash-hover)', color: 'var(--dash-text-dim)' }}
    >
      <Icon size={iconSize} />
      {showLabel && <span>{t(meta.labelKey)}</span>}
    </span>
  );
};

export default AppBadge;
