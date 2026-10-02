import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: () => ({ userId: 'user-1' }),
}));

type MenuEntry = { _id: string; menuName: string; items: unknown[]; isActive: boolean };

const menusData = vi.hoisted(() => ({
  data: [] as MenuEntry[],
  isLoading: false,
}));

vi.mock('../../shared/hooks/queries/useMenu', () => ({
  useGetAllMenus: () => menusData,
}));

import Menus from './Menus';

beforeEach(() => {
  vi.clearAllMocks();
  menusData.data = [];
  menusData.isLoading = false;
});

describe('Menus list: true-empty state', () => {
  it('shows the illustration, title and a create-item call-to-action when there are no menus', () => {
    render(<Menus />);

    expect(screen.getByText('menu.emptyNoMenusTitle')).toBeInTheDocument();
    expect(screen.getByText('menu.emptyNoMenuItemsYet')).toBeInTheDocument();
  });

  it('navigates to the create-item screen when the true-empty call-to-action is clicked', () => {
    render(<Menus />);

    fireEvent.click(screen.getByRole('button', { name: 'menu.createItem' }));

    expect(navigateMock).toHaveBeenCalledWith({ to: '/create-item' });
  });
});

describe('Menus list: search-empty state', () => {
  it('shows the search-empty title and a clear-search call-to-action when a search matches nothing', () => {
    menusData.data = [{ _id: '1', menuName: 'Lunch Menu', items: [], isActive: true }];
    render(<Menus />);

    fireEvent.change(screen.getByPlaceholderText('Search menus'), {
      target: { value: 'no such menu' },
    });

    expect(screen.getByText('menu.emptyTryAnotherSearch')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.clearSearch' })).toBeInTheDocument();
    expect(screen.queryByText('menu.emptyNoMenuItemsYet')).not.toBeInTheDocument();
  });

  it('clears the search query when the search-empty call-to-action is clicked', () => {
    menusData.data = [{ _id: '1', menuName: 'Lunch Menu', items: [], isActive: true }];
    render(<Menus />);

    const searchInput = screen.getByPlaceholderText('Search menus');
    fireEvent.change(searchInput, { target: { value: 'no such menu' } });
    expect(screen.getByText('menu.emptyTryAnotherSearch')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.clearSearch' }));

    expect(screen.queryByText('menu.emptyTryAnotherSearch')).not.toBeInTheDocument();
    expect(screen.getByText('Lunch Menu')).toBeInTheDocument();
  });

  it('does not show the empty state when menus match', () => {
    menusData.data = [{ _id: '1', menuName: 'Lunch Menu', items: [], isActive: true }];
    render(<Menus />);

    expect(screen.queryByText('menu.emptyNoMenusTitle')).not.toBeInTheDocument();
  });
});
