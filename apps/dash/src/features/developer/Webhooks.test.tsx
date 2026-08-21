import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const webhooksData = vi.hoisted(() => ({
  data: [] as { id: string; url: string }[],
  isLoading: false,
}));

vi.mock('../../shared/hooks/queries/useWebhooks', () => ({
  useWebhooks: () => webhooksData,
  useCreateWebhook: () => ({ mutateAsync: vi.fn(), isPending: false }),
  useDeleteWebhook: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

import Webhooks from './Webhooks';

beforeEach(() => {
  vi.clearAllMocks();
  webhooksData.data = [];
  webhooksData.isLoading = false;
});

describe('Webhooks empty state', () => {
  it('shows the illustration, title and a create call-to-action when there are no webhooks', () => {
    render(<Webhooks />);

    expect(screen.getByText('developer.webhooks.emptyTitle')).toBeInTheDocument();
    expect(screen.getByText('developer.webhooks.emptyDescription')).toBeInTheDocument();
    expect(
      screen.getByRole('button', { name: 'developer.webhooks.newWebhook' })
    ).toBeInTheDocument();
  });

  it('opens the create-webhook dialog when the call-to-action is clicked', () => {
    render(<Webhooks />);

    expect(screen.queryByRole('heading', { name: 'New webhook' })).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'developer.webhooks.newWebhook' }));

    expect(screen.getByRole('heading', { name: 'New webhook' })).toBeInTheDocument();
  });

  it('does not show the empty state when there are webhooks', () => {
    webhooksData.data = [
      {
        id: '1',
        url: 'https://api.example.com/hooks',
        status: 'active',
        events: ['order.created'],
        secretPreview: 'whsec_***',
      } as unknown as { id: string; url: string },
    ];
    render(<Webhooks />);

    expect(screen.queryByText('developer.webhooks.emptyTitle')).not.toBeInTheDocument();
  });
});
