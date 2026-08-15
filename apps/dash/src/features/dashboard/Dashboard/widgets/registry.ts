import { getTenantKey } from '../../../../shared/utils/tenantStorage';

export interface WidgetConfig {
  id: string;
  label: string;
  description: string;
  defaultW: number;
  defaultH: number;
  minW?: number;
  minH?: number;
  maxW?: number;
  maxH?: number;
}

export const WIDGET_CATALOG: WidgetConfig[] = [
  {
    id: 'kpi-revenue',
    label: 'Revenue',
    description: "Today's revenue vs yesterday",
    defaultW: 1,
    defaultH: 2,
    minW: 1,
    minH: 2,
    maxH: 2,
  },
  {
    id: 'kpi-orders',
    label: 'Orders',
    description: "Today's order count",
    defaultW: 1,
    defaultH: 2,
    minW: 1,
    minH: 2,
    maxH: 2,
  },
  {
    id: 'kpi-team',
    label: 'Team',
    description: 'Staff on shift right now',
    defaultW: 1,
    defaultH: 2,
    minW: 1,
    minH: 2,
    maxH: 2,
  },
  {
    id: 'kpi-invoices',
    label: 'Open Invoices',
    description: 'Unpaid invoices',
    defaultW: 1,
    defaultH: 2,
    minW: 1,
    minH: 2,
    maxH: 2,
  },
  {
    id: 'activity',
    label: 'Recent Activity',
    description: 'Latest actions across your business',
    defaultW: 2,
    defaultH: 4,
    minW: 1,
    minH: 3,
  },
  {
    id: 'team',
    label: 'Team Overview',
    description: 'Staff on shift and recent hires',
    defaultW: 2,
    defaultH: 3,
    minW: 1,
    minH: 2,
  },
  {
    id: 'low-stock',
    label: 'Low Stock Alerts',
    description: 'Products that need reordering',
    defaultW: 2,
    defaultH: 3,
    minW: 1,
    minH: 2,
  },
  {
    id: 'revenue-chart',
    label: 'Revenue Chart',
    description: 'Revenue trend over the last 7 days',
    defaultW: 2,
    defaultH: 3,
    minW: 2,
    minH: 3,
  },
  {
    id: 'revenue-by-app',
    label: 'Revenue by App',
    description: 'MTD revenue split across POS, Finance, ERP, CRM',
    defaultW: 2,
    defaultH: 3,
    minW: 1,
    minH: 2,
  },
  {
    id: 'invoice-aging',
    label: 'Invoice Aging',
    description: 'A/R buckets - 0-30, 31-60, 61-90, 90+ days',
    defaultW: 2,
    defaultH: 3,
    minW: 1,
    minH: 2,
  },
  {
    id: 'top-customers',
    label: 'Top Customers',
    description: 'Top 5 by lifetime spend',
    defaultW: 2,
    defaultH: 3,
    minW: 1,
    minH: 2,
  },
  {
    id: 'top-products',
    label: 'Top Products',
    description: 'Best selling items this week',
    defaultW: 2,
    defaultH: 3,
    minW: 1,
    minH: 2,
  },
  {
    id: 'pending',
    label: 'Pending Approvals',
    description: 'Time-off requests and PO approvals waiting',
    defaultW: 2,
    defaultH: 3,
    minW: 1,
    minH: 2,
  },
  {
    id: 'apps',
    label: 'App Shortcuts',
    description: 'Quick access to your apps',
    defaultW: 4,
    defaultH: 2,
    minW: 2,
    minH: 2,
    maxH: 2,
  },
];

// Every row fills all four columns so the default layout has no dead space:
//   row 0 — four KPI cards (1 col each)
//   row 2 — Recent activity (3 cols) + Pending approvals sidebar (1 col)
//   row 6 — Team overview (full width)
export const DEFAULT_LAYOUT = [
  { i: 'kpi-revenue', x: 0, y: 0, w: 1, h: 2 },
  { i: 'kpi-orders', x: 1, y: 0, w: 1, h: 2 },
  { i: 'kpi-team', x: 2, y: 0, w: 1, h: 2 },
  { i: 'kpi-invoices', x: 3, y: 0, w: 1, h: 2 },
  { i: 'activity', x: 0, y: 2, w: 3, h: 4 },
  { i: 'pending', x: 3, y: 2, w: 1, h: 4 },
  { i: 'team', x: 0, y: 6, w: 4, h: 3 },
];

// Saved layouts that still reference a retired widget id are reset to default so
// the user never sees an empty card for a widget that no longer exists.
//   'kpis'          — old combined KPI widget, split into four kpi-* cards
//   'quick-actions' — removed (its actions dead-linked to /apps), replaced by pending
const RETIRED_WIDGET_IDS = ['kpis', 'quick-actions'];

const STORAGE_KEY = 'rtylr-dashboard-layout';
const WIDGETS_KEY = 'rtylr-dashboard-widgets';

export function loadLayout() {
  try {
    const raw = localStorage.getItem(getTenantKey(STORAGE_KEY));
    if (!raw) return DEFAULT_LAYOUT;
    const parsed = JSON.parse(raw);
    // If the saved layout references a retired widget, reset to default.
    if (parsed.some((l: Record<string, unknown>) => RETIRED_WIDGET_IDS.includes(l.i as string)))
      return DEFAULT_LAYOUT;
    return parsed;
  } catch {
    return DEFAULT_LAYOUT;
  }
}

export function saveLayout(layout: Record<string, unknown>[]) {
  localStorage.setItem(getTenantKey(STORAGE_KEY), JSON.stringify(layout));
}

export function loadActiveWidgets(): string[] {
  try {
    const raw = localStorage.getItem(getTenantKey(WIDGETS_KEY));
    if (!raw) return DEFAULT_LAYOUT.map((l) => l.i);
    const parsed = JSON.parse(raw);
    // If the saved widget set references a retired widget, reset to default.
    if (parsed.some((id: string) => RETIRED_WIDGET_IDS.includes(id)))
      return DEFAULT_LAYOUT.map((l) => l.i);
    return parsed;
  } catch {
    return DEFAULT_LAYOUT.map((l) => l.i);
  }
}

export function saveActiveWidgets(ids: string[]) {
  localStorage.setItem(getTenantKey(WIDGETS_KEY), JSON.stringify(ids));
}
