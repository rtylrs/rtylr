export interface WidgetConfig {
  id: string;
  label: string;
  description: string;
  icon: string;
  defaultW: number;
  defaultH: number;
  minW?: number;
  minH?: number;
  maxW?: number;
  maxH?: number;
}

export const WIDGET_CATALOG: WidgetConfig[] = [
  { id: 'kpis',          label: 'KPI Cards',         description: 'Revenue, orders, team, and invoices at a glance',   icon: 'trending-up',   defaultW: 4, defaultH: 2, minW: 2, minH: 2, maxH: 2 },
  { id: 'activity',      label: 'Recent Activity',   description: 'Latest actions across your business',               icon: 'activity',      defaultW: 2, defaultH: 4, minW: 2, minH: 3 },
  { id: 'changelog',     label: "What's New",        description: 'Latest platform updates and features',              icon: 'sparkles',      defaultW: 2, defaultH: 4, minW: 1, minH: 3 },
  { id: 'quick-actions', label: 'Quick Actions',     description: 'Shortcuts to common tasks',                         icon: 'zap',           defaultW: 2, defaultH: 3, minW: 1, minH: 2 },
  { id: 'team',          label: 'Team Overview',      description: 'Staff on shift and recent hires',                   icon: 'users',         defaultW: 2, defaultH: 3, minW: 1, minH: 2 },
  { id: 'low-stock',     label: 'Low Stock Alerts',   description: 'Products that need reordering',                     icon: 'package',       defaultW: 2, defaultH: 3, minW: 1, minH: 2 },
  { id: 'revenue-chart', label: 'Revenue Chart',      description: 'Revenue trend over the last 7 days',               icon: 'bar-chart-3',   defaultW: 2, defaultH: 3, minW: 2, minH: 3 },
  { id: 'top-products',  label: 'Top Products',       description: 'Best selling items this week',                      icon: 'shopping-bag',  defaultW: 2, defaultH: 3, minW: 1, minH: 2 },
  { id: 'pending',       label: 'Pending Approvals',  description: 'Time-off requests and PO approvals waiting',        icon: 'clock',         defaultW: 2, defaultH: 3, minW: 1, minH: 2 },
  { id: 'apps',          label: 'App Shortcuts',       description: 'Quick access to your apps',                         icon: 'grid-3x3',     defaultW: 4, defaultH: 2, minW: 2, minH: 2, maxH: 2 },
];

export const DEFAULT_LAYOUT = [
  { i: 'kpis',          x: 0, y: 0, w: 4, h: 2 },
  { i: 'activity',      x: 0, y: 2, w: 2, h: 4 },
  { i: 'changelog',     x: 2, y: 2, w: 1, h: 4 },
  { i: 'quick-actions', x: 3, y: 2, w: 1, h: 4 },
];

const STORAGE_KEY = 'rtylr-dashboard-layout';
const WIDGETS_KEY = 'rtylr-dashboard-widgets';

export function loadLayout() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : DEFAULT_LAYOUT;
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
    return raw ? JSON.parse(raw) : DEFAULT_LAYOUT.map((l) => l.i);
  } catch {
    return DEFAULT_LAYOUT.map((l) => l.i);
  }
}

export function saveActiveWidgets(ids: string[]) {
  localStorage.setItem(WIDGETS_KEY, JSON.stringify(ids));
}
