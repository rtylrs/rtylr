import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';

// i18next is deliberately not mocked. An identity `t` makes an assertion and
// the component agree on a key, so neither notices the key never resolves,
// which is how this suite passed while every string on screen was a dotted
// path (rtylr#1269). The real locale files mean the copy below also proves the
// lookup.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
}));

const navigateSpy = vi.hoisted(() => vi.fn());
// The search string is per-test now: whether a public_token is present decides
// whether this screen can poll at all.
const searchState = vi.hoisted(() => ({ value: '?businessId=1&public_token=tok-abc' }));
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ id: '999' }),
  useRouterState: () => searchState.value,
  useNavigate: () => navigateSpy,
}));

const apiGet = vi.hoisted(() => vi.fn());
vi.mock('../../shared/lib/clients', () => ({
  orderClient: { getPublicOrderStatus: apiGet },
}));

import OrderStatus from './OrderStatus';
import { initTestI18n } from '../../shared/testing/i18n';

beforeAll(async () => {
  await initTestI18n('en');
});

describe('OrderStatus not-found state', () => {
  beforeEach(() => {
    navigateSpy.mockClear();
    apiGet.mockReset();
    searchState.value = '?businessId=1&public_token=tok-abc';
  });

  it('shows the shared EmptyState and navigates back to the menu on the CTA', async () => {
    apiGet.mockRejectedValue({ response: { status: 404 } });

    render(<OrderStatus />);

    await waitFor(() => expect(screen.getByText('Order not found')).toBeInTheDocument());
    expect(
      screen.getByText("We couldn't find an order matching this ID. Please check your order link.")
    ).toBeInTheDocument();

    // The always-on "Order Another" CTA should not double up with the
    // EmptyState's own action once the order is confirmed not found.
    expect(screen.queryByText('Order Another')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Back to menu' }));

    expect(navigateSpy).toHaveBeenCalledWith({ to: '/' });
  });

  it('sends the public token from the URL with the status request', async () => {
    apiGet.mockResolvedValue({ status: 'preparing' });

    render(<OrderStatus />);

    await waitFor(() => expect(apiGet).toHaveBeenCalled());
    expect(apiGet).toHaveBeenCalledWith({
      orderId: '999',
      businessId: '1',
      publicToken: 'tok-abc',
    });
  });

  // A link made before the token existed, or one hand-edited to drop it, cannot
  // succeed: the endpoint rejects an empty token as malformed rather than as a
  // missing order. Polling it would just produce errors, so the screen shows
  // the same dead end without issuing the request.
  it('shows not-found without calling the API when the URL carries no token', async () => {
    searchState.value = '?businessId=1';
    apiGet.mockResolvedValue({ status: 'preparing' });

    render(<OrderStatus />);

    await waitFor(() => expect(screen.getByText('Order not found')).toBeInTheDocument());
    expect(apiGet).not.toHaveBeenCalled();
  });
});
