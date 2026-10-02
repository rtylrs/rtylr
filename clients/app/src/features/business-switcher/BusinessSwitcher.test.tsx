import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';
import path from 'node:path';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

vi.mock('@voxire/auth', () => ({
  setCookie: vi.fn(),
  removeCookie: vi.fn(),
  getDefaultCookieDomain: () => 'rtylr.com',
}));

const navigateMock = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useNavigate: () => navigateMock,
}));

type OwnedBusiness = { id: number; name: string; logoUrl?: string };

const businessesData = vi.hoisted(() => ({
  data: { businesses: [] as OwnedBusiness[] } as { businesses: OwnedBusiness[] } | undefined,
}));

vi.mock('../../shared/hooks/queries/useBusinesses', () => ({
  useListMyBusinesses: () => businessesData,
  useSwitchBusiness: () => ({ mutate: vi.fn(), isPending: false }),
}));

vi.mock('../../shared/stores/userStore', () => ({
  useUserStore: () => ({ user: { businessId: 1, businessName: 'Acme' } }),
}));

import BusinessSwitcher from './BusinessSwitcher';

// Search only appears once there are enough businesses to warrant it.
const manyBusinesses: OwnedBusiness[] = Array.from({ length: 7 }, (_, i) => ({
  id: i + 1,
  name: `Business ${i + 1}`,
}));

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

beforeEach(() => {
  vi.clearAllMocks();
  businessesData.data = { businesses: manyBusinesses };
});

describe('BusinessSwitcher: search-empty state', () => {
  it('shows the illustration, title and a clear-search call-to-action when a search matches nothing', () => {
    render(<BusinessSwitcher />);

    fireEvent.keyDown(screen.getByRole('button', { name: /Business 1/ }), { key: 'Enter' });
    fireEvent.change(screen.getByPlaceholderText('Search businesses...'), {
      target: { value: 'no such business' },
    });

    expect(screen.getByText('No businesses found')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Clear search' })).toBeInTheDocument();
  });

  it('clears the search query when the call-to-action is clicked', () => {
    render(<BusinessSwitcher />);

    fireEvent.keyDown(screen.getByRole('button', { name: /Business 1/ }), { key: 'Enter' });
    const searchInput = screen.getByPlaceholderText('Search businesses...');
    fireEvent.change(searchInput, { target: { value: 'no such business' } });
    expect(screen.getByText('No businesses found')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Clear search' }));

    expect(screen.queryByText('No businesses found')).not.toBeInTheDocument();
    expect(screen.getByText('Business 2')).toBeInTheDocument();
  });

  it('does not show the empty state when the search matches a business', () => {
    render(<BusinessSwitcher />);

    fireEvent.keyDown(screen.getByRole('button', { name: /Business 1/ }), { key: 'Enter' });
    fireEvent.change(screen.getByPlaceholderText('Search businesses...'), {
      target: { value: 'Business 2' },
    });

    expect(screen.queryByText('No businesses found')).not.toBeInTheDocument();
  });
});
