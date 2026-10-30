import type { ComponentType } from 'react';
import type { TFunction } from 'i18next';
import type { AnyRouter } from '@tanstack/react-router';
import { Settings } from 'lucide-react';
import {
  AutomationsIcon,
  CrmIcon,
  ErpIcon,
  FinanceIcon,
  FlowIcon,
  HomeIcon,
  HrIcon,
  InsightsIcon,
  RecruitIcon,
} from '@voxire/ui';
import { CaretIcon } from '@voxire/caret';
import { rawMenuSubItems, rawNavItems, settingsNavItems } from '../../shell/navigation';
import { crmSidebarItems } from '../../modules/crm/navigation';
import { erpNavItems } from '../../modules/erp/navigation';
import { financeNavItems } from '../../modules/finance/navigation';
import { hrNavItems } from '../../modules/hr/navigation';
import { recruitNavItems } from '../../modules/recruit/navigation';
import { insightsNavItems } from '../../modules/insights/navigation';
import { flowNavItems } from '../../modules/flow/navigation';

export type PortalApp =
  | 'workspace'
  | 'insights'
  | 'flow'
  | 'crm'
  | 'erp'
  | 'finance'
  | 'hr'
  | 'recruit'
  | 'automations'
  | 'settings'
  | 'caret';

interface NavEntry {
  key: string;
  path: string;
}

const AUTOMATION_ITEMS: NavEntry[] = [
  { key: 'automationWorkflows', path: '/automations' },
  { key: 'newAutomation', path: '/automations/new' },
  { key: 'webhooks', path: '/automations/webhooks' },
];

// Each app's sidebar and the i18n prefix of its labels, as PortalShell shows them.
const APP_NAV: Record<PortalApp, { items: NavEntry[]; label: (key: string) => string }> = {
  workspace: {
    items: [...rawNavItems, ...rawMenuSubItems].map((item) => ({
      key: item.key ?? item.path,
      path: item.path,
    })),
    label: (key) => `common.nav.${key}`,
  },
  insights: { items: entries(insightsNavItems), label: (key) => `insightsNav.${key}` },
  flow: { items: entries(flowNavItems), label: (key) => key },
  crm: { items: entries(crmSidebarItems), label: (key) => `crmNav.${key}` },
  erp: { items: entries(erpNavItems), label: (key) => `erpNav.${key}` },
  finance: { items: entries(financeNavItems), label: (key) => `financeNav.${key}` },
  hr: { items: entries(hrNavItems), label: (key) => `hrNav.${key}` },
  recruit: { items: entries(recruitNavItems), label: (key) => `recruitAdmin.nav.${key}` },
  automations: { items: AUTOMATION_ITEMS, label: (key) => `common.nav.${key}` },
  settings: { items: entries(settingsNavItems), label: (key) => `common.nav.${key}` },
  caret: { items: [], label: (key) => key },
};

const APP_ICONS: Record<PortalApp, ComponentType<{ size?: number }>> = {
  workspace: HomeIcon,
  insights: InsightsIcon,
  flow: FlowIcon,
  crm: CrmIcon,
  erp: ErpIcon,
  finance: FinanceIcon,
  hr: HrIcon,
  recruit: RecruitIcon,
  automations: AutomationsIcon,
  settings: Settings,
  caret: CaretIcon,
};

function entries(items: ReadonlyArray<{ key?: string; path: string }>): NavEntry[] {
  return items.map((item) => ({ key: item.key ?? item.path, path: item.path }));
}

export function appForPath(path: string): PortalApp {
  const first = path.split(/[?#]/)[0].split('/')[1] ?? '';
  if (first === 'settings') return 'settings';
  if (
    ['insights', 'flow', 'crm', 'erp', 'finance', 'hr', 'recruit', 'automations', 'caret'].includes(
      first
    )
  )
    return first as PortalApp;
  return 'workspace';
}

/** Where a path sits in the Portal: its app, that app's sidebar with the
 *  page marked, and the page's name. */
export interface PortalPlace {
  app: PortalApp;
  appLabel: string;
  icon: ComponentType<{ size?: number }>;
  items: { key: string; label: string; active: boolean }[];
  title?: string;
}

export function describePlace(path: string, t: TFunction): PortalPlace {
  const pathname = path.split(/[?#]/)[0].replace(/\/+$/, '') || '/';
  const app = appForPath(pathname);
  const nav = APP_NAV[app];
  const match = nav.items
    .filter(
      (item) =>
        pathname === item.path || (item.path !== '/' && pathname.startsWith(`${item.path}/`))
    )
    .sort((a, b) => b.path.length - a.path.length)[0];
  const items = nav.items.map((item) => ({
    key: item.key,
    label: t(nav.label(item.key)),
    active: item === match,
  }));
  const appLabel =
    app === 'workspace'
      ? t('common.portal.workspace')
      : app === 'settings'
        ? t('common.nav.settings')
        : app === 'automations'
          ? t('common.nav.automations')
          : app === 'caret'
            ? 'Caret'
            : t(`appNames.short.${app}`);
  return {
    app,
    appLabel,
    icon: APP_ICONS[app],
    items,
    title: match ? t(nav.label(match.key)) : undefined,
  };
}

/** Whether the router has a page at this path, rather than the not-found page. */
export function isPortalRoute(router: AnyRouter, path: string): boolean {
  const pathname = path.split(/[?#]/)[0];
  const { foundRoute } = router.getMatchedRoutes(pathname);
  return Boolean(foundRoute && foundRoute.path !== '*' && foundRoute.path !== '$');
}
