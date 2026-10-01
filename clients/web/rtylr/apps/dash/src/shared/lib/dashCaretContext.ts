import { buildCaretRouteContext, type CaretPageContext } from '@voxire/caret';
import { todayIso } from './dateUtils';

export interface DashSystemContextInput {
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
    page: 'Dashboard home',
    hint: 'KPIs and quick overview',
  },
  '/team': {
    module: 'team',
    page: 'Team management',
    hint: 'members, invites, roles, and permissions',
  },
  '/apps': {
    module: 'apps',
    page: 'Apps marketplace',
    hint: 'enabled apps and available add-ons',
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
  '/chat': {
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
  const match = Object.keys(PAGE_HINTS).find((p) => pathname.startsWith(p) && p !== '/');
  return match
    ? PAGE_HINTS[match]
    : {
        module: 'dash',
        page: 'Unknown Dash page',
        hint: 'navigating the dash app',
      };
}

function selectedFromPath(pathname: string): Record<string, string> | undefined {
  const selected: Record<string, string> = {};
  const routes: Array<[RegExp, string]> = [
    [/^\/orders\/([^/]+)/, 'order_id'],
    [/^\/menus\/([^/]+)/, 'menu_id'],
    [/^\/team\/([^/]+)/, 'employee_id'],
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
      return ['settings', 'billing', 'addons'];
    case 'apps':
      return ['apps', 'billing'];
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

export function buildDashSystemContext({
  pathname,
  search,
  businessName,
  apps,
}: DashSystemContextInput): string {
  const ctx = buildDashPageContext({ pathname, search, businessName, apps });
  const lines = [
    'App: Dash business command center.',
    businessName ? `Business: ${businessName}.` : '',
    `Route: ${pathname}.`,
    `Current module: ${ctx.module ?? 'dash'}.`,
    `Current page: ${ctx.page ?? 'Dash'}.`,
    ctx.hints?.length ? `Page hint: ${ctx.hints.join('; ')}.` : '',
  ].filter(Boolean);
  return lines.join('\n');
}

export function buildDashPageContext({
  pathname,
  search,
  businessName,
  apps,
}: DashSystemContextInput): CaretPageContext {
  const info = pageInfo(pathname);
  const selected = selectedFromPath(pathname);
  const cacheContext = buildCacheContext(info, pathname, selected);
  const context = buildCaretRouteContext({
    appId: 'dash',
    appName: 'Dash business command center',
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
