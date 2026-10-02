import { buildCaretRouteContext, type CaretPageContext } from '@voxire/caret';
import { todayIso } from './dateUtils';

export interface PortalSystemContextInput {
  pathname: string;
  search?: string;
  businessName?: string;
  apps?: string[];
}

interface PageInfo {
  module: string;
  page: string;
  hint: string;
}

const PAGE_HINTS: Record<string, PageInfo> = {
  '/': {
    module: 'dashboard',
    page: 'Portal Home',
    hint: 'account shortcuts and available business apps',
  },
  '/flow': {
    module: 'flow',
    page: 'Flow',
    hint: 'projects, tasks, time tracking and team planning',
  },
  '/insights': {
    module: 'insights',
    page: 'Business Overview',
    hint: 'cross-app business analytics',
  },
  '/insights/revenue': {
    module: 'insights',
    page: 'Revenue',
    hint: 'revenue across business apps',
  },
  '/insights/customers': { module: 'insights', page: 'Customers', hint: 'customer analytics' },
  '/insights/products': { module: 'insights', page: 'Products', hint: 'product analytics' },
  '/insights/operations': {
    module: 'insights',
    page: 'Operations',
    hint: 'business operations analytics',
  },
  '/insights/financial': {
    module: 'insights',
    page: 'Financial',
    hint: 'business financial analytics',
  },
  '/insights/reports': { module: 'insights', page: 'Reports', hint: 'business reporting' },
  '/insights/settings': {
    module: 'insights',
    page: 'Insights settings',
    hint: 'analytics and scheduled report settings',
  },
  '/team': {
    module: 'team',
    page: 'Team management',
    hint: 'members, invites, roles, and permissions',
  },
  '/menus': {
    module: 'storefront',
    page: 'Menus',
    hint: 'menu items and storefront catalog setup',
  },
  '/orders': {
    module: 'storefront',
    page: 'Orders',
    hint: 'storefront order review',
  },
  '/qr-codes': {
    module: 'storefront',
    page: 'QR Codes',
    hint: 'QR code generation',
  },
  '/design': {
    module: 'storefront',
    page: 'Design',
    hint: 'storefront design and branding',
  },
  '/settings': {
    module: 'settings',
    page: 'Settings',
    hint: 'business settings',
  },
  '/caret': {
    module: 'caret',
    page: 'Caret chat',
    hint: 'full-page AI assistant',
  },
};

const ROUTE_LABELS = Object.fromEntries(
  Object.entries(PAGE_HINTS).map(([path, info]) => [path, info.page])
);

const ROUTE_MODULES = Object.fromEntries(
  Object.entries(PAGE_HINTS).map(([path, info]) => [path, info.module])
);

function pageInfo(pathname: string): PageInfo {
  if (PAGE_HINTS[pathname]) return PAGE_HINTS[pathname];
  const match = Object.keys(PAGE_HINTS)
    .filter((path) => path !== '/' && pathname.startsWith(`${path}/`))
    .sort((a, b) => b.length - a.length)[0];
  return match
    ? PAGE_HINTS[match]
    : {
        module: 'home',
        page: 'Portal',
        hint: 'navigating the business workspace',
      };
}

function selectedFromPath(pathname: string): Record<string, string> | undefined {
  const selected: Record<string, string> = {};
  const routes: Array<[RegExp, string]> = [
    [/^\/flow\/p\/([^/]+)/, 'project_id'],
    [/^\/flow\/task\/([^/]+)/, 'task_id'],
    [/^\/orders\/([^/]+)/, 'order_id'],
    [/^\/menus\/([^/]+)/, 'menu_id'],
    [/^\/team\/groups\/(\d+)(?:\/|$)/, 'group_id'],
    [/^\/team\/(\d+)(?:\/|$)/, 'employee_id'],
    [/^\/customers\/([^/]+)/, 'customer_id'],
    [/^\/products\/([^/]+)/, 'product_id'],
    [/^\/items\/([^/]+)/, 'product_id'],
    [/^\/settings\/locations\/([^/]+)/, 'location_id'],
  ];

  for (const [pattern, key] of routes) {
    const match = pathname.match(pattern);
    if (match?.[1]) selected[key] = decodeURIComponent(match[1]);
  }

  return Object.keys(selected).length > 0 ? selected : undefined;
}

function browserTimezone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'Asia/Beirut';
  } catch {
    return 'Asia/Beirut';
  }
}

function browserLocale(): string {
  try {
    return navigator.language || 'en-US';
  } catch {
    return 'en-US';
  }
}

function cacheDomainsForModule(module: string): string[] {
  switch (module) {
    case 'storefront':
      return ['storefront', 'catalog', 'orders'];
    case 'team':
      return ['hr', 'team'];
    case 'settings':
      return ['settings', 'billing', 'extensions'];
    case 'insights':
      return ['analytics', 'crm', 'erp', 'finance', 'hr', 'pos'];
    case 'caret':
      return ['analytics', 'crm', 'erp', 'finance', 'hr', 'pos', 'storefront'];
    case 'dashboard':
      return ['analytics', 'orders', 'finance', 'inventory', 'crm'];
    default:
      return [module || 'dash'];
  }
}

function buildBusinessMemory(): NonNullable<CaretPageContext['businessMemory']> {
  return {
    timezone: browserTimezone(),
    locale: browserLocale(),
    currency: 'USD',
    answerStyle: 'lead with the answer, then show the useful numbers',
    tablePreference: 'compact tables for comparisons',
  };
}

function buildCacheContext(
  info: PageInfo,
  pathname: string,
  selected: Record<string, string> | undefined
): NonNullable<CaretPageContext['cacheContext']> {
  return {
    domains: cacheDomainsForModule(info.module),
    day: todayIso(),
    route: pathname,
    selected: selected ?? {},
  };
}

export function buildPortalSystemContext({
  pathname,
  search,
  businessName,
  apps,
}: PortalSystemContextInput): string {
  const ctx = buildPortalPageContext({ pathname, search, businessName, apps });
  const lines = [
    `App: ${ctx.appName}.`,
    businessName ? `Business: ${businessName}.` : '',
    `Route: ${pathname}.`,
    `Current module: ${ctx.module ?? 'dash'}.`,
    `Current page: ${ctx.page ?? 'Portal'}.`,
    ctx.hints?.length ? `Page hint: ${ctx.hints.join('; ')}.` : '',
  ].filter(Boolean);
  return lines.join('\n');
}

export function buildPortalPageContext({
  pathname,
  search,
  businessName,
  apps,
}: PortalSystemContextInput): CaretPageContext {
  const info = pageInfo(pathname);
  const selected = selectedFromPath(pathname);
  const cacheContext = buildCacheContext(info, pathname, selected);
  const context = buildCaretRouteContext({
    appId: info.module === 'flow' ? 'flow' : info.module === 'insights' ? 'insights' : 'dash',
    appName:
      info.module === 'flow'
        ? 'RTYLR Flow'
        : info.module === 'insights'
          ? 'RTYLR Insights'
          : 'RTYLR Portal',
    pathname,
    search,
    routeLabels: ROUTE_LABELS,
    routeModules: ROUTE_MODULES,
    businessName,
    selected,
    businessMemory: buildBusinessMemory(),
    cacheContext,
    dataFreshness: cacheContext,
    hints: [info.hint],
  });

  return {
    ...context,
    enabledApps: apps?.slice(0, 20),
  };
}
