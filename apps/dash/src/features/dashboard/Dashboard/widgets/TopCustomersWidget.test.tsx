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

import TopCustomersWidget from './TopCustomersWidget';

function renderWithClient(): void {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <TopCustomersWidget />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('TopCustomersWidget empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when there are no customers', async () => {
    mockGet.mockResolvedValue({ customers: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByText('dashboard.topCustomersWidget.emptyTitle')).toBeInTheDocument()
    );
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches when the refresh call-to-action is clicked', async () => {
    mockGet.mockResolvedValue({ customers: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument()
    );
    const callsBefore = mockGet.mock.calls.length;
    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    await waitFor(() => expect(mockGet.mock.calls.length).toBeGreaterThan(callsBefore));
  });

  it('does not show the empty state when there are top customers', async () => {
    mockGet.mockResolvedValue({
      customers: [{ id: 1, name: 'Alex', lifetimeSpendMinorUnits: 5000, orderCount: 3 }],
    });

    renderWithClient();

    await waitFor(() => expect(screen.getByText('Alex')).toBeInTheDocument());
    expect(screen.queryByText('dashboard.topCustomersWidget.emptyTitle')).not.toBeInTheDocument();
  });
});
