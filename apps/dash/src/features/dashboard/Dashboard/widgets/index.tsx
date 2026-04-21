import React from 'react';
import { KpiRevenue, KpiOrders, KpiTeam, KpiInvoices } from './KpisWidget';
import ActivityWidget from './ActivityWidget';
import ChangelogWidget from './ChangelogWidget';
import QuickActionsWidget from './QuickActionsWidget';
import TeamWidget from './TeamWidget';
import LowStockWidget from './LowStockWidget';
import RevenueChartWidget from './RevenueChartWidget';
import RevenueByAppWidget from './RevenueByAppWidget';
import TopProductsWidget from './TopProductsWidget';
import PendingWidget from './PendingWidget';
import AppsWidget from './AppsWidget';

const WIDGET_COMPONENTS: Record<string, React.FC> = {
  'kpi-revenue': KpiRevenue,
  'kpi-orders': KpiOrders,
  'kpi-team': KpiTeam,
  'kpi-invoices': KpiInvoices,
  activity: ActivityWidget,
  changelog: ChangelogWidget,
  'quick-actions': QuickActionsWidget,
  team: TeamWidget,
  'low-stock': LowStockWidget,
  'revenue-chart': RevenueChartWidget,
  'revenue-by-app': RevenueByAppWidget,
  'top-products': TopProductsWidget,
  pending: PendingWidget,
  apps: AppsWidget,
};

export function getWidgetComponent(id: string): React.FC | null {
  return WIDGET_COMPONENTS[id] ?? null;
}

export {
  WIDGET_CATALOG,
  DEFAULT_LAYOUT,
  loadLayout,
  saveLayout,
  loadActiveWidgets,
  saveActiveWidgets,
} from './registry';
