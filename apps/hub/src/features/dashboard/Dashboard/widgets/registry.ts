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
  { id: 'kpi-revenue',   label: 'Revenue',           description: 'Today\'s revenue vs yesterday',                  defaultW: 1, defaultH: 2, minW: 1, minH: 2, maxH: 2 },
  { id: 'kpi-orders',    label: 'Orders',             description: 'Today\'s order count',                           defaultW: 1, defaultH: 2, minW: 1, minH: 2, maxH: 2 },
  { id: 'kpi-team',      label: 'Team',               description: 'Staff on shift right now',                       defaultW: 1, defaultH: 2, minW: 1, minH: 2, maxH: 2 },
  { id: 'kpi-invoices',  label: 'Open Invoices',      description: 'Unpaid invoices',                                defaultW: 1, defaultH: 2, minW: 1, minH: 2, maxH: 2 },
  { id: 'activity',      label: 'Recent Activity',    description: 'Latest actions across your business',            defaultW: 2, defaultH: 4, minW: 1, minH: 3 },
  { id: 'changelog',     label: "What's New",         description: 'Latest platform updates and features',           defaultW: 1, defaultH: 4, minW: 1, minH: 3 },
  { id: 'quick-actions', label: 'Quick Actions',      description: 'Shortcuts to common tasks',                      defaultW: 1, defaultH: 4, minW: 1, minH: 2 },
  { id: 'team',          label: 'Team Overview',       description: 'Staff on shift and recent hires',                defaultW: 2, defaultH: 3, minW: 1, minH: 2 },
  { id: 'low-stock',     label: 'Low Stock Alerts',    description: 'Products that need reordering',                  defaultW: 2, defaultH: 3, minW: 1, minH: 2 },
  { id: 'revenue-chart', label: 'Revenue Chart',       description: 'Revenue trend over the last 7 days',            defaultW: 2, defaultH: 3, minW: 2, minH: 3 },
  { id: 'top-products',  label: 'Top Products',        description: 'Best selling items this week',                   defaultW: 2, defaultH: 3, minW: 1, minH: 2 },
  { id: 'pending',       label: 'Pending Approvals',   description: 'Time-off requests and PO approvals waiting',    defaultW: 2, defaultH: 3, minW: 1, minH: 2 },
  { id: 'apps',          label: 'App Shortcuts',        description: 'Quick access to your apps',                     defaultW: 4, defaultH: 2, minW: 2, minH: 2, maxH: 2 },
];

export const DEFAULT_LAYOUT = [
  { i: 'kpi-revenue',    x: 0, y: 0, w: 1, h: 2 },
  { i: 'kpi-orders',     x: 1, y: 0, w: 1, h: 2 },
  { i: 'kpi-team',       x: 2, y: 0, w: 1, h: 2 },
  { i: 'kpi-invoices',   x: 3, y: 0, w: 1, h: 2 },
  { i: 'activity',       x: 0, y: 2, w: 2, h: 4 },
  { i: 'changelog',      x: 2, y: 2, w: 1, h: 4 },
  { i: 'quick-actions',  x: 3, y: 2, w: 1, h: 4 },
];

const STORAGE_KEY = 'rtylr-dashboard-layout';
const WIDGETS_KEY = 'rtylr-dashboard-widgets';

export function loadLayout() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return DEFAULT_LAYOUT;
    const parsed = JSON.parse(raw);
    // If old layout has 'kpis' (the combined widget), reset to default
    if (parsed.some((l: any) => l.i === 'kpis')) return DEFAULT_LAYOUT;
    return parsed;
  } catch {
    return DEFAULT_LAYOUT;
  }
}

export function saveLayout(layout: any[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(layout));
}

export function loadActiveWidgets(): string[] {
  try {
    const raw = localStorage.getItem(WIDGETS_KEY);
    if (!raw) return DEFAULT_LAYOUT.map((l) => l.i);
    const parsed = JSON.parse(raw);
    // If old widgets has 'kpis', reset
    if (parsed.includes('kpis')) return DEFAULT_LAYOUT.map((l) => l.i);
    return parsed;
  } catch {
    return DEFAULT_LAYOUT.map((l) => l.i);
  }
}

export function saveActiveWidgets(ids: string[]) {
  localStorage.setItem(WIDGETS_KEY, JSON.stringify(ids));
}
