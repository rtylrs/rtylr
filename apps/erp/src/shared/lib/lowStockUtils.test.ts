import { describe, it, expect } from 'vitest';
import {
  urgencyScore,
  alertSeverity,
  alertMessage,
  buildAlert,
  buildAlerts,
  criticalAlerts,
  actionableAlerts,
  alertSummary,
} from './lowStockUtils';
import type { LowStockItem } from '../services/erpApi';

const makeItem = (stock: number, minStock: number, name = 'Widget'): LowStockItem => ({
  name,
  sku: 'WGT-001',
  stock,
  minStock,
  unit: 'pcs',
});

describe('urgencyScore', () => {
  it('returns 100 for out-of-stock', () => {
    expect(urgencyScore(makeItem(0, 10))).toBe(100);
  });

  it('returns 0 for fully stocked', () => {
    expect(urgencyScore(makeItem(10, 10))).toBe(0);
  });

  it('returns proportional score for partial stock', () => {
    expect(urgencyScore(makeItem(5, 10))).toBe(50);
  });

  it('returns 0 when minStock is 0', () => {
    expect(urgencyScore(makeItem(0, 0))).toBe(0);
  });
});

describe('alertSeverity', () => {
  it('returns critical for score >= 80', () => {
    expect(alertSeverity(makeItem(0, 10))).toBe('critical');
    expect(alertSeverity(makeItem(1, 10))).toBe('critical');
  });

  it('returns warning for score 40-79', () => {
    expect(alertSeverity(makeItem(3, 10))).toBe('warning');
    expect(alertSeverity(makeItem(6, 10))).toBe('warning');
  });

  it('returns info for score below 40', () => {
    expect(alertSeverity(makeItem(8, 10))).toBe('info');
  });
});

describe('alertMessage', () => {
  it('mentions out of stock when stock is 0', () => {
    const msg = alertMessage(makeItem(0, 10));
    expect(msg).toContain('out of stock');
  });

  it('includes current and minimum stock', () => {
    const msg = alertMessage(makeItem(3, 10));
    expect(msg).toContain('3');
    expect(msg).toContain('10');
  });
});

describe('buildAlert', () => {
  it('returns a complete StockAlert object', () => {
    const alert = buildAlert(makeItem(0, 10, 'Milk'));
    expect(alert.severity).toBe('critical');
    expect(alert.urgencyScore).toBe(100);
    expect(alert.message).toContain('Milk');
  });
});

describe('buildAlerts', () => {
  it('sorts alerts by urgency score descending', () => {
    const items = [makeItem(8, 10, 'A'), makeItem(0, 10, 'B'), makeItem(3, 10, 'C')];
    const alerts = buildAlerts(items);
    expect(alerts[0].item.name).toBe('B');
    expect(alerts[2].item.name).toBe('A');
  });
});

describe('criticalAlerts', () => {
  it('returns only critical severity alerts', () => {
    const items = [makeItem(0, 10, 'A'), makeItem(8, 10, 'B')];
    const critical = criticalAlerts(items);
    expect(critical).toHaveLength(1);
    expect(critical[0].item.name).toBe('A');
  });
});

describe('actionableAlerts', () => {
  it('returns warning and critical alerts', () => {
    const items = [makeItem(0, 10, 'A'), makeItem(3, 10, 'B'), makeItem(9, 10, 'C')];
    const actionable = actionableAlerts(items);
    expect(actionable.length).toBeGreaterThanOrEqual(2);
    expect(actionable.find((a) => a.item.name === 'C')).toBeUndefined();
  });
});

describe('alertSummary', () => {
  it('counts alerts by severity', () => {
    const items = [makeItem(0, 10, 'A'), makeItem(3, 10, 'B'), makeItem(9, 10, 'C')];
    const summary = alertSummary(items);
    expect(summary.critical).toBe(1);
    expect(typeof summary.warning).toBe('number');
    expect(typeof summary.info).toBe('number');
    expect(summary.critical + summary.warning + summary.info).toBe(3);
  });
});
