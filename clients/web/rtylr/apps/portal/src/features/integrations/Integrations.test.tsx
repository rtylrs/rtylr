import path from 'node:path';
import i18next from 'i18next';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';
import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

type Integration = {
  id: string;
  provider: string;
  name: string;
  category: string;
  logo: string;
  status: string;
  description: string;
  connectedAt?: string;
  createdAt?: string;
};

const integrationsData = vi.hoisted(() => ({
  data: [] as Integration[],
  isLoading: false,
  isError: false,
  refetch: vi.fn(),
}));
const connectMutate = vi.hoisted(() => vi.fn());
const disconnectMutate = vi.hoisted(() => vi.fn());

vi.mock('../../shared/hooks/queries/useIntegrations', () => ({
  useIntegrations: () => integrationsData,
  useConnectIntegration: () => ({ mutate: connectMutate, isPending: false }),
  useDisconnectIntegration: () => ({ mutate: disconnectMutate, isPending: false }),
}));

import Integrations from './Integrations';

const stripe = (overrides: Partial<Integration> = {}): Integration => ({
  id: '17',
  provider: 'stripe',
  name: 'Stripe',
  category: 'payments',
  logo: 'stripe',
  status: 'connected',
  description: 'Accept online payments with Stripe.',
  ...overrides,
});

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

beforeEach(() => {
  vi.clearAllMocks();
  integrationsData.data = [];
  integrationsData.isLoading = false;
  integrationsData.isError = false;
});

describe('Integrations: contract-backed rows', () => {
  it('shows the illustration, title and a clear-filters call-to-action when no integrations match', () => {
    integrationsData.data = [stripe({ id: '1', status: 'connected' })];
    render(<Integrations />);

    fireEvent.change(
      screen.getByPlaceholderText(i18next.t('common.management.searchIntegrations')),
      {
        target: { value: 'no such integration' },
      }
    );

    expect(screen.getByText(i18next.t('integrations.emptyMatchTitle'))).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: i18next.t('common.clearFilters') })
    ).toBeInTheDocument();
  });

  it('clears the search and category filters when the call-to-action is clicked', () => {
    integrationsData.data = [stripe({ id: '1' })];
    render(<Integrations />);

    fireEvent.click(screen.getByRole('button', { name: 'Messaging' }));
    fireEvent.change(
      screen.getByPlaceholderText(i18next.t('common.management.searchIntegrations')),
      {
        target: { value: 'Stripe' },
      }
    );
    expect(screen.getByText(i18next.t('integrations.emptyMatchTitle'))).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: i18next.t('common.clearFilters') }));

    expect(screen.queryByText(i18next.t('integrations.emptyMatchTitle'))).not.toBeInTheDocument();
    expect(screen.getByText('Stripe')).toBeInTheDocument();
  });

  it('renders a connected provider returned by the current API contract', () => {
    integrationsData.data = [stripe()];
    render(<Integrations />);

    expect(screen.getByText('Stripe')).toBeInTheDocument();
    expect(screen.getByText('Connected')).toBeInTheDocument();
  });

  it('offers the provider catalogue for a new business and connects by provider', () => {
    render(<Integrations />);

    fireEvent.click(screen.getByRole('button', { name: 'Connect' }));

    expect(connectMutate).toHaveBeenCalledWith('stripe');
  });

  it('disconnects a connected row by its numeric id', () => {
    integrationsData.data = [stripe()];
    render(<Integrations />);

    fireEvent.click(screen.getByRole('button', { name: 'Disconnect' }));

    expect(disconnectMutate).toHaveBeenCalledWith('17');
  });

  it('does not show the empty state when an integration matches', () => {
    integrationsData.data = [stripe()];
    render(<Integrations />);

    expect(screen.queryByText(i18next.t('integrations.emptyMatchTitle'))).not.toBeInTheDocument();
  });

  it('shows a retryable error instead of the provider catalogue when loading fails', () => {
    integrationsData.isError = true;
    render(<Integrations />);

    expect(screen.getByText(i18next.t('integrations.loadFailed'))).toBeInTheDocument();
    expect(screen.getByRole('button', { name: i18next.t('common.refresh') })).toBeInTheDocument();
    expect(screen.queryByText('Stripe')).not.toBeInTheDocument();
  });
});
