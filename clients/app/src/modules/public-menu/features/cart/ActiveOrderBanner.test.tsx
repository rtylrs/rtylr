import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

const navigateSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateSpy,
}));

const getPublicOrderStatus = vi.hoisted(() => vi.fn());
vi.mock('../../shared/lib/clients', () => ({
  orderClient: { getPublicOrderStatus },
}));

import ActiveOrderBanner from './ActiveOrderBanner';
import {
  clearActiveOrders,
  getActiveOrdersForBusiness,
  saveActiveOrder,
} from '../../shared/stores/activeOrderStore';
import { initTestI18n } from '../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

describe('ActiveOrderBanner', () => {
  beforeEach(() => {
    clearActiveOrders();
    navigateSpy.mockClear();
    getPublicOrderStatus.mockReset();
  });

  it('renders nothing and makes no API call when there is no active order', () => {
    const { container } = render(<ActiveOrderBanner businessId={5} />);
    expect(container).toBeEmptyDOMElement();
    expect(getPublicOrderStatus).not.toHaveBeenCalled();
  });

  it('shows the active order banner, refreshes live status, and navigates back to order progress', async () => {
    getPublicOrderStatus.mockResolvedValue({ status: 'ready' });
    saveActiveOrder({
      orderId: '42',
      orderNumber: 'ORD-42',
      businessId: '5',
      slug: 'voxire',
      publicToken: 'tok-xyz',
      status: 'pending',
      createdAt: Date.now(),
    });

    render(<ActiveOrderBanner businessId={5} />);

    await waitFor(() => expect(screen.getByText('Order #ORD-42 · Ready')).toBeInTheDocument());

    fireEvent.click(screen.getByRole('button', { name: /Order #ORD-42/ }));
    expect(navigateSpy).toHaveBeenCalledWith({
      to: '/menu/order-status/42?number=ORD-42&business_id=5&slug=voxire&public_token=tok-xyz',
    });
  });

  it('auto-clears the banner when the server reports the order is completed', async () => {
    getPublicOrderStatus.mockResolvedValue({ status: 'completed' });
    saveActiveOrder({
      orderId: '43',
      orderNumber: 'ORD-43',
      businessId: '5',
      publicToken: 'tok-done',
      status: 'preparing',
      createdAt: Date.now(),
    });

    render(<ActiveOrderBanner businessId={5} />);

    await waitFor(() =>
      expect(screen.queryByTestId('active-order-banner')).not.toBeInTheDocument()
    );
    expect(getActiveOrdersForBusiness(5)).toHaveLength(0);
  });

  it('allows the customer to dismiss the banner manually', async () => {
    getPublicOrderStatus.mockResolvedValue({ status: 'preparing' });
    saveActiveOrder({
      orderId: '44',
      orderNumber: 'ORD-44',
      businessId: '5',
      publicToken: 'tok-keep',
      status: 'preparing',
      createdAt: Date.now(),
    });

    render(<ActiveOrderBanner businessId={5} />);

    expect(await screen.findByText('Order #ORD-44 · Preparing')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Dismiss order banner' }));

    expect(screen.queryByTestId('active-order-banner')).not.toBeInTheDocument();
    expect(getActiveOrdersForBusiness(5)).toHaveLength(0);
  });
});
