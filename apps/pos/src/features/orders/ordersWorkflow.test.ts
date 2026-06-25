import { describe, expect, it } from 'vitest';
import { mapRemoteOrders, normalizeOrderResponse, normalizeOrdersResponse } from './ordersWorkflow';

describe('mapRemoteOrders', () => {
  it('maps remote orders with normalized amounts and text', () => {
    expect(
      mapRemoteOrders([
        {
          id: 42.9,
          number: ' ORD-42 ',
          tableNumber: ' T5 ',
          status: 'paid',
          subtotal: '10.5',
          tax: Number.NaN,
          tip: -3,
          discount: Number.POSITIVE_INFINITY,
          total: '12.25',
          createdAt: '2026-05-06T00:00:00.000Z',
          items: [{ id: ' 1 ', name: ' Burger ', price: -5, quantity: 2.8 }],
        },
      ])
    ).toEqual([
      {
        id: '42',
        number: 'ORD-42',
        tableNumber: 'T5',
        status: 'paid',
        items: [{ id: '1', name: 'Burger', price: 0, quantity: 2 }],
        subtotal: 10.5,
        tax: 0,
        tip: 0,
        discount: 0,
        total: 12.25,
        createdAt: '2026-05-06T00:00:00.000Z',
      },
    ]);
  });

  it('drops orders without a usable id', () => {
    expect(mapRemoteOrders([null, { id: '   ', total: 10 }, { total: 20 }])).toEqual([]);
  });

  it('defaults invalid statuses to pending', () => {
    expect(mapRemoteOrders([{ id: 1, status: 'mystery', total: 10 }])[0].status).toBe('pending');
  });

  it('normalizes common order API response envelopes', () => {
    expect(
      normalizeOrdersResponse({
        data: {
          orders: [
            {
              id: ' 43 ',
              order_number: ' ORD-43 ',
              table_number: ' Patio ',
              status: 'ready',
              subtotal: '12',
              tax: '1',
              tip: '2',
              discount: '0.5',
              total: '14.5',
              created_at: '2026-05-06T10:00:00.000Z',
              items: [{ id: ' 5 ', name: ' Tea ', unit_price: '3', qty: '2' }],
            },
          ],
        },
      }),
    ).toEqual([
      {
        id: '43',
        number: 'ORD-43',
        tableNumber: 'Patio',
        status: 'ready',
        items: [{ id: '5', name: 'Tea', price: 3, quantity: 2 }],
        subtotal: 12,
        tax: 1,
        tip: 2,
        discount: 0.5,
        total: 14.5,
        createdAt: '2026-05-06T10:00:00.000Z',
      },
    ]);
  });

  it('returns an empty list for non-array order responses', () => {
    expect(normalizeOrdersResponse({ data: { orders: 'bad' } })).toEqual([]);
    expect(normalizeOrdersResponse(null)).toEqual([]);
  });

  it('normalizes backend order detail responses with purchased items', () => {
    expect(
      normalizeOrderResponse({
        order: {
          id: 77,
          order_number: 'ORD-77',
          status: 'paid',
          subtotal: 40,
          tax_amount: 4,
          discount_amount: 2,
          tip_amount: 3,
          total: 45,
          created_at: '2026-06-25T10:00:00Z',
          paid_at: '2026-06-25T10:05:00Z',
          items: [
            {
              id: 9,
              name: 'Creatine',
              sku: 'SUP-CRE',
              quantity: 2,
              unit_price: 20,
              total: 40,
              status: 'paid',
              notes: 'Demo item',
            },
          ],
        },
      })
    ).toEqual({
      id: '77',
      number: 'ORD-77',
      status: 'paid',
      items: [
        {
          id: '9',
          name: 'Creatine',
          price: 20,
          quantity: 2,
          sku: 'SUP-CRE',
          total: 40,
          status: 'paid',
          notes: 'Demo item',
        },
      ],
      subtotal: 40,
      tax: 4,
      tip: 3,
      discount: 2,
      total: 45,
      createdAt: '2026-06-25T10:00:00Z',
      paidAt: '2026-06-25T10:05:00Z',
    });
  });
});
