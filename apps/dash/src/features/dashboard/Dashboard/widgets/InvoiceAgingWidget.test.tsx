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

const mockGetInvoiceAging = vi.fn();
vi.mock('../../../../shared/lib/clients', () => ({
  insightsFinancialClient: {
    getInvoiceAging: (...args: unknown[]) => mockGetInvoiceAging(...args),
  },
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
    mockGetInvoiceAging.mockResolvedValue({
      aging: {
        buckets: [
          { bucket: '0-30', amount: 0, count: 0 },
          { bucket: '31-60', amount: 0, count: 0 },
        ],
      },
    });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByText('dashboard.invoiceAgingWidget.emptyTitle')).toBeInTheDocument()
    );
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches when the refresh call-to-action is clicked', async () => {
    mockGetInvoiceAging.mockResolvedValue({ aging: { buckets: [] } });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument()
    );
    const callsBefore = mockGetInvoiceAging.mock.calls.length;
    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    await waitFor(() => expect(mockGetInvoiceAging.mock.calls.length).toBeGreaterThan(callsBefore));
  });

  it('does not show the empty state when a bucket has outstanding invoices', async () => {
    mockGetInvoiceAging.mockResolvedValue({
      aging: { buckets: [{ bucket: '0-30', amount: 5000, count: 2 }] },
    });

    renderWithClient();

    await waitFor(() => expect(screen.getByText('0-30 days')).toBeInTheDocument());
    expect(screen.queryByText('dashboard.invoiceAgingWidget.emptyTitle')).not.toBeInTheDocument();
  });
});
