import React from 'react';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor, cleanup } from '@testing-library/react';
import { MemoryRouter } from 'react-router-dom';

const listOrdersMock = vi.hoisted(() => vi.fn());
const updateOrderStatusMock = vi.hoisted(() => vi.fn());
const voidOrderMock = vi.hoisted(() => vi.fn());
const resolveOpenSessionIdMock = vi.hoisted(() => vi.fn());

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => ({
    listOrders: listOrdersMock,
    updateOrderStatus: updateOrderStatusMock,
    voidOrder: voidOrderMock,
  }),
}));

vi.mock('../close-session/closeSessionWorkflow', () => ({
  resolveOpenSessionId: resolveOpenSessionIdMock,
}));

vi.mock('../../layouts/ScreenLayout', () => ({
  default: ({ title, children }: { title: string; children: React.ReactNode }) => (
    <div>
      <h1>{title}</h1>
      {children}
    </div>
  ),
}));

vi.mock('../../shared/hooks/useCapabilities', () => ({
  useCapabilities: () => ({ canVoid: true }),
}));

import OrdersScreen from './index';

describe('POS OrdersScreen online menu integration', () => {
  beforeEach(() => {
    listOrdersMock.mockReset();
    updateOrderStatusMock.mockReset();
    voidOrderMock.mockReset();
    resolveOpenSessionIdMock.mockReset();
  });

  afterEach(() => {
    cleanup();
  });

  it('shows incoming online menu orders by default without requiring an open shift session and confirms to kitchen', async () => {
    resolveOpenSessionIdMock.mockRejectedValue(new Error('No open register session'));
    listOrdersMock.mockResolvedValue({
      orders: [
        {
          id: '4',
          sessionId: '0',
          orderNumber: 'ORD-20261126-0002',
          status: 'pending',
          customerName: 'Anonymous Client Test',
          total: 5.5,
          createdAt: '2026-11-26T22:11:06Z',
          items: [{ id: '1', name: 'Iced Matcha Latte', quantity: 1 }],
        },
        {
          id: '5',
          sessionId: '4',
          orderNumber: 'ORD-20261126-0003',
          status: 'paid',
          total: 4.5,
          createdAt: '2026-11-26T22:12:00Z',
          items: [{ id: '2', name: 'Signature Espresso', quantity: 1 }],
        },
      ],
      total: 2,
      hasMore: false,
    });
    updateOrderStatusMock.mockResolvedValue({ message: 'Order status updated' });

    render(
      <MemoryRouter>
        <OrdersScreen />
      </MemoryRouter>
    );

    await waitFor(() => {
      expect(screen.getByText('#ORD-20261126-0002')).toBeTruthy();
    });

    // Default scope ('all') must omit sessionId so online menu orders are never hidden.
    expect(listOrdersMock).toHaveBeenCalledWith(
      expect.objectContaining({ sessionId: '' })
    );
    expect(screen.getAllByText('Online Menu').length).toBeGreaterThanOrEqual(2);
    expect(screen.getByText('Anonymous Client Test')).toBeTruthy();

    fireEvent.click(screen.getByRole('button', { name: 'Confirm & Send to Kitchen' }));

    await waitFor(() => {
      expect(updateOrderStatusMock).toHaveBeenCalledWith({
        orderId: '4',
        body: { status: 'confirmed' },
      });
    });

    // Switching to "Online Menu" filters strictly to orders without a register session.
    fireEvent.click(screen.getByRole('button', { name: 'Online Menu' }));
    await waitFor(() => {
      expect(screen.getByText('#ORD-20261126-0002')).toBeTruthy();
      expect(screen.queryByText('#ORD-20261126-0003')).toBeNull();
    });
  });

  it('scopes by open register session when This Shift is selected', async () => {
    resolveOpenSessionIdMock.mockResolvedValue(4);
    listOrdersMock.mockResolvedValue({ orders: [], total: 0, hasMore: false });

    render(
      <MemoryRouter>
        <OrdersScreen />
      </MemoryRouter>
    );

    await waitFor(() => expect(listOrdersMock).toHaveBeenCalled());

    fireEvent.click(screen.getByRole('button', { name: 'This Shift' }));

    await waitFor(() => {
      expect(listOrdersMock).toHaveBeenLastCalledWith(
        expect.objectContaining({ sessionId: '4' })
      );
    });
  });
});
