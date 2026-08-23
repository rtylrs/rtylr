import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

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

describe('OrderStatus not-found state', () => {
  beforeEach(() => {
    navigateSpy.mockClear();
    apiGet.mockReset();
    searchState.value = '?businessId=1&public_token=tok-abc';
  });

  it('shows the shared EmptyState and navigates back to the menu on the CTA', async () => {
    apiGet.mockRejectedValue({ response: { status: 404 } });

    render(<OrderStatus />);

    await waitFor(() => expect(screen.getByText('notFoundTitle')).toBeInTheDocument());
    expect(screen.getByText('notFoundDescription')).toBeInTheDocument();

    // The always-on "Order Another" CTA should not double up with the
    // EmptyState's own action once the order is confirmed not found.
    expect(screen.queryByText('Order Another')).not.toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'backToMenu' }));

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

    await waitFor(() => expect(screen.getByText('notFoundTitle')).toBeInTheDocument());
    expect(apiGet).not.toHaveBeenCalled();
  });
});
