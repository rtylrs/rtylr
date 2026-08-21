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

const mockGetRevenueByApp = vi.fn();
vi.mock('../../../../shared/lib/clients', () => ({
  revenueClient: { getRevenueByApp: (...args: unknown[]) => mockGetRevenueByApp(...args) },
}));

import RevenueByAppWidget from './RevenueByAppWidget';

function renderWithClient(): void {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <RevenueByAppWidget />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('RevenueByAppWidget empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when there is no revenue', async () => {
    mockGetRevenueByApp.mockResolvedValue({ apps: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByText('dashboard.revenueByAppWidget.emptyTitle')).toBeInTheDocument()
    );
    expect(screen.getByText('dashboard.revenueByAppWidget.emptyDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches when the refresh call-to-action is clicked', async () => {
    mockGetRevenueByApp.mockResolvedValue({ apps: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument()
    );
    const callsBefore = mockGetRevenueByApp.mock.calls.length;
    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    await waitFor(() => expect(mockGetRevenueByApp.mock.calls.length).toBeGreaterThan(callsBefore));
  });

  it('does not show the empty state when there is revenue', async () => {
    mockGetRevenueByApp.mockResolvedValue({
      apps: [{ app: 'pos', amount: 1000, percentage: 100 }],
    });

    renderWithClient();

    await waitFor(() => expect(screen.getByText('pos')).toBeInTheDocument());
    expect(screen.queryByText('dashboard.revenueByAppWidget.emptyTitle')).not.toBeInTheDocument();
  });
});
