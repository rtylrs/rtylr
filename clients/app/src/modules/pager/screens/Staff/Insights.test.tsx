import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));
vi.mock('@tanstack/react-router', () => ({
  Link: ({ to, children, ...rest }: { to: string; children: React.ReactNode }) => (
    <a href={to} {...rest}>
      {children}
    </a>
  ),
}));

const api = vi.hoisted(() => ({ getStats: vi.fn() }));
vi.mock('../../shared/lib/clients', () => ({ staffPagerClient: api }));

import Insights from './Insights';
import { initTestI18n } from '../../shared/testing/i18n';

const byHour = Array.from({ length: 24 }, (_, hour) => (hour === 13 ? 9 : hour === 9 ? 3 : 0));

function renderInsights() {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  return render(
    <QueryClientProvider client={client}>
      <Insights />
    </QueryClientProvider>
  );
}

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  api.getStats.mockReset().mockResolvedValue({
    days: 7,
    timezone: 'Asia/Beirut',
    orders: 12,
    readyOrders: 10,
    avgWaitSeconds: 270,
    medianWaitSeconds: 240,
    longestWaitSeconds: 45,
    byHour,
    byDay: [
      { date: '2026-11-21', orders: 5 },
      { date: '2026-11-22', orders: 7 },
    ],
  });
});

describe('pager insights', () => {
  it('shows the orders, the waits and the busiest hour of the last week', async () => {
    renderInsights();

    expect(await screen.findByText('12')).toBeInTheDocument();
    expect(api.getStats).toHaveBeenCalledWith({ days: 7 });
    expect(screen.getByText('4.5 min')).toBeInTheDocument();
    expect(screen.getByText('4 min')).toBeInTheDocument();
    // Under a minute reads in seconds.
    expect(screen.getByText('45 s')).toBeInTheDocument();
    expect(
      screen.getByRole('img', { name: 'Busiest at 13:00, with 9 orders.' })
    ).toBeInTheDocument();
    expect(screen.getByText('Hours and days are in Asia/Beirut.')).toBeInTheDocument();
  });

  it('switches the period', async () => {
    renderInsights();
    await screen.findByText('12');

    fireEvent.click(screen.getByRole('button', { name: 'Last 30 days' }));
    await waitFor(() => expect(api.getStats).toHaveBeenLastCalledWith({ days: 30 }));
    expect(screen.getByRole('button', { name: 'Last 30 days' })).toHaveAttribute(
      'aria-pressed',
      'true'
    );
  });

  it('says so when there were no orders', async () => {
    api.getStats.mockResolvedValue({ days: 7, orders: 0, byHour: [], byDay: [] });
    renderInsights();

    expect(await screen.findByText('No orders in this period yet.')).toBeInTheDocument();
  });
});
