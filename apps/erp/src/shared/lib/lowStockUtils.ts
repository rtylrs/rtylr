/**
 * Low-stock alert helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

import type { LowStockItem } from '../services/erpApi';

export type AlertSeverity = 'critical' | 'warning' | 'info';

export interface StockAlert {
  item: LowStockItem;
  severity: AlertSeverity;
  urgencyScore: number;
  message: string;
}

/**
 * Compute a stock urgency score (0-100).
 * 0 = completely stocked, 100 = completely out.
 */
export function urgencyScore(item: LowStockItem): number {
  if (item.minStock === 0) return 0;
  const ratio = item.stock / item.minStock;
  if (ratio <= 0) return 100;
  if (ratio >= 1) return 0;
  return Math.round((1 - ratio) * 100);
}

/** Classify an item into an alert severity. */
export function alertSeverity(item: LowStockItem): AlertSeverity {
  const score = urgencyScore(item);
  if (score >= 80) return 'critical';
  if (score >= 40) return 'warning';
  return 'info';
}

/** Build a human-readable alert message for a low-stock item. */
export function alertMessage(item: LowStockItem): string {
  if (item.stock === 0)
    return `${item.name} is completely out of stock (min: ${item.minStock} ${item.unit}).`;
  return `${item.name} has only ${item.stock} ${item.unit} left (min: ${item.minStock} ${item.unit}).`;
}

/** Build a full StockAlert from a LowStockItem. */
export function buildAlert(item: LowStockItem): StockAlert {
  return {
    item,
    severity: alertSeverity(item),
    urgencyScore: urgencyScore(item),
    message: alertMessage(item),
  };
}

/** Build and sort alerts by urgency score (highest first). */
export function buildAlerts(items: LowStockItem[]): StockAlert[] {
  return items.map(buildAlert).sort((a, b) => b.urgencyScore - a.urgencyScore);
}

/** Return only critical alerts. */
export function criticalAlerts(items: LowStockItem[]): StockAlert[] {
  return buildAlerts(items).filter((a) => a.severity === 'critical');
}

/** Return only warning or critical alerts. */
export function actionableAlerts(items: LowStockItem[]): StockAlert[] {
  return buildAlerts(items).filter((a) => a.severity !== 'info');
}

/** Return the count of items at each severity level. */
export function alertSummary(items: LowStockItem[]): Record<AlertSeverity, number> {
  const alerts = buildAlerts(items);
  return {
    critical: alerts.filter((a) => a.severity === 'critical').length,
    warning: alerts.filter((a) => a.severity === 'warning').length,
    info: alerts.filter((a) => a.severity === 'info').length,
  };
}
