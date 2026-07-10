import { describe, expect, it } from 'vitest';
import { billingSummarySchema, invoiceSchema } from './billing';
import { dashKpisSchema } from './kpis';
import { dashOrderSchema } from './orders';

describe('API wire format normalization', () => {
  it('maps snake_case KPI fields and numeric strings', () => {
    const result = dashKpisSchema.parse({
      revenue: '1250.50',
      order_count: '12',
      staff_count: '4',
      customer_count: '9',
      low_stock_count: '2',
      active_apps: '3',
      revenue_series: ['10', 20, 'invalid'],
    });

    expect(result).toMatchObject({
      revenue: 1250.5,
      orderCount: 12,
      staffCount: 4,
      customerCount: 9,
      lowStockCount: 2,
      activeApps: 3,
      revenueSeries: [10, 20, 0],
    });
  });

  it('maps snake_case order fields without losing money or customer data', () => {
    const result = dashOrderSchema.parse({
      id: '42',
      owner_id: '5',
      customer_name: 'Suplab Store',
      customer_email: 'suplab@example.com',
      shipping_cost: '3.50',
      subtotal: '30',
      tax: '2.25',
      total: '35.75',
      created_at: '2026-07-10T08:00:00Z',
    });

    expect(result).toMatchObject({
      id: '42',
      ownerId: '5',
      customerName: 'Suplab Store',
      customerEmail: 'suplab@example.com',
      shippingCost: 3.5,
      subtotal: 30,
      tax: 2.25,
      total: 35.75,
      createdAt: '2026-07-10T08:00:00Z',
    });
  });

  it('maps billing snake_case fields and prevents invalid numbers', () => {
    expect(
      billingSummarySchema.parse({
        plan: 'pro',
        next_billing_date: '2026-08-01',
        current_usage: '6',
      })
    ).toMatchObject({
      plan: 'pro',
      nextBillingDate: '2026-08-01',
      currentUsage: 6,
    });

    expect(
      invoiceSchema.parse({
        id: 'inv-1',
        amount: 'not-a-number',
        created_at: '2026-07-10',
        pdf_url: 'https://example.com/invoice.pdf',
      })
    ).toMatchObject({
      amount: 0,
      createdAt: '2026-07-10',
      pdfUrl: 'https://example.com/invoice.pdf',
    });
  });
});
