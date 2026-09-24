import React from 'react';
import { act, renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { useConnectIntegration, useIntegrations } from './useIntegrations';

const mockIntegrationClient = vi.hoisted(() => ({
  listIntegrations: vi.fn(),
  connectIntegration: vi.fn(),
  disconnectIntegration: vi.fn(),
}));

vi.mock('../../lib/clients', () => ({ integrationClient: mockIntegrationClient }));

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

function createWrapper() {
  const client = new QueryClient({
    defaultOptions: { queries: { retry: false }, mutations: { retry: false } },
  });
  return ({ children }: { children: React.ReactNode }) =>
    React.createElement(QueryClientProvider, { client }, children);
}

describe('useIntegrations', () => {
  beforeEach(() => vi.clearAllMocks());

  it('maps the current provider-based response into the UI model', async () => {
    mockIntegrationClient.listIntegrations.mockResolvedValue({
      integrations: [
        {
          id: '17',
          provider: 'stripe',
          status: 'connected',
          config: '{}',
          connectedAt: '2026-09-01T12:00:00Z',
          createdAt: '2026-08-01T12:00:00Z',
        },
      ],
    });

    const { result } = renderHook(() => useIntegrations(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));

    expect(result.current.data?.[0]).toMatchObject({
      id: '17',
      provider: 'stripe',
      name: 'Stripe',
      category: 'payments',
      status: 'connected',
    });
  });

  it('sends the provider when connecting, not the row id', async () => {
    mockIntegrationClient.connectIntegration.mockResolvedValue({ id: '18' });

    const { result } = renderHook(() => useConnectIntegration(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.mutateAsync('stripe');
    });

    expect(mockIntegrationClient.connectIntegration).toHaveBeenCalledWith({ provider: 'stripe' });
  });
});
