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
  // Contract field names. hub.dashboard.v1.GetKPIsResponse carries revenue,
  // orderCount, staffCount, customerCount, lowStockCount and activeApps, and
  // nothing else. The previous fixture used revenueToday/ordersToday/
  // newCustomers/revenueSeries, none of which exist on it, so it agreed with the
  // schema and disagreed with the server.
  data: undefined as { revenue: number; orderCount: number; customerCount: number } | undefined,
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
  kpisState.data = { revenue: 0, orderCount: 0, customerCount: 0 };
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

/**
 * rtylr#947. The stat cards read revenueToday, ordersToday and newCustomers.
 * None of those exist on GetKPIsResponse, and every one was .optional() with a
 * transform default of 0, so the screen showed $0.00 / 0 / 0 no matter how much
 * business had happened.
 *
 * revenue and orderCount are genuinely today-scoped server-side: onion's
 * kpis.sql defines them as SumRevenueToday and CountOrdersToday, both filtered
 * to `created_at >= CURDATE()`. So the "Today" labels were right all along and
 * only the field names were wrong.
 */
describe('Analytics stat cards read the contract fields (rtylr#947)', () => {
  it('shows revenue rather than a permanent zero', () => {
    kpisState.data = { revenue: 1234.5, orderCount: 10, customerCount: 7 };
    render(<Analytics />);
    expect(screen.getByText('$1234.50')).toBeInTheDocument();
  });

  it('shows the order count', () => {
    kpisState.data = { revenue: 1234.5, orderCount: 10, customerCount: 7 };
    render(<Analytics />);
    expect(screen.getByText('10')).toBeInTheDocument();
  });

  it('derives average order value from revenue and order count', () => {
    kpisState.data = { revenue: 250, orderCount: 4, customerCount: 0 };
    render(<Analytics />);
    // 250 / 4
    expect(screen.getByText('$62.50')).toBeInTheDocument();
  });

  it('avoids dividing by zero when nothing was ordered', () => {
    kpisState.data = { revenue: 0, orderCount: 0, customerCount: 0 };
    render(<Analytics />);
    // Both the revenue card and the derived average read $0.00 here, so this
    // asserts on the pair rather than a single match. A NaN from 0/0 would
    // render as "$NaN" and fail.
    expect(screen.getAllByText('$0.00')).toHaveLength(2);
    expect(screen.queryByText(/NaN/)).not.toBeInTheDocument();
  });

  it('shows the customer count', () => {
    kpisState.data = { revenue: 0, orderCount: 0, customerCount: 42 };
    render(<Analytics />);
    expect(screen.getByText('42')).toBeInTheDocument();
  });
});

/**
 * rtylr#946. The 7d/30d/90d controls sent nothing: GetKPIsRequest is an empty
 * message and GetBestSellersRequest carries only `limit`, so the range reached
 * the queryKey and nowhere else and every view showed identical numbers under a
 * different heading.
 */
describe('Analytics range control is gated, not decorative (rtylr#946)', () => {
  it('marks every range tab disabled rather than letting it look interactive', () => {
    render(<Analytics />);
    const tabs = screen.getAllByRole('tab');
    expect(tabs.length).toBeGreaterThan(0);
    for (const tab of tabs) expect(tab).toHaveAttribute('aria-disabled', 'true');
  });

  it('states why the range cannot be changed', () => {
    render(<Analytics />);
    expect(screen.getByText('analytics.rangeUnavailable')).toBeInTheDocument();
  });

  it('uses aria-disabled rather than the native attribute, so the reason stays reachable', () => {
    render(<Analytics />);
    // A natively disabled button is skipped by some screen-reader navigation,
    // which would hide the explanation the tab is described by.
    for (const tab of screen.getAllByRole('tab')) expect(tab).not.toBeDisabled();
  });
});
