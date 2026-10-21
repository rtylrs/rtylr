import React from 'react';
import { fireEvent, render, screen } from '@testing-library/react';
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

  it('offers rotation only when a handler is given and reports the webhook id', () => {
    const webhook: WebhookRowProps['webhook'] = {
      id: '9',
      url: 'https://example.com/webhooks',
      events: ['order_created'],
      isActive: true,
    };
    const onRotateSecret = vi.fn();

    const { rerender } = render(<WebhookRow webhook={webhook} />);
    expect(screen.queryByRole('button', { name: 'common.webhook.rotate' })).not.toBeInTheDocument();

    rerender(<WebhookRow webhook={webhook} onRotateSecret={onRotateSecret} />);
    fireEvent.click(screen.getByRole('button', { name: 'common.webhook.rotate' }));
    expect(onRotateSecret).toHaveBeenCalledWith('9');
  });
});
