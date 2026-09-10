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
