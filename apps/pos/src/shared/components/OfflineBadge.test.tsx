import { render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import OfflineBadge from './OfflineBadge';
import { useOfflineQueueStore } from '../stores/offlineQueueStore';
import { useSessionCloseQueueStore } from '../stores/sessionCloseQueueStore';
import { useCashMovementQueueStore } from '../stores/cashMovementQueueStore';

describe('OfflineBadge', () => {
  beforeEach(() => {
    useOfflineQueueStore.setState({ queue: [], online: true });
    useSessionCloseQueueStore.setState({ queue: [] });
    useCashMovementQueueStore.setState({ queue: [] });
  });

  it('counts pending orders, session closes, and cash movements together', () => {
    useOfflineQueueStore.setState({
      online: true,
      queue: [
        {
          id: 'oq_1',
          order: {
            id: 'order-1',
            status: 'paid',
            items: [],
            subtotal: 10,
            tax: 0,
            tip: 0,
            discount: 0,
            total: 10,
            createdAt: '2026-05-06T00:00:00.000Z',
          },
          queuedAt: '2026-05-06T00:00:00.000Z',
          retries: 0,
          status: 'pending',
        },
      ],
    });
    useSessionCloseQueueStore.getState().enqueue({
      sessionId: 'current',
      expected: 100,
      actual: 95,
      variance: -5,
    });
    useCashMovementQueueStore.getState().enqueue({
      action: 'drop',
      payload: { amount: 25, reason: 'Safe deposit' },
    });

    render(<OfflineBadge />);

    expect(screen.getByText('3 pending sync')).toBeTruthy();
  });
});
