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

// The four step labels and the four progress messages are looked up with a
// template key built from the step name, so `i18n-key-resolution.test.ts`,
// which reads literal keys only, cannot see them. Each family is exercised
// here instead, through the backend status that selects it.
describe('OrderStatus progress copy', () => {
  beforeEach(() => {
    navigateSpy.mockClear();
    apiGet.mockReset();
    searchState.value = '?businessId=1&public_token=tok-abc';
  });

  it.each([
    ['pending', 'Order received!'],
    ['confirmed', 'Order received!'],
    ['preparing', 'Being prepared...'],
    ['held', 'Being prepared...'],
    ['ready', 'Ready for pickup!'],
    ['served', 'Order complete!'],
    ['paid', 'Order complete!'],
  ])('renders the message for %s', async (backendStatus, message) => {
    apiGet.mockResolvedValue({ status: backendStatus });

    render(<OrderStatus />);

    await waitFor(() => expect(screen.getByText(message)).toBeInTheDocument());
    for (const label of ['Received', 'Preparing', 'Ready', 'Done']) {
      expect(screen.getByText(label)).toBeInTheDocument();
    }
  });

  it('falls back to the raw status for one the display steps do not cover', async () => {
    apiGet.mockResolvedValue({ status: 'awaiting_courier' });

    render(<OrderStatus />);

    await waitFor(() => expect(screen.getByText('Status: awaiting_courier')).toBeInTheDocument());
  });

  it.each([
    ['cancelled', 'Order cancelled'],
    ['voided', 'Order cancelled'],
    ['refunded', 'Order refunded'],
  ])('renders the terminal copy for %s', async (backendStatus, message) => {
    apiGet.mockResolvedValue({ status: backendStatus });

    render(<OrderStatus />);

    await waitFor(() => expect(screen.getByText(message)).toBeInTheDocument());
  });

  // Pins the number the notice shows. It does not prove the number came from
  // POLL_INTERVAL_MS, but changing that constant without the copy now fails
  // here, which is the drift worth catching.
  it('states the poll interval', () => {
    apiGet.mockResolvedValue({ status: 'preparing' });

    render(<OrderStatus />);

    expect(
      screen.getByText('Status refreshes automatically every 30 seconds.')
    ).toBeInTheDocument();
  });

  it('offers a translated retry when the status cannot be read', async () => {
    apiGet.mockRejectedValue({ response: { status: 500 } });

    render(<OrderStatus />);

    await waitFor(() =>
      expect(screen.getByText('Could not load order status.')).toBeInTheDocument()
    );
    expect(screen.getByRole('button', { name: 'Retry' })).toBeInTheDocument();
  });
});
