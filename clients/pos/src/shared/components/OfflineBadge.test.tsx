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
            type: 'dine_in',
            paymentMethod: 'cash',
            items: [],
          },
          queuedAt: '2026-05-06T00:00:00.000Z',
          retries: 0,
          status: 'pending',
        },
      ],
    });
    useSessionCloseQueueStore.getState().enqueue({
      sessionId: 42,
      closingFloat: 95,
      expected: 100,
      variance: -5,
    });
    useCashMovementQueueStore.getState().enqueue({
      action: 'drop',
      payload: { branch_id: 1, session_id: 0, employee_id: 1, amount: 25, reason: 'Safe deposit' },
    });

    render(<OfflineBadge />);

    expect(screen.getByText('3 pending sync')).toBeTruthy();
  });
});
