import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}));

vi.mock('../../shared/lib/api', () => ({
  default: { post: vi.fn(), get: vi.fn() },
}));

import { useCartStore } from '../../shared/stores/cartStore';
import Checkout from './Checkout';

describe('Checkout empty cart', () => {
  beforeEach(() => {
    useCartStore.getState().clearCart();
  });

  it('shows the shared EmptyState with the emptyCart title and a working back action', () => {
    const backSpy = vi.spyOn(window.history, 'back').mockImplementation(() => undefined);

    render(<Checkout />);

    expect(screen.getByText('emptyCart')).toBeInTheDocument();

    const backButton = screen.getByRole('button', { name: 'backToMenu' });
    fireEvent.click(backButton);

    expect(backSpy).toHaveBeenCalledTimes(1);
    backSpy.mockRestore();
  });
});
