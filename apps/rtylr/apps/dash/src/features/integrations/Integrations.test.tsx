import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

type Integration = { id: string; name: string; category: string; status: string };

const integrationsData = vi.hoisted(() => ({
  data: [] as Integration[],
  isLoading: false,
}));

vi.mock('../../shared/hooks/queries/useIntegrations', () => ({
  useIntegrations: () => integrationsData,
  useConnectIntegration: () => ({ mutate: vi.fn(), isPending: false }),
  useDisconnectIntegration: () => ({ mutate: vi.fn(), isPending: false }),
}));

import Integrations from './Integrations';

beforeEach(() => {
  vi.clearAllMocks();
  integrationsData.data = [];
  integrationsData.isLoading = false;
});

describe('Integrations: filter-empty state', () => {
  it('shows the illustration, title and a clear-filters call-to-action when no integrations match', () => {
    integrationsData.data = [
      { id: '1', name: 'Stripe', category: 'payments', status: 'available' },
    ];
    render(<Integrations />);

    fireEvent.change(screen.getByPlaceholderText('Search integrations...'), {
      target: { value: 'no such integration' },
    });

    expect(screen.getByText('integrations.emptyMatchTitle')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'common.clearFilters' })).toBeInTheDocument();
  });

  it('clears the search and category filters when the call-to-action is clicked', () => {
    integrationsData.data = [
      { id: '1', name: 'Stripe', category: 'payments', status: 'available' },
    ];
    render(<Integrations />);

    fireEvent.click(screen.getByRole('button', { name: 'Messaging' }));
    fireEvent.change(screen.getByPlaceholderText('Search integrations...'), {
      target: { value: 'Stripe' },
    });
    expect(screen.getByText('integrations.emptyMatchTitle')).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'common.clearFilters' }));

    expect(screen.queryByText('integrations.emptyMatchTitle')).not.toBeInTheDocument();
    expect(screen.getByText('Stripe')).toBeInTheDocument();
  });

  it('does not show the empty state when an integration matches', () => {
    integrationsData.data = [
      { id: '1', name: 'Stripe', category: 'payments', status: 'available' },
    ];
    render(<Integrations />);

    expect(screen.queryByText('integrations.emptyMatchTitle')).not.toBeInTheDocument();
  });
});
