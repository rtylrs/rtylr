import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const apiKeysData = vi.hoisted(() => ({
  data: [] as { id: string; name: string }[],
  isLoading: false,
}));

const createMutate = vi.hoisted(() => vi.fn());

vi.mock('../../shared/hooks/queries/useApiKeys', () => ({
  useApiKeys: () => apiKeysData,
  useCreateApiKey: () => ({ mutateAsync: createMutate, isPending: false }),
  useRevokeApiKey: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

import ApiKeys from './ApiKeys';

beforeEach(() => {
  vi.clearAllMocks();
  apiKeysData.data = [];
  apiKeysData.isLoading = false;
});

describe('API keys empty state', () => {
  it('shows the illustration, title and a create call-to-action when there are no keys', () => {
    render(<ApiKeys />);

    expect(screen.getByText('developer.apiKeys.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('developer.apiKeys.emptyDescription')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'developer.apiKeys.newKey' })).toBeInTheDocument();
  });

  it('opens the create-key dialog when the call-to-action is clicked', () => {
    render(<ApiKeys />);

    expect(screen.queryByText('Create API key')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'developer.apiKeys.newKey' }));

    expect(screen.getByText('Create API key')).toBeInTheDocument();
  });

  it('does not show the empty state when there are keys', () => {
    apiKeysData.data = [
      {
        id: '1',
        name: 'Production',
        keyPrefix: 'rk_live_',
        scopes: ['upload'],
        createdAt: new Date().toISOString(),
      } as unknown as { id: string; name: string },
    ];
    render(<ApiKeys />);

    expect(screen.queryByText('developer.apiKeys.emptyTitle')).not.toBeInTheDocument();
  });
});
