import React from 'react';
import { render, screen, fireEvent, waitFor, act } from '@testing-library/react';
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

import OrderStatus, { isSameOriginReturnUrl } from './OrderStatus';
import { useCartStore } from '../../shared/stores/cartStore';
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

    // No slug in the URL (a POS-menu order): back to that business's menu, not
    // the Portal root, which sits behind the sign-in gate.
    expect(navigateSpy).toHaveBeenCalledWith({ to: '/b/1/menu' });
  });

  it('falls back to the root only when neither a slug nor a business id is known', async () => {
    searchState.value = '?public_token=tok-abc';
    apiGet.mockResolvedValue({ status: 'preparing' });

    render(<OrderStatus />);

    fireEvent.click(screen.getByRole('button', { name: /Order Another|Back to menu/ }));

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
    ['processing', 'Being prepared...'],
    ['ready', 'Ready for pickup!'],
    ['shipped', 'Ready for pickup!'],
    ['delivered', 'Order complete!'],
    ['completed', 'Order complete!'],
    ['done', 'Order complete!'],
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

describe('OrderStatus return-to-menu', () => {
  beforeEach(() => {
    navigateSpy.mockClear();
    apiGet.mockReset();
    useCartStore.getState().clearCart();
    searchState.value = '?businessId=1&public_token=tok-abc';
  });

  it('shows the order number carried in the query', async () => {
    searchState.value = '?businessId=1&public_token=tok-abc&number=ORD-9001';
    apiGet.mockResolvedValue({ status: 'preparing' });

    render(<OrderStatus />);

    expect(screen.getByText('ORD-9001')).toBeInTheDocument();
  });

  it('clears the cart when ordering another', async () => {
    searchState.value = '?businessId=1&public_token=tok-abc&slug=acme';
    apiGet.mockResolvedValue({ status: 'preparing' });
    useCartStore.getState().setBusinessId(1);
    useCartStore.getState().addItem({ menuItemId: 7, name: 'Falafel wrap', unitPrice: 5.5 });

    render(<OrderStatus />);
    await waitFor(() => expect(apiGet).toHaveBeenCalled());

    fireEvent.click(screen.getByRole('button', { name: 'Order Another' }));

    expect(useCartStore.getState().items).toHaveLength(0);
  });

  // return_url is attacker-controllable, so a link to another origin must not
  // be followed. Before the fix this set window.location.href and skipped the
  // in-app fallback entirely, which is the open redirect this pins.
  it('ignores a return_url on another origin and uses the slug instead', async () => {
    searchState.value =
      '?businessId=1&public_token=tok-abc&slug=acme' +
      '&return_url=https%3A%2F%2Fevil.example%2Fsteal';
    apiGet.mockResolvedValue({ status: 'preparing' });

    render(<OrderStatus />);
    await waitFor(() => expect(apiGet).toHaveBeenCalled());

    fireEvent.click(screen.getByRole('button', { name: 'Order Another' }));

    expect(navigateSpy).toHaveBeenCalledWith({ to: '/acme/menu' });
  });
});

describe('isSameOriginReturnUrl', () => {
  it.each([
    ['/acme/menu', true],
    ['/menu/checkout?slug=acme', true],
    ['https://evil.example/steal', false],
    ['//evil.example/steal', false],
    ['javascript:alert(1)', false],
  ])('treats %s as same-origin=%s', (value, expected) => {
    expect(isSameOriginReturnUrl(value)).toBe(expected);
  });

  it('accepts an absolute URL on the current origin', () => {
    expect(isSameOriginReturnUrl(`${window.location.origin}/acme/menu`)).toBe(true);
  });
});

describe('OrderStatus polling', () => {
  beforeEach(() => {
    navigateSpy.mockClear();
    apiGet.mockReset();
    searchState.value = '?businessId=1&public_token=tok-abc';
  });

  it('refreshes the status when the poll interval fires', async () => {
    apiGet.mockResolvedValueOnce({ status: 'pending' }).mockResolvedValue({ status: 'preparing' });
    const callbacks: (() => void)[] = [];
    const setIntervalSpy = vi.spyOn(globalThis, 'setInterval').mockImplementation((cb) => {
      callbacks.push(cb as () => void);
      return 0 as unknown as ReturnType<typeof setInterval>;
    });

    try {
      render(<OrderStatus />);
      await waitFor(() => expect(screen.getByText('Order received!')).toBeInTheDocument());

      await act(async () => {
        callbacks.forEach((cb) => cb());
      });

      await waitFor(() => expect(screen.getByText('Being prepared...')).toBeInTheDocument());
      expect(apiGet).toHaveBeenCalledTimes(2);
    } finally {
      setIntervalSpy.mockRestore();
    }
  });

  it('clears the polling interval on unmount', () => {
    apiGet.mockResolvedValue({ status: 'preparing' });
    const clearIntervalSpy = vi.spyOn(globalThis, 'clearInterval');

    const { unmount } = render(<OrderStatus />);
    unmount();

    expect(clearIntervalSpy).toHaveBeenCalled();
    clearIntervalSpy.mockRestore();
  });
});
