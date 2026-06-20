import type { CaretPageContext, CaretPageContextValue } from "../types/context";

export interface BuildCaretRouteContextInput {
  appId: string;
  appName?: string;
  pathname?: string;
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
  return typeof value === "string" && value.trim() ? value.trim() : undefined;
}

function normalizePath(pathname: string | undefined): string {
  const value = clean(pathname) ?? "/";
  const pathOnly = value.split("?")[0]?.split("#")[0] ?? "/";
  return pathOnly.startsWith("/") ? pathOnly : `/${pathOnly}`;
}

function bestRouteMatch<T>(
  pathname: string,
  entries: Record<string, T> | undefined,
): T | undefined {
  if (!entries) return undefined;
  if (entries[pathname]) return entries[pathname];

  const sorted = Object.keys(entries)
    .filter(
      (path) =>
        path !== "/" && (pathname === path || pathname.startsWith(`${path}/`)),
    )
    .sort((a, b) => b.length - a.length);

  return sorted.length > 0 ? entries[sorted[0]] : entries["/"];
}

function firstRouteSegment(pathname: string): string | undefined {
  return clean(pathname.split("/").filter(Boolean)[0]);
}

function humanize(value: string | undefined): string | undefined {
  const text = clean(value);
  if (!text) return undefined;
  return text
    .replace(/[_-]+/g, " ")
    .replace(/\s+/g, " ")
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

export function buildCaretRouteContext(
  input: BuildCaretRouteContextInput,
): CaretPageContext {
  const pathname = normalizePath(input.pathname);
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
    selected: input.selected,
    filters: input.filters,
    visibleRows: input.visibleRows,
    businessMemory: input.businessMemory,
    cacheContext: input.cacheContext,
    dataFreshness: input.dataFreshness,
    hints: uniqueHints(input.hints ?? []),
  };
}
