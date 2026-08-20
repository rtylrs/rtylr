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

const mockGet = vi.fn();
vi.mock('../../../../shared/lib/api', () => ({
  safe: { get: (...args: unknown[]) => mockGet(...args) },
}));

import InvoiceAgingWidget from './InvoiceAgingWidget';

function renderWithClient(): void {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <InvoiceAgingWidget />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('InvoiceAgingWidget empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when every bucket is empty', async () => {
    mockGet.mockResolvedValue({
      buckets: [
        { label: '0-30', amountMinorUnits: 0, count: 0 },
        { label: '31-60', amountMinorUnits: 0, count: 0 },
      ],
    });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByText('dashboard.invoiceAgingWidget.emptyTitle')).toBeInTheDocument()
    );
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches when the refresh call-to-action is clicked', async () => {
    mockGet.mockResolvedValue({ buckets: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument()
    );
    const callsBefore = mockGet.mock.calls.length;
    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    await waitFor(() => expect(mockGet.mock.calls.length).toBeGreaterThan(callsBefore));
  });

  it('does not show the empty state when a bucket has outstanding invoices', async () => {
    mockGet.mockResolvedValue({ buckets: [{ label: '0-30', amountMinorUnits: 5000, count: 2 }] });

    renderWithClient();

    await waitFor(() => expect(screen.getByText('0-30 days')).toBeInTheDocument());
    expect(screen.queryByText('dashboard.invoiceAgingWidget.emptyTitle')).not.toBeInTheDocument();
  });
});
