import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/api', () => ({
  unwrap: (v: unknown) => v,
}));

const mockGet = vi.fn();
vi.mock('../../../../shared/lib/api', () => ({
  safe: { get: (...args: unknown[]) => mockGet(...args) },
}));

import PendingWidget from './PendingWidget';

function renderWithClient(): void {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  render(
    <QueryClientProvider client={client}>
      <PendingWidget />
    </QueryClientProvider>
  );
}

beforeEach(() => {
  vi.clearAllMocks();
});

describe('PendingWidget empty state', () => {
  it('shows the illustration, title and a refresh call-to-action when nothing is pending', async () => {
    mockGet.mockResolvedValue({ requests: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByText('dashboard.pendingWidget.emptyTitle')).toBeInTheDocument()
    );
    expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument();
  });

  it('refetches when the refresh call-to-action is clicked', async () => {
    mockGet.mockResolvedValue({ requests: [] });

    renderWithClient();

    await waitFor(() =>
      expect(screen.getByRole('button', { name: 'common.refresh' })).toBeInTheDocument()
    );
    const callsBefore = mockGet.mock.calls.length;
    fireEvent.click(screen.getByRole('button', { name: 'common.refresh' }));

    await waitFor(() => expect(mockGet.mock.calls.length).toBeGreaterThan(callsBefore));
  });

  it('does not show the empty state when there are pending requests', async () => {
    mockGet.mockResolvedValue({
      requests: [{ staffName: 'Jane', type: 'Time off', status: 'pending' }],
    });

    renderWithClient();

    await waitFor(() => expect(screen.getByText('Jane')).toBeInTheDocument());
    expect(screen.queryByText('dashboard.pendingWidget.emptyTitle')).not.toBeInTheDocument();
  });
});
