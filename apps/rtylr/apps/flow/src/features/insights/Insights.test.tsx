// @vitest-environment happy-dom
import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi, type Mock } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

vi.mock('@tanstack/react-router', () => ({ useNavigate: () => vi.fn() }));

vi.mock('@voxire/api', () => ({ unwrap: (value: unknown) => value }));

vi.mock('motion/react', () => ({
  motion: {
    div: ({ children, ...rest }: { children: React.ReactNode }) =>
      React.createElement('div', rest, children),
  },
}));

vi.mock('../../shared/services/flowApi', () => ({
  flowApi: { getInsights: vi.fn() },
}));

import Insights from './Insights';
import { flowApi } from '../../shared/services/flowApi';

describe('Insights tolerates an overview response with no series', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the burndown empty state instead of throwing when the series are absent', async () => {
    // The registered v2 overview handler returns an empty response and the
    // generated decoder leaves velocity/burndown/throughput/timeByMember
    // undefined. `data.velocity.reduce(...)` then threw and the route error
    // boundary replaced the whole screen.
    (flowApi.getInsights as unknown as Mock).mockResolvedValue({});

    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    render(
      <QueryClientProvider client={client}>
        <Insights />
      </QueryClientProvider>
    );

    await waitFor(() => expect(screen.getByText('common.insights.noData')).toBeInTheDocument());
    expect(screen.getByText('common.insights.heading')).toBeInTheDocument();
  });
});
