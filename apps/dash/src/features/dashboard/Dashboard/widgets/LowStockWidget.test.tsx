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

const mockGetLowStock = vi.fn();
vi.mock('../../../../shared/lib/clients', () => ({
  inventoryClient: { getLowStock: (...args: unknown[]) => mockGetLowStock(...args) },
}));

import LowStockWidget from './LowStockWidget';

function renderWithClient(): void {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <LowStockWidget />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('LowStockWidget empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when stock is healthy', async () => {
    mockGetLowStock.mockResolvedValue({ items: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByText('dashboard.lowStockWidget.emptyTitle')).toBeInTheDocument()
    );
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches when the refresh call-to-action is clicked', async () => {
    mockGetLowStock.mockResolvedValue({ items: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument()
    );
    const callsBefore = mockGetLowStock.mock.calls.length;
    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    await waitFor(() => expect(mockGetLowStock.mock.calls.length).toBeGreaterThan(callsBefore));
  });

  it('does not show the empty state when items are low on stock', async () => {
    mockGetLowStock.mockResolvedValue({ items: [{ productName: 'Cups', stockQuantity: 2 }] });

    renderWithClient();

    await waitFor(() => expect(screen.getByText('Cups')).toBeInTheDocument());
    expect(screen.queryByText('dashboard.lowStockWidget.emptyTitle')).not.toBeInTheDocument();
  });
});
