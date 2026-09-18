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
