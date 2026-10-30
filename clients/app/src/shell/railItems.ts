import type { ComponentType } from 'react';
import type { TFunction } from 'i18next';
import { resolveAppUrl } from '@voxire/config';
import {
  AutomationsIcon,
  CrmIcon,
  ErpIcon,
  FinanceIcon,
  FlowIcon,
  HomeIcon,
  HrIcon,
  InsightsIcon,
  PosIcon,
} from '@voxire/ui';
import { CaretIcon } from '@voxire/caret';
import type { StoredUser } from '../shared/stores/userStore';
import { canAccessPortalPath } from '../shared/utils/canAccessPortalPath';
import { insightsNavItems } from '../modules/insights/navigation';

/** One icon of the app bar. Ids match the server's navigation map. */
export interface RailItem {
  id: string;
  label: string;
  icon: ComponentType<{ size?: number }>;
  /** A route of this Portal; `href` is another app. */
  to?: string;
  href?: string;
}

// Caret sits below the apps until the person moves it.
export const RAIL_BOTTOM = ['caret'];

export const RAIL_ICONS: Record<string, ComponentType<{ size?: number }>> = {
  home: HomeIcon,
  insights: InsightsIcon,
  flow: FlowIcon,
  crm: CrmIcon,
  erp: ErpIcon,
  finance: FinanceIcon,
  hr: HrIcon,
  pos: PosIcon,
  automations: AutomationsIcon,
  caret: CaretIcon,
};

/** The app bar icons this person can use, in default order. */
export function railItems(user: StoredUser | null, t: TFunction): RailItem[] {
  const insights = insightsNavItems.find((item) => canAccessPortalPath(user, item.path));
  const apps = user?.effectiveAccess?.apps
    .filter((app) => app.licensed && app.hasAccess)
    .map((app) => app.appId);
  const canAutomations = canAccessPortalPath(user, '/automations');
  return [
    { id: 'home', label: t('common.portal.home'), icon: RAIL_ICONS.home, to: '/' },
    ...(insights
      ? [
          {
            id: 'insights',
            label: t('appNames.short.insights'),
            icon: RAIL_ICONS.insights,
            to: insights.path,
          },
        ]
      : []),
    ...(
      [
        ['flow', '/flow'],
        ['crm', '/crm'],
        ['erp', '/erp'],
        ['finance', '/finance'],
        ['hr', '/hr'],
      ] as const
    )
      .filter(([, path]) => canAccessPortalPath(user, path))
      .map(([id, path]) => ({
        id,
        label: t(`appNames.short.${id}`),
        icon: RAIL_ICONS[id],
        to: path,
      })),
    ...(apps?.includes('pos')
      ? [
          {
            id: 'pos',
            label: t('appNames.short.pos'),
            icon: RAIL_ICONS.pos,
            href: resolveAppUrl('pos'),
          },
        ]
      : []),
    ...(canAutomations || canAccessPortalPath(user, '/automations/webhooks')
      ? [
          {
            id: 'automations',
            label: t('common.nav.automations'),
            icon: RAIL_ICONS.automations,
            to: canAutomations ? '/automations' : '/automations/webhooks',
          },
        ]
      : []),
    { id: 'caret', label: 'Caret', icon: RAIL_ICONS.caret, to: '/caret' },
  ];
}
