import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/api', () => ({
  unwrap: (v: unknown) => v,
  toFiniteNumber: (v: unknown) => {
    const n = Number(v);
    return Number.isFinite(n) ? n : 0;
  },
}));

const mockGetBestSellers = vi.fn();
vi.mock('../../../../shared/lib/clients', () => ({
  insightsProductClient: { getBestSellers: (...args: unknown[]) => mockGetBestSellers(...args) },
}));

import TopProductsWidget from './TopProductsWidget';

function renderWithClient(): void {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <TopProductsWidget />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('TopProductsWidget empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when there are no products', async () => {
    mockGetBestSellers.mockResolvedValue({ products: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByText('dashboard.topProductsWidget.emptyTitle')).toBeInTheDocument()
    );
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches when the refresh call-to-action is clicked', async () => {
    mockGetBestSellers.mockResolvedValue({ products: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument()
    );
    const callsBefore = mockGetBestSellers.mock.calls.length;
    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    await waitFor(() => expect(mockGetBestSellers.mock.calls.length).toBeGreaterThan(callsBefore));
  });

  it('does not show the empty state when there are products', async () => {
    mockGetBestSellers.mockResolvedValue({
      products: [{ name: 'Espresso', unitsSold: 10, revenue: 25 }],
    });

    renderWithClient();

    await waitFor(() => expect(screen.getByText('Espresso')).toBeInTheDocument());
    expect(screen.queryByText('dashboard.topProductsWidget.emptyTitle')).not.toBeInTheDocument();
  });
});
