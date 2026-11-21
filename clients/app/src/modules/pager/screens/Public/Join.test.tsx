import React from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

// Real translations: an assertion written in English also proves the key resolves.

vi.mock('@voxire/components', () => ({
  HelmetContainer: () => null,
  FullScreenLoader: () => <p>Loading</p>,
}));

const navigateSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ slug: 'Acme' }),
  useNavigate: () => navigateSpy,
}));

const api = vi.hoisted(() => ({ getPublicVenue: vi.fn(), joinTicket: vi.fn() }));
vi.mock('../../shared/lib/clients', () => ({ publicPagerClient: api }));

const unlockSpy = vi.hoisted(() => vi.fn());
vi.mock('../../../../shared/alarm/readyAlarm', () => ({ readyAlarm: { unlock: unlockSpy } }));

vi.mock('../../../../shared/services/i18next', () => ({ selectLanguage: vi.fn() }));

// The server's push key as the page has loaded it; the asking itself is real.
const pushKey = vi.hoisted(() => ({ value: '' }));
vi.mock('../../shared/lib/pushAlerts', async (importOriginal) => ({
  ...(await importOriginal<typeof import('../../shared/lib/pushAlerts')>()),
  usePushKey: () => pushKey.value,
}));

import Join from './Join';
import { JOIN_RETRY_INTERVAL_MS, JOIN_RETRY_WINDOW_MS } from '../../shared/lib/joinWithRetry';
import { installFakePushBrowser, type FakePushBrowser } from '../../shared/testing/fakePushBrowser';
import { initTestI18n } from '../../shared/testing/i18n';

const notYet = { statusCode: 404, code: 'ticket_not_found' };

function renderJoin() {
  const client = new QueryClient({ defaultOptions: { queries: { retryDelay: 0 } } });
  return render(
    <QueryClientProvider client={client}>
      <Join />
    </QueryClientProvider>
  );
}

async function typeNumber(value: string): Promise<void> {
  fireEvent.change(await screen.findByLabelText('Order number'), { target: { value } });
}

const advance = (ms: number): Promise<void> =>
  act(async () => {
    await vi.advanceTimersByTimeAsync(ms);
  });

beforeAll(async () => {
  await initTestI18n('en');
});

let pushBrowser: FakePushBrowser;

beforeEach(() => {
  pushBrowser = installFakePushBrowser();
  pushKey.value = '';
  // Real time still moves the fake clock, so findBy* keeps working while the
  // five-second retry and two-minute window are stepped by hand.
  vi.useFakeTimers({ shouldAdvanceTime: true });
  navigateSpy.mockReset();
  unlockSpy.mockReset();
  api.getPublicVenue.mockReset().mockResolvedValue({ name: 'Acme Cafe', enabled: true });
  api.joinTicket.mockReset();
});

afterEach(() => {
  vi.useRealTimers();
  pushBrowser.restore();
});

describe('pager join page', () => {
  it('unlocks sound in the tap, before the request, then opens the ticket in the same document', async () => {
    api.joinTicket.mockResolvedValue({ publicToken: 'tok-1', number: '42', status: 'waiting' });
    renderJoin();

    expect(await screen.findByText('Acme Cafe')).toBeInTheDocument();
    expect(api.getPublicVenue).toHaveBeenCalledWith({ slug: 'acme' });
    await typeNumber('42');
    fireEvent.click(screen.getByRole('button', { name: "Ring me when it's ready" }));

    await waitFor(() => expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager/t/tok-1' }));
    expect(api.joinTicket).toHaveBeenCalledWith(
      { slug: 'acme', number: '42' },
      expect.objectContaining({ signal: expect.any(AbortSignal) })
    );
    expect(unlockSpy).toHaveBeenCalledTimes(1);
    expect(unlockSpy.mock.invocationCallOrder[0]).toBeLessThan(
      api.joinTicket.mock.invocationCallOrder[0]
    );
  });

  it('asks to notify on the lock screen in the same tap, before the request', async () => {
    pushKey.value = 'server-key';
    api.joinTicket.mockResolvedValue({ publicToken: 'tok-1', number: '42', status: 'waiting' });
    renderJoin();

    await typeNumber('42');
    fireEvent.click(screen.getByRole('button', { name: "Ring me when it's ready" }));

    await waitFor(() => expect(navigateSpy).toHaveBeenCalled());
    expect(pushBrowser.requestPermission).toHaveBeenCalledTimes(1);
    expect(pushBrowser.requestPermission.mock.invocationCallOrder[0]).toBeLessThan(
      api.joinTicket.mock.invocationCallOrder[0]
    );
  });

  it('does not ask to notify while the server has push off', async () => {
    api.joinTicket.mockResolvedValue({ publicToken: 'tok-1', number: '42', status: 'waiting' });
    renderJoin();

    await typeNumber('42');
    fireEvent.click(screen.getByRole('button', { name: "Ring me when it's ready" }));

    await waitFor(() => expect(navigateSpy).toHaveBeenCalled());
    expect(unlockSpy).toHaveBeenCalledTimes(1);
    expect(pushBrowser.requestPermission).not.toHaveBeenCalled();
  });

  it('offers a numeric keypad and cannot be sent empty', async () => {
    renderJoin();

    const input = await screen.findByLabelText('Order number');
    expect(input).toHaveAttribute('inputmode', 'numeric');
    expect(screen.getByRole('button', { name: "Ring me when it's ready" })).toBeDisabled();
  });

  it('sends Arabic-Indic digits as ASCII', async () => {
    api.joinTicket.mockResolvedValue({ publicToken: 'tok-1' });
    renderJoin();

    await typeNumber('٤٢');
    expect(screen.getByLabelText('Order number')).toHaveValue('42');
    fireEvent.click(screen.getByRole('button', { name: "Ring me when it's ready" }));

    await waitFor(() => expect(api.joinTicket).toHaveBeenCalled());
    expect(api.joinTicket.mock.calls[0][0]).toEqual({ slug: 'acme', number: '42' });
  });

  it('waits for the cashier at a strict venue and joins once the number exists', async () => {
    api.joinTicket
      .mockRejectedValueOnce(notYet)
      .mockResolvedValue({ publicToken: 'tok-9', number: '42', status: 'waiting' });
    renderJoin();
    await typeNumber('42');
    fireEvent.click(screen.getByRole('button', { name: "Ring me when it's ready" }));

    expect(
      await screen.findByText('Waiting for the cashier to enter your number')
    ).toBeInTheDocument();
    expect(
      screen.getByText('Order 42 is not on the list yet. Stay on this page: we keep checking.')
    ).toBeInTheDocument();
    expect(navigateSpy).not.toHaveBeenCalled();

    await advance(JOIN_RETRY_INTERVAL_MS);

    expect(api.joinTicket).toHaveBeenCalledTimes(2);
    expect(navigateSpy).toHaveBeenCalledWith({ to: '/pager/t/tok-9' });
    // The retries ride on the first tap; nothing asked for a second one.
    expect(unlockSpy).toHaveBeenCalledTimes(1);
  });

  it('stops after two minutes and offers a retry that unlocks sound again', async () => {
    api.joinTicket.mockRejectedValue(notYet);
    renderJoin();
    await typeNumber('42');
    fireEvent.click(screen.getByRole('button', { name: "Ring me when it's ready" }));
    await screen.findByText('Waiting for the cashier to enter your number');

    await advance(JOIN_RETRY_WINDOW_MS);

    expect(await screen.findByText("We still can't find order 42")).toBeInTheDocument();
    const attempts = api.joinTicket.mock.calls.length;
    await advance(JOIN_RETRY_INTERVAL_MS * 2);
    expect(api.joinTicket).toHaveBeenCalledTimes(attempts);

    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

    expect(unlockSpy).toHaveBeenCalledTimes(2);
    expect(api.joinTicket).toHaveBeenCalledTimes(attempts + 1);
    expect(
      await screen.findByText('Waiting for the cashier to enter your number')
    ).toBeInTheDocument();
  });

  it('lets the customer change the number while waiting, and stops asking', async () => {
    api.joinTicket.mockRejectedValue(notYet);
    renderJoin();
    await typeNumber('42');
    fireEvent.click(screen.getByRole('button', { name: "Ring me when it's ready" }));
    await screen.findByText('Waiting for the cashier to enter your number');

    fireEvent.click(screen.getByRole('button', { name: 'Change the number' }));
    await advance(JOIN_RETRY_INTERVAL_MS * 3);

    expect(screen.getByLabelText('Order number')).toHaveValue('42');
    expect(api.joinTicket).toHaveBeenCalledTimes(1);
  });

  it.each([
    [
      { statusCode: 400, code: 'invalid_number' },
      'Use the letters and digits printed on your receipt.',
    ],
    [{ statusCode: 429 }, 'Too many attempts. Wait a moment, then try again.'],
    [
      { statusCode: 429, code: 'too_many_open_tickets' },
      'This venue has too many open orders right now. Please ask at the counter.',
    ],
    [{ statusCode: 500 }, "We couldn't reach the pager. Check your connection and try again."],
  ])('explains a failed join: %j', async (failure, message) => {
    api.joinTicket.mockRejectedValue(failure);
    renderJoin();
    await typeNumber('42');
    fireEvent.click(screen.getByRole('button', { name: "Ring me when it's ready" }));

    expect(await screen.findByText(message)).toBeInTheDocument();
    expect(api.joinTicket).toHaveBeenCalledTimes(1);
    expect(navigateSpy).not.toHaveBeenCalled();
  });

  it('says so when the venue has no pager, and offers no form', async () => {
    api.getPublicVenue.mockRejectedValue({ statusCode: 404, code: 'venue_not_found' });
    renderJoin();

    expect(await screen.findByText("This pager isn't available")).toBeInTheDocument();
    expect(screen.queryByLabelText('Order number')).not.toBeInTheDocument();
    expect(api.getPublicVenue).toHaveBeenCalledTimes(1);
  });

  it('offers a retry when the venue cannot be loaded', async () => {
    api.getPublicVenue.mockRejectedValue({ statusCode: 500 });
    renderJoin();

    expect(await screen.findByText("We couldn't load this page.")).toBeInTheDocument();
    api.getPublicVenue.mockResolvedValue({ name: 'Acme Cafe', enabled: true });
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByLabelText('Order number')).toBeInTheDocument();
  });
});
