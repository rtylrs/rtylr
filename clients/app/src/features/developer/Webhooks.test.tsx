import path from 'node:path';
import i18next from 'i18next';
import { initTestI18nFrom } from '../../../test/i18nTestInstance';
import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';
import type { Webhook } from '../../shared/hooks/queries/useWebhooks';

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const createMutateAsync = vi.hoisted(() => vi.fn().mockResolvedValue({ id: '1' }));
const webhooksData = vi.hoisted(() => ({
  data: [] as Webhook[],
  isLoading: false,
}));

vi.mock('../../shared/hooks/queries/useWebhooks', () => ({
  useWebhooks: () => webhooksData,
  useCreateWebhook: () => ({ mutateAsync: createMutateAsync, isPending: false }),
  useDeleteWebhook: () => ({ mutateAsync: vi.fn(), isPending: false }),
}));

import Webhooks from './Webhooks';

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));

beforeEach(() => {
  vi.clearAllMocks();
  createMutateAsync.mockResolvedValue({ id: '1' });
  webhooksData.data = [];
  webhooksData.isLoading = false;
});

describe('Webhooks empty state', () => {
  it('shows the illustration, title and a create call-to-action when there are no webhooks', () => {
    render(<Webhooks />);

    expect(screen.getByText(i18next.t('developer.webhooks.emptyTitle'))).toBeInTheDocument();
    expect(screen.getByText(i18next.t('developer.webhooks.emptyDescription'))).toBeInTheDocument();
    expect(
      screen.getAllByRole('button', { name: i18next.t('developer.webhooks.newWebhook') })[1]
    ).toBeInTheDocument();
  });

  it('opens the create-webhook dialog when the call-to-action is clicked', () => {
    render(<Webhooks />);

    expect(screen.queryByRole('heading', { name: 'New webhook' })).not.toBeInTheDocument();

    fireEvent.click(
      screen.getAllByRole('button', { name: i18next.t('developer.webhooks.newWebhook') })[1]
    );

    expect(screen.getByRole('heading', { name: 'New webhook' })).toBeInTheDocument();
  });

  it('does not show the empty state when there are webhooks', () => {
    webhooksData.data = [
      {
        id: '1',
        url: 'https://api.example.com/hooks',
        isActive: true,
        events: ['order_created'],
        lastTriggeredAt: '2026-09-23T11:00:00Z',
      },
    ];
    render(<Webhooks />);

    expect(screen.queryByText(i18next.t('developer.webhooks.emptyTitle'))).not.toBeInTheDocument();
  });

  it('rejects an invalid destination before calling create', () => {
    render(<Webhooks />);
    fireEvent.click(screen.getAllByRole('button', { name: 'New webhook' })[0]);
    fireEvent.change(screen.getByLabelText('Endpoint URL'), { target: { value: 'not-a-url' } });
    fireEvent.click(screen.getByRole('checkbox'));
    fireEvent.click(screen.getByRole('button', { name: 'Create webhook' }));

    expect(createMutateAsync).not.toHaveBeenCalled();
  });

  it('sends a trimmed HTTP(S) destination', async () => {
    render(<Webhooks />);
    fireEvent.click(screen.getAllByRole('button', { name: 'New webhook' })[0]);
    fireEvent.change(screen.getByLabelText('Endpoint URL'), {
      target: { value: '  https://api.example.com/hooks  ' },
    });
    fireEvent.click(screen.getByRole('checkbox'));
    fireEvent.click(screen.getByRole('button', { name: 'Create webhook' }));

    await waitFor(() =>
      expect(createMutateAsync).toHaveBeenCalledWith({
        url: 'https://api.example.com/hooks',
        events: ['order_created'],
      })
    );
  });
});
