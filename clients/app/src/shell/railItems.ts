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

/** The app bar icons this person can use, in default order. */
export function railItems(user: StoredUser | null, t: TFunction): RailItem[] {
  const insights = insightsNavItems.find((item) => canAccessPortalPath(user, item.path));
  const apps = user?.effectiveAccess?.apps
    .filter((app) => app.licensed && app.hasAccess)
    .map((app) => app.appId);
  const canAutomations = canAccessPortalPath(user, '/automations');
  return [
    { id: 'home', label: t('common.portal.home'), icon: HomeIcon, to: '/' },
    ...(insights
      ? [{ id: 'insights', label: t('appNames.short.insights'), icon: InsightsIcon, to: insights.path }]
      : []),
    ...(
      [
        ['flow', '/flow', FlowIcon],
        ['crm', '/crm', CrmIcon],
        ['erp', '/erp', ErpIcon],
        ['finance', '/finance', FinanceIcon],
        ['hr', '/hr', HrIcon],
      ] as const
    )
      .filter(([, path]) => canAccessPortalPath(user, path))
      .map(([id, path, icon]) => ({ id, label: t(`appNames.short.${id}`), icon, to: path })),
    ...(apps?.includes('pos')
      ? [{ id: 'pos', label: t('appNames.short.pos'), icon: PosIcon, href: resolveAppUrl('pos') }]
      : []),
    ...(canAutomations || canAccessPortalPath(user, '/automations/webhooks')
      ? [
          {
            id: 'automations',
            label: t('common.nav.automations'),
            icon: AutomationsIcon,
            to: canAutomations ? '/automations' : '/automations/webhooks',
          },
        ]
      : []),
    { id: 'caret', label: 'Caret', icon: CaretIcon, to: '/caret' },
  ];
}
