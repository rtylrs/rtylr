import React from 'react';
import { render, screen } from '@testing-library/react';
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
  GridLayout: ({ categories }: { categories: { name: string }[] }) => (
    <div data-testid="grid">{categories.map((c) => c.name).join(',')}</div>
  ),
}));

const params = vi.hoisted(() => ({ value: {} as Record<string, string> }));
vi.mock('@tanstack/react-router', () => ({
  useParams: () => params.value,
  useNavigate: () => vi.fn(),
}));

const menuQuery = vi.hoisted(() => ({
  data: undefined as unknown,
  isLoading: false,
  isError: false,
}));
vi.mock('../../../shared/hooks/usePublicMenu', () => ({
  usePublicMenu: () => menuQuery,
}));

vi.mock('../../../shared/stores/cartStore', () => ({
  useCartStore: () => ({
    clearCart: vi.fn(),
    setBusinessId: vi.fn(),
    addItem: vi.fn(),
  }),
}));

import PosMenu from './PosMenu';
import { initTestI18n } from '../../../shared/testing/i18n';

const business = { id: 1, name: 'Acme Diner', logoUrl: '', currency: 'USD', timezone: 'UTC' };

beforeAll(async () => {
  await initTestI18n('en');
});

describe('PosMenu', () => {
  beforeEach(() => {
    params.value = { businessId: '42' };
    menuQuery.data = undefined;
    menuQuery.isLoading = false;
    menuQuery.isError = false;
  });

  // `/b/:businessId` is one of the two URL shapes printed on QR codes, so a
  // link that has lost its id is a real state, not a developer mistake.
  it.each([[{}], [{ businessId: 'not-a-number' }], [{ businessId: '0' }]])(
    'explains an unusable URL for %o',
    (value) => {
      params.value = value as Record<string, string>;

      render(<PosMenu />);

      expect(screen.getByRole('heading', { name: 'Invalid menu URL' })).toBeInTheDocument();
      expect(screen.getByText('This link is missing a business ID.')).toBeInTheDocument();
    }
  );

  it('explains a failed fetch', () => {
    menuQuery.isError = true;

    render(<PosMenu />);

    expect(screen.getByRole('heading', { name: 'Failed to load menu' })).toBeInTheDocument();
    expect(screen.getByText('Please try refreshing the page.')).toBeInTheDocument();
  });

  it('explains an empty menu', () => {
    menuQuery.data = { business, items: [] };

    render(<PosMenu />);

    expect(screen.getByRole('heading', { name: 'No menu items available' })).toBeInTheDocument();
    expect(
      screen.getByText('This menu is currently empty. Please check back later.')
    ).toBeInTheDocument();
  });

  // The POS menu table is flat, so the screen invents the single category it
  // groups everything under. That name is copy a customer reads.
  it('names the synthesised category from the locale files', () => {
    menuQuery.data = {
      business,
      items: [{ id: 3, name: 'Falafel wrap', salePrice: 5.5, imageUrl: '', isActive: true }],
    };

    render(<PosMenu />);

    expect(screen.getByTestId('grid')).toHaveTextContent('Menu');
  });
});
