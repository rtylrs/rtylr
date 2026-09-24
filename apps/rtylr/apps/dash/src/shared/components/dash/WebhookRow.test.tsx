import React from 'react';
import { render, screen } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import WebhookRow, { type WebhookRowProps } from './WebhookRow';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

afterEach(() => {
  vi.useRealTimers();
});

describe('WebhookRow', () => {
  it('renders paused state and the real last-triggered timestamp from the contract', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-09-23T12:00:00Z'));
    const webhook: WebhookRowProps['webhook'] = {
      id: '1',
      url: 'https://example.com/webhooks',
      events: ['order.created'],
      isActive: false,
      lastTriggeredAt: '2026-09-23T11:55:00Z',
    };

    render(<WebhookRow webhook={webhook} />);

    expect(screen.getByText('common.webhook.paused')).toBeInTheDocument();
    expect(screen.queryByText('common.webhook.active')).not.toBeInTheDocument();
    expect(screen.getByText('Last delivery 5m ago')).toBeInTheDocument();
  });
});
