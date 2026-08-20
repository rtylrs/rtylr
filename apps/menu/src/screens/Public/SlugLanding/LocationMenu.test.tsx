import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) =>
      opts ? `${key}:${JSON.stringify(opts)}` : key,
  }),
}));

vi.mock('@voxire/components', () => ({
  FullScreenLoader: () => null,
  HelmetContainer: () => null,
}));

vi.mock('@voxire/menu', () => ({
  GridLayout: () => null,
}));

const navigateSpy = vi.hoisted(() => vi.fn());
const backSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ slug: 'acme', location: 'downtown' }),
  useNavigate: () => navigateSpy,
  useRouter: () => ({ history: { back: backSpy } }),
}));

const businessQuery = vi.hoisted(() => ({
  data: undefined as unknown,
  isLoading: false,
  isError: true,
}));
const menuQuery = vi.hoisted(() => ({
  data: undefined as unknown,
  isLoading: false,
  isError: false,
}));
vi.mock('../../../shared/hooks/usePublicMenu', () => ({
  usePublicBusinessBySlug: () => businessQuery,
  usePublicMenu: () => menuQuery,
}));

vi.mock('../../../shared/stores/cartStore', () => ({
  useCartStore: () => ({
    clearCart: vi.fn(),
    setBusinessId: vi.fn(),
    setLocationId: vi.fn(),
    addItem: vi.fn(),
    updateQuantity: vi.fn(),
  }),
}));

import LocationMenu from './LocationMenu';

describe('LocationMenu empty states', () => {
  beforeEach(() => {
    navigateSpy.mockClear();
    backSpy.mockClear();
    businessQuery.data = undefined;
    businessQuery.isLoading = false;
    businessQuery.isError = true;
    menuQuery.data = undefined;
    menuQuery.isLoading = false;
    menuQuery.isError = false;
  });

  it('shows business-not-found and goes back on the CTA', () => {
    render(<LocationMenu />);

    expect(screen.getByText('businessNotFoundTitle')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'goBack' }));
    expect(backSpy).toHaveBeenCalledTimes(1);
  });

  it('shows location-not-found and navigates back to the business on the CTA', () => {
    businessQuery.isError = false;
    businessQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      locations: [],
    };

    render(<LocationMenu />);

    expect(screen.getByText('locationNotFoundTitle')).toBeInTheDocument();

    const cta = screen.getByRole('button', {
      name: 'backToBusiness:{"businessName":"Acme Diner"}',
    });
    fireEvent.click(cta);

    expect(navigateSpy).toHaveBeenCalledWith({ to: '/$slug', params: { slug: 'acme' } });
  });

  it('shows no-items-available and navigates back to the business on the CTA', () => {
    businessQuery.isError = false;
    businessQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      locations: [{ id: 5, name: 'Downtown', slug: 'downtown', address: '', timezone: 'UTC' }],
    };
    menuQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      items: [],
    };

    render(<LocationMenu />);

    expect(screen.getByText('noItemsTitle')).toBeInTheDocument();

    const cta = screen.getByRole('button', {
      name: 'backToBusiness:{"businessName":"Acme Diner"}',
    });
    fireEvent.click(cta);

    expect(navigateSpy).toHaveBeenCalledWith({ to: '/$slug', params: { slug: 'acme' } });
  });
});
