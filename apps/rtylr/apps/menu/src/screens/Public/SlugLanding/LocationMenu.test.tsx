import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves,
// which is how this suite passed while every string on screen was a dotted
// path (rtylr#1269). The real locale files mean the copy below also proves the
// lookup.

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

const cartStoreMock = vi.hoisted(() => ({
  businessId: 0,
  clearCart: vi.fn(),
  setBusinessId: vi.fn((id: number) => {
    cartStoreMock.businessId = id;
  }),
  setLocationId: vi.fn(),
  addItem: vi.fn(),
  updateQuantity: vi.fn(),
}));

vi.mock('../../../shared/stores/cartStore', () => ({
  useCartStore: () => cartStoreMock,
}));

import LocationMenu from './LocationMenu';
import { initTestI18n } from '../../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

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

    expect(screen.getByText('Business not found')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: 'Go back' }));
    expect(backSpy).toHaveBeenCalledTimes(1);
  });

  it('shows location-not-found and navigates back to the business on the CTA', () => {
    businessQuery.isError = false;
    businessQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      locations: [],
    };

    render(<LocationMenu />);

    expect(screen.getByText('Location not found')).toBeInTheDocument();

    const cta = screen.getByRole('button', {
      name: 'Back to Acme Diner',
    });
    fireEvent.click(cta);

    expect(navigateSpy).toHaveBeenCalledWith({ to: '/$slug', params: { slug: 'acme' } });
  });

  it('shows a translated message when the menu itself fails to load', () => {
    businessQuery.isError = false;
    businessQuery.data = {
      business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
      locations: [{ id: 5, name: 'Downtown', slug: 'downtown', address: '', timezone: 'UTC' }],
    };
    menuQuery.isError = true;

    render(<LocationMenu />);

    expect(screen.getByRole('heading', { name: "Couldn't load the menu" })).toBeInTheDocument();
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

    expect(screen.getByText('No menu items available')).toBeInTheDocument();

    const cta = screen.getByRole('button', {
      name: 'Back to Acme Diner',
    });
    fireEvent.click(cta);

    expect(navigateSpy).toHaveBeenCalledWith({ to: '/$slug', params: { slug: 'acme' } });
  });
});

describe('LocationMenu cart reconciliation', () => {
  const acmeBusiness = {
    business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
    locations: [{ id: 5, name: 'Downtown', slug: 'downtown', address: '', timezone: 'UTC' }],
  };
  const acmeMenu = {
    business: { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' },
    items: [{ id: 10, name: 'Falafel wrap', salePrice: 5, isActive: true }],
  };

  beforeEach(() => {
    businessQuery.isLoading = false;
    businessQuery.isError = false;
    businessQuery.data = acmeBusiness;
    menuQuery.isLoading = false;
    menuQuery.isError = false;
    menuQuery.data = acmeMenu;
    cartStoreMock.businessId = 0;
    cartStoreMock.clearCart.mockClear();
    cartStoreMock.setBusinessId.mockClear();
    cartStoreMock.setLocationId.mockClear();
  });

  it('clears a persisted cart from another business before adopting this business', () => {
    cartStoreMock.businessId = 99; // abandoned checkout at a different restaurant

    render(<LocationMenu />);

    expect(cartStoreMock.clearCart).toHaveBeenCalledTimes(1);
    expect(cartStoreMock.clearCart.mock.invocationCallOrder[0]).toBeLessThan(
      cartStoreMock.setBusinessId.mock.invocationCallOrder[0]
    );
    expect(cartStoreMock.setBusinessId).toHaveBeenCalledWith(1);
  });

  it('keeps the persisted cart when it already belongs to this business', () => {
    cartStoreMock.businessId = 1;

    render(<LocationMenu />);

    expect(cartStoreMock.clearCart).not.toHaveBeenCalled();
    expect(cartStoreMock.setBusinessId).toHaveBeenCalledWith(1);
  });

  it('does not clear an empty cart', () => {
    cartStoreMock.businessId = 0;

    render(<LocationMenu />);

    expect(cartStoreMock.clearCart).not.toHaveBeenCalled();
  });
});
