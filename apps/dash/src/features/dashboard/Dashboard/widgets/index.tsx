import React from 'react';
import { KpiRevenue, KpiOrders, KpiTeam, KpiInvoices } from './KpisWidget';
import ActivityWidget from './ActivityWidget';
import TeamWidget from './TeamWidget';
import LowStockWidget from './LowStockWidget';
import RevenueChartWidget from './RevenueChartWidget';
import RevenueByAppWidget from './RevenueByAppWidget';
import InvoiceAgingWidget from './InvoiceAgingWidget';
import TopCustomersWidget from './TopCustomersWidget';
import TopProductsWidget from './TopProductsWidget';
import PendingWidget from './PendingWidget';
import AppsWidget from './AppsWidget';

const WIDGET_COMPONENTS: Record<string, React.FC> = {
  'kpi-revenue': KpiRevenue,
  'kpi-orders': KpiOrders,
  'kpi-team': KpiTeam,
  'kpi-invoices': KpiInvoices,
  activity: ActivityWidget,
  team: TeamWidget,
  'low-stock': LowStockWidget,
  'revenue-chart': RevenueChartWidget,
  'revenue-by-app': RevenueByAppWidget,
  'invoice-aging': InvoiceAgingWidget,
  'top-customers': TopCustomersWidget,
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
