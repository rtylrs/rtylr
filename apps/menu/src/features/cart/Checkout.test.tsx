import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves,
// which is how this suite passed while every string on screen was a dotted
// path (rtylr#1269). The real locale files mean the copy below also proves the
// lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
  useRouterState: () => '',
}));

vi.mock('../../shared/lib/clients', () => ({
  orderClient: { createPublicOrder: vi.fn() },
}));

import { useCartStore } from '../../shared/stores/cartStore';
import Checkout from './Checkout';
import { initTestI18n } from '../../shared/testing/i18n';

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

  it('places the price beside the translated submit label', () => {
    render(<Checkout />);

    expect(screen.getByRole('button', { name: 'Place Order · $5.50' })).toBeInTheDocument();
  });

  it('shows a translated validation message rather than a key', () => {
    render(<Checkout />);

    fireEvent.click(screen.getByRole('button', { name: 'Place Order · $5.50' }));

    expect(screen.getByRole('alert')).toHaveTextContent('Please enter your name.');
  });
});
