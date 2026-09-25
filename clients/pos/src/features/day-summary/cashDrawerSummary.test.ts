import { describe, expect, it } from 'vitest';
import { buildLocalDaySummary, normalizeDaySummaryResponse } from './cashDrawerSummary';
import type { CashMovement } from '../../shared/stores/cashDrawerStore';

const movements: CashMovement[] = [
  {
    id: 'sale-2',
    type: 'sale',
    amount: 30,
    reason: 'order-2',
    at: '2026-05-06T12:00:00.000Z',
  },
  {
    id: 'refund-1',
    type: 'refund',
    amount: 10,
    reason: 'order-1',
    at: '2026-05-06T11:00:00.000Z',
  },
  {
    id: 'drop-1',
    type: 'drop',
    amount: 50,
    reason: 'safe',
    at: '2026-05-06T10:30:00.000Z',
  },
  {
    id: 'sale-1',
    type: 'sale',
    amount: 70,
    reason: 'order-1',
    at: '2026-05-06T10:00:00.000Z',
  },
];

describe('buildLocalDaySummary', () => {
  it('derives sales, refunds, drops, and average ticket from local cash movements', () => {
    expect(
      buildLocalDaySummary({
        date: '2026-05-06',
        openingFloat: 100,
        closingCash: 140,
        movements,
      })
    ).toEqual({
      date: '2026-05-06',
      grossSales: 100,
      netSales: 90,
      taxCollected: 0,
      tips: 0,
      refunds: 10,
      voids: 0,
      orderCount: 2,
      averageTicket: 50,
      cashDrops: 50,
      openingFloat: 100,
      closingCash: 140,
    });
  });

  it('normalizes a decoded day-summary response into local summary fields', () => {
    expect(
      normalizeDaySummaryResponse(
        {
          totalSales: 100.5,
          totalTax: 10.25,
          totalTips: 8,
          totalRefunds: 2,
          totalOrders: 4,
        },
        {
          date: '2026-05-06',
          openingFloat: 100,
          closingCash: 160,
        },
      ),
    ).toEqual({
      date: '2026-05-06',
      grossSales: 100.5,
      netSales: 98.5,
      taxCollected: 10.25,
      tips: 8,
      refunds: 2,
      voids: 0,
      orderCount: 4,
      averageTicket: 25.125,
      cashDrops: 0,
      openingFloat: 100,
      closingCash: 160,
    });
  });

  it('uses cash drawer fallbacks for missing remote opening and closing cash', () => {
    expect(
      normalizeDaySummaryResponse(
        { totalSales: 25, totalRefunds: 5 },
        { date: '2026-05-06', openingFloat: 50, closingCash: 75 },
      ),
    ).toMatchObject({
      date: '2026-05-06',
      grossSales: 25,
      netSales: 20,
      openingFloat: 50,
      closingCash: 75,
    });
  });
});
