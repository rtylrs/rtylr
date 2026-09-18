import type { CaretPageContext, CaretPageContextValue } from '../types/context';

export interface BuildCaretRouteContextInput {
  appId: string;
  appName?: string;
  pathname?: string;
  search?: string;
  page?: string;
  module?: string;
  routeLabels?: Record<string, string>;
  routeModules?: Record<string, string>;
  businessName?: string;
  role?: string;
  selected?: Record<string, CaretPageContextValue>;
  filters?: Record<string, CaretPageContextValue>;
  visibleRows?: CaretPageContextValue[];
  businessMemory?: Record<string, CaretPageContextValue>;
  cacheContext?: Record<string, CaretPageContextValue>;
  dataFreshness?: CaretPageContextValue;
  hints?: string[];
}

function clean(value: unknown): string | undefined {
  return typeof value === 'string' && value.trim() ? value.trim() : undefined;
}

function normalizePath(pathname: string | undefined): string {
  const value = clean(pathname) ?? '/';
  const pathOnly = value.split('?')[0]?.split('#')[0] ?? '/';
  return pathOnly.startsWith('/') ? pathOnly : `/${pathOnly}`;
}

function bestRouteMatch<T>(
  pathname: string,
  entries: Record<string, T> | undefined
): T | undefined {
  if (!entries) return undefined;
  if (entries[pathname]) return entries[pathname];

  const sorted = Object.keys(entries)
    .filter((path) => path !== '/' && (pathname === path || pathname.startsWith(`${path}/`)))
    .sort((a, b) => b.length - a.length);

  return sorted.length > 0 ? entries[sorted[0]] : entries['/'];
}

function firstRouteSegment(pathname: string): string | undefined {
  return clean(pathname.split('/').filter(Boolean)[0]);
}

function pathSegments(pathname: string): string[] {
  return pathname
    .split('/')
    .filter(Boolean)
    .map((part) => {
      try {
        return decodeURIComponent(part);
      } catch {
        return part;
      }
    })
    .map((part) => part.trim())
    .filter(Boolean);
}

function selectedKeyForSegment(segment: string | undefined): string | undefined {
  switch (segment?.toLowerCase()) {
    case 'customers':
    case 'customer':
      return 'customer_id';
    case 'products':
    case 'product':
    case 'inventory':
      return 'product_id';
    case 'suppliers':
    case 'supplier':
      return 'supplier_id';
    case 'warehouses':
    case 'warehouse':
      return 'warehouse_id';
    case 'invoices':
    case 'invoice':
      return 'invoice_id';
    case 'expenses':
    case 'expense':
      return 'expense_id';
    case 'bills':
    case 'bill':
      return 'bill_id';
    case 'quotes':
    case 'quote':
      return 'quote_id';
    case 'staff':
    case 'employees':
    case 'employee':
      return 'employee_id';
    case 'payroll':
      return 'payroll_run_id';
    case 'tasks':
    case 'task':
      return 'task_id';
    case 'projects':
    case 'project':
      return 'project_id';
    case 'campaigns':
    case 'campaign':
      return 'campaign_id';
    case 'segments':
    case 'segment':
      return 'segment_id';
    case 'orders':
    case 'order':
      return 'order_id';
    default:
      return undefined;
  }
}

function looksLikeRecordID(value: string | undefined): value is string {
  if (!value) return false;
  if (/^\d+$/.test(value)) return true;
  if (/^[0-9a-f]{8}-[0-9a-f-]{27,}$/i.test(value)) return true;
  return value.length >= 6 && /[0-9]/.test(value) && /^[a-z0-9_-]+$/i.test(value);
}

function inferSelectedFromPath(
  pathname: string
): Record<string, CaretPageContextValue> | undefined {
  const segments = pathSegments(pathname);
  if (segments.length < 2) return undefined;

  const key = selectedKeyForSegment(segments[0]);
  const value = segments[1];
  if (!key || !looksLikeRecordID(value)) return undefined;

  return { [key]: value };
}

function humanize(value: string | undefined): string | undefined {
  const text = clean(value);
  if (!text) return undefined;
  return text
    .replace(/[_-]+/g, ' ')
    .replace(/\s+/g, ' ')
    .trim()
    .replace(/\b\w/g, (letter) => letter.toUpperCase());
}

function uniqueHints(values: Array<string | undefined>): string[] | undefined {
  const seen = new Set<string>();
  const out: string[] = [];

  for (const value of values) {
    const hint = clean(value);
    if (!hint) continue;
    const key = hint.toLowerCase();
    if (seen.has(key)) continue;
    seen.add(key);
    out.push(hint);
  }

  return out.length > 0 ? out.slice(0, 5) : undefined;
}

const hiddenSearchKeys = new Set([
  'access_token',
  'code',
  'id_token',
  'password',
  'refresh_token',
  'secret',
  'state',
  'token',
]);

function cleanSearchValue(value: string): string | undefined {
  const normalized = value.trim();
  if (!normalized) return undefined;
  return normalized.length > 120 ? `${normalized.slice(0, 117)}...` : normalized;
}

function parseSearchFilters(
  search: string | undefined
): Record<string, CaretPageContextValue> | undefined {
  const raw = clean(search);
  if (!raw) return undefined;

  const params = new URLSearchParams(raw.startsWith('?') ? raw.slice(1) : raw);
  const out: Record<string, CaretPageContextValue> = {};

  params.forEach((value, key) => {
    const normalizedKey = key.trim();
    if (!normalizedKey || hiddenSearchKeys.has(normalizedKey.toLowerCase())) return;
    const cleanValue = cleanSearchValue(value);
    if (!cleanValue) return;

    const existing = out[normalizedKey];
    if (Array.isArray(existing)) {
      out[normalizedKey] = [...existing, cleanValue];
    } else if (existing !== undefined) {
      out[normalizedKey] = [existing, cleanValue];
    } else {
      out[normalizedKey] = cleanValue;
    }
  });

  return Object.keys(out).length > 0 ? out : undefined;
}

function mergeRecords(
  first: Record<string, CaretPageContextValue> | undefined,
  second: Record<string, CaretPageContextValue> | undefined
): Record<string, CaretPageContextValue> | undefined {
  if (!first && !second) return undefined;
  return { ...(first ?? {}), ...(second ?? {}) };
}

export function buildCaretRouteContext(input: BuildCaretRouteContextInput): CaretPageContext {
  const pathname = normalizePath(input.pathname);
  const selected = mergeRecords(inferSelectedFromPath(pathname), input.selected);
  const filters = mergeRecords(parseSearchFilters(input.search), input.filters);
  const page =
    clean(input.page) ??
    clean(bestRouteMatch(pathname, input.routeLabels)) ??
    humanize(firstRouteSegment(pathname));
  const module =
    clean(input.module) ??
    clean(bestRouteMatch(pathname, input.routeModules)) ??
    firstRouteSegment(pathname) ??
    input.appId;

  return {
    app: input.appId,
    appName: clean(input.appName),
    route: pathname,
    page,
    module,
    businessName: clean(input.businessName),
    role: clean(input.role),
    selected,
    filters,
    visibleRows: input.visibleRows,
    businessMemory: input.businessMemory,
    cacheContext: input.cacheContext,
    dataFreshness: input.dataFreshness,
    hints: uniqueHints(input.hints ?? []),
  };
}
