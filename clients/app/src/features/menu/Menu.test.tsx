import React from 'react';
import { render, screen } from '@testing-library/react';
import { describe, it, expect, beforeEach, vi } from 'vitest';

vi.mock('@voxire/components', () => ({
  Spinner: () => <div data-testid="spinner" />,
  HelmetContainer: () => null,
}));

vi.mock('@voxire/config', async (importOriginal) => ({
  ...(await importOriginal<Record<string, unknown>>()),
  cdnImageUrl: () => '',
}));

vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => vi.fn(),
}));

const queryState = vi.hoisted(() => ({ isLoading: false }));

vi.mock('../../shared/hooks/queries/useMenu', () => ({
  useGetMenu: () => queryState,
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: () => ({ userId: 'u1' }),
}));

const menuState = vi.hoisted(() => ({
  items: [] as unknown[],
  status: 'success' as string,
  id: 'pos-menu',
  deleteCategory: vi.fn(),
  deleteBanner: vi.fn(),
  deleteMenuItem: vi.fn(),
}));

vi.mock('../../shared/stores/menuStore', () => ({
  useMenuStore: () => menuState,
}));

import Menu from './Menu';

beforeEach(() => {
  menuState.items = [];
  menuState.status = 'success';
  queryState.isLoading = false;
});

describe('Menu item pricing', () => {
  // pos_menu_items / erp products store sale_price as DECIMAL(10,2), the
  // contract declares float64, and useMenu's toStoreItems copies salePrice
  // straight into details.price. So details.price is major units and the
  // screen must not divide it.
  const item = (price: number) => ({
    _id: 'm1',
    itemId: 'm1',
    name: 'Kebab',
    type: 'item' as const,
    details: { name: 'Kebab', price, sku: 'K1', isActive: true },
  });

  it('renders the price as major units', () => {
    menuState.items = [item(12.5)];
    render(<Menu />);
    expect(screen.getByText('$12.50')).toBeInTheDocument();
  });

  it('does not divide the price by a hundred', () => {
    menuState.items = [item(1250)];
    render(<Menu />);
    expect(screen.getByText('$1,250.00')).toBeInTheDocument();
    expect(screen.queryByText('$12.50')).not.toBeInTheDocument();
  });

  it('renders a zero price as zero rather than blank', () => {
    menuState.items = [item(0)];
    render(<Menu />);
    expect(screen.getByText('$0.00')).toBeInTheDocument();
  });

  // renderItems, the nested path, is a second copy of the price markup. The
  // flat-item tests above do not reach it, which a mutation probe caught: the
  // /100 could be put back on this line with every test still passing.
  const category = (price: number) => ({
    _id: 'c1',
    itemId: 'c1',
    name: 'Grills',
    type: 'category' as const,
    menuItems: [],
    details: { name: 'Grills', menuItems: [item(price)] },
  });

  it('renders a nested item price as major units', () => {
    menuState.items = [category(12.5)];
    render(<Menu />);
    expect(screen.getByText('$12.50')).toBeInTheDocument();
  });

  it('does not divide a nested item price by a hundred', () => {
    menuState.items = [category(1250)];
    render(<Menu />);
    expect(screen.getByText('$1,250.00')).toBeInTheDocument();
    expect(screen.queryByText('$12.50')).not.toBeInTheDocument();
  });

  it('renders a fractional price without losing the cents', () => {
    menuState.items = [item(9.99)];
    render(<Menu />);
    expect(screen.getByText('$9.99')).toBeInTheDocument();
  });

  it('renders a spinner while menu query is loading', () => {
    queryState.isLoading = true;
    render(<Menu />);
    expect(screen.getByTestId('spinner')).toBeInTheDocument();
    expect(screen.queryByText('No POS menu items yet.')).not.toBeInTheDocument();
    queryState.isLoading = false;
  });
});
