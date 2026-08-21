import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

type BestSeller = { productId: number; name: string; unitsSold: number; revenue: number };

const bestSellersState = vi.hoisted(() => ({
  data: [] as BestSeller[],
  isLoading: false,
  refetch: vi.fn(),
}));

vi.mock('@tanstack/react-query', () => ({
  useQuery: () => bestSellersState,
}));

const kpisState = vi.hoisted(() => ({
  data: undefined as
    | { revenueToday: number; ordersToday: number; newCustomers: number; revenueSeries: number[] }
    | undefined,
  isLoading: false,
}));

vi.mock('../../shared/hooks/queries/useKpis', () => ({
  useKpis: () => kpisState,
}));

import Analytics from './Analytics';

beforeEach(() => {
  vi.clearAllMocks();
  bestSellersState.data = [];
  bestSellersState.isLoading = false;
  kpisState.data = { revenueToday: 0, ordersToday: 0, newCustomers: 0, revenueSeries: [] };
  kpisState.isLoading = false;
});

describe('Analytics: top-items empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when there is no best-seller data', () => {
    render(<Analytics />);

    expect(screen.getByText('analytics.emptyNoDataTitle')).toBeInTheDocument();
    expect(screen.getByText('analytics.emptyNoDataDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches best sellers when the refresh call-to-action is clicked', () => {
    render(<Analytics />);

    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    expect(bestSellersState.refetch).toHaveBeenCalledTimes(1);
  });

  it('does not show the empty state when there is best-seller data', () => {
    bestSellersState.data = [{ productId: 1, name: 'Burger', unitsSold: 10, revenue: 50 }];
    render(<Analytics />);

    expect(screen.queryByText('analytics.emptyNoDataTitle')).not.toBeInTheDocument();
    expect(screen.getByText('Burger')).toBeInTheDocument();
  });
});
