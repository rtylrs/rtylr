import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, afterEach, beforeAll } from 'vitest';

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves,
// which is how this suite passed while every string on screen was a dotted
// path (rtylr#1269). The real locale files mean the copy below also proves the
// lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateSpy = vi.hoisted(() => vi.fn());
// The search string decides whether the checkout forwards a slug to the
// order-status page, so each test can set it.
const searchState = vi.hoisted(() => ({ value: '' }));
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateSpy,
  useRouterState: () => searchState.value,
}));

vi.mock('../../shared/lib/clients', () => ({
  orderClient: { createPublicOrder: vi.fn() },
}));

import { useCartStore } from '../../shared/stores/cartStore';
import { orderClient } from '../../shared/lib/clients';
import Checkout from './Checkout';
import { initTestI18n } from '../../shared/testing/i18n';

const createPublicOrder = vi.mocked(orderClient.createPublicOrder);

beforeAll(async () => {
  await initTestI18n('en');
});

describe('Checkout empty cart', () => {
  beforeEach(() => {
    useCartStore.getState().clearCart();
  });

  it('shows the shared EmptyState with the emptyCart title and a working back action', () => {
    const backSpy = vi.spyOn(window.history, 'back').mockImplementation(() => undefined);

    render(<Checkout />);

    expect(screen.getByText('Your cart is empty')).toBeInTheDocument();

    const backButton = screen.getByRole('button', { name: 'Back to Menu' });
    fireEvent.click(backButton);

    expect(backSpy).toHaveBeenCalledTimes(1);
    backSpy.mockRestore();
  });
});

describe('Checkout form', () => {
  beforeEach(() => {
    useCartStore.getState().clearCart();
    useCartStore.getState().setBusinessId(1);
    useCartStore.getState().addItem({ menuItemId: 7, name: 'Falafel wrap', unitPrice: 5.5 });
  });

  // The empty-cart case above renders four strings. Everything else on this
  // screen, the whole form, sits behind a non-empty cart, so without this the
  // suite proves nothing about the copy a customer who is actually ordering
  // reads.
  it('renders the form copy from the locale files rather than its keys', () => {
    render(<Checkout />);

    expect(screen.getByRole('heading', { level: 1, name: 'Checkout' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Order Summary' })).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Your Info' })).toBeInTheDocument();
    expect(screen.getByText('Total')).toBeInTheDocument();
    expect(screen.getByLabelText('Name *')).toBeInTheDocument();
    expect(screen.getByLabelText('Phone Number *')).toBeInTheDocument();
    expect(screen.getByLabelText('Email (optional)')).toBeInTheDocument();
    expect(screen.getByLabelText('Order Notes (optional)')).toBeInTheDocument();
    expect(screen.getByPlaceholderText('Your name')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: '+ Add note' })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Go back' })).toBeInTheDocument();
  });

  // The price is isolated from the label rather than concatenated into it.
  // In Arabic the concatenated form renders as 11.00$: the currency sign is a
  // bidi-neutral and takes the direction of the text around it. happy-dom has
  // no bidi layout, so this pins the structure that prevents it instead.
  it('places the price beside the translated submit label, in its own bdi', () => {
    render(<Checkout />);

    const button = screen.getByRole('button', { name: 'Place Order · $5.50' });
    const isolated = button.querySelector('bdi');

    expect(isolated).not.toBeNull();
    expect(isolated).toHaveTextContent('$5.50');
  });

  it('shows a translated validation message rather than a key', () => {
    render(<Checkout />);

    fireEvent.click(screen.getByRole('button', { name: 'Place Order · $5.50' }));

    expect(screen.getByRole('alert')).toHaveTextContent('Please enter your name.');
  });
});

describe('Checkout validation and submit', () => {
  beforeEach(() => {
    useCartStore.getState().clearCart();
    useCartStore.getState().setBusinessId(1);
    useCartStore.getState().addItem({ menuItemId: 7, name: 'Falafel wrap', unitPrice: 5.5 });
    navigateSpy.mockClear();
    createPublicOrder.mockReset();
    searchState.value = '';
  });

  it('requires a phone number after the name is filled', () => {
    render(<Checkout />);

    fireEvent.change(screen.getByLabelText('Name *'), { target: { value: 'Sam' } });
    fireEvent.click(screen.getByRole('button', { name: /Place Order/ }));

    expect(screen.getByRole('alert')).toHaveTextContent('Please enter your phone number.');
    expect(createPublicOrder).not.toHaveBeenCalled();
  });

  it('rejects a malformed email', () => {
    render(<Checkout />);

    fireEvent.change(screen.getByLabelText('Name *'), { target: { value: 'Sam' } });
    fireEvent.change(screen.getByLabelText('Phone Number *'), { target: { value: '+1 555 0199' } });
    fireEvent.change(screen.getByLabelText('Email (optional)'), {
      target: { value: 'not-an-email' },
    });
    fireEvent.click(screen.getByRole('button', { name: /Place Order/ }));

    expect(screen.getByRole('alert')).toHaveTextContent('Please enter a valid email address.');
    expect(createPublicOrder).not.toHaveBeenCalled();
  });

  it('edits an item note and carries it into the request', async () => {
    createPublicOrder.mockResolvedValue({
      orderId: '9001',
      orderNumber: 'ORD-9001',
      publicToken: 'tok-abc',
    });

    render(<Checkout />);

    fireEvent.click(screen.getByRole('button', { name: '+ Add note' }));
    fireEvent.change(screen.getByPlaceholderText('e.g. no pickles, extra sauce...'), {
      target: { value: 'no pickles' },
    });
    expect(useCartStore.getState().items[0].notes).toBe('no pickles');

    fireEvent.change(screen.getByLabelText('Name *'), { target: { value: 'Sam' } });
    fireEvent.change(screen.getByLabelText('Phone Number *'), { target: { value: '+1 555 0199' } });
    fireEvent.click(screen.getByRole('button', { name: /Place Order/ }));

    await waitFor(() => expect(createPublicOrder).toHaveBeenCalledTimes(1));
    expect(createPublicOrder.mock.calls[0][0]).toMatchObject({
      ownerId: '1',
      customerName: 'Sam',
      customerPhone: '+1 555 0199',
      items: [
        { menuItemId: '7', name: 'Falafel wrap', quantity: 1, unitPrice: 5.5, notes: 'no pickles' },
      ],
    });
  });

  it('clears the cart and navigates to the order status page with the slug and token', async () => {
    searchState.value = '?slug=acme';
    createPublicOrder.mockResolvedValue({
      orderId: '9001',
      orderNumber: 'ORD-9001',
      publicToken: 'tok-abc',
    });

    render(<Checkout />);

    fireEvent.change(screen.getByLabelText('Name *'), { target: { value: 'Sam' } });
    fireEvent.change(screen.getByLabelText('Phone Number *'), { target: { value: '+1 555 0199' } });
    fireEvent.click(screen.getByRole('button', { name: /Place Order/ }));

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({
        to: '/menu/order-status/9001?number=ORD-9001&business_id=1&slug=acme&public_token=tok-abc',
      })
    );
    expect(useCartStore.getState().items).toHaveLength(0);
  });

  it('shows an error and keeps the cart when the order request fails', async () => {
    createPublicOrder.mockRejectedValue(new Error('boom'));

    render(<Checkout />);

    fireEvent.change(screen.getByLabelText('Name *'), { target: { value: 'Sam' } });
    fireEvent.change(screen.getByLabelText('Phone Number *'), { target: { value: '+1 555 0199' } });
    fireEvent.click(screen.getByRole('button', { name: /Place Order/ }));

    await waitFor(() =>
      expect(screen.getByRole('alert')).toHaveTextContent(
        'Failed to place order. Please try again.'
      )
    );
    expect(navigateSpy).not.toHaveBeenCalled();
    expect(useCartStore.getState().items).toHaveLength(1);
    expect(screen.getByRole('button', { name: /Place Order/ })).toBeEnabled();
  });

  it('disables the submit button while the order is in flight', () => {
    // A promise that never settles keeps the submitting state observable.
    createPublicOrder.mockReturnValue(new Promise(() => {}));

    render(<Checkout />);

    fireEvent.change(screen.getByLabelText('Name *'), { target: { value: 'Sam' } });
    fireEvent.change(screen.getByLabelText('Phone Number *'), { target: { value: '+1 555 0199' } });
    fireEvent.click(screen.getByRole('button', { name: /Place Order/ }));

    expect(screen.getByRole('button', { name: 'Placing Order...' })).toBeDisabled();
    expect(createPublicOrder).toHaveBeenCalledTimes(1);
  });
});

describe('Checkout ordering kill switch', () => {
  afterEach(() => {
    vi.unstubAllEnvs();
  });

  it('shows the disabled screen when VITE_ONLINE_ORDERING_ENABLED is false', () => {
    vi.stubEnv('VITE_ONLINE_ORDERING_ENABLED', 'false');
    useCartStore.getState().clearCart();
    useCartStore.getState().addItem({ menuItemId: 7, name: 'Falafel wrap', unitPrice: 5.5 });

    render(<Checkout />);

    expect(screen.getByText('Online Ordering Disabled')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: /Place Order/ })).not.toBeInTheDocument();
  });
});
