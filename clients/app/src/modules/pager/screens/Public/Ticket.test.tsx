import React from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

const params = vi.hoisted(() => ({ token: 'tok-1' as string | undefined }));
vi.mock('@tanstack/react-router', () => ({ useParams: () => params }));

const api = vi.hoisted(() => ({ getPublicTicket: vi.fn(), acknowledgeTicket: vi.fn() }));
vi.mock('../../shared/lib/clients', () => ({ publicPagerClient: api }));

vi.mock('../../../../shared/services/i18next', () => ({ selectLanguage: vi.fn() }));

const push = vi.hoisted(() => ({
  key: '',
  ask: vi.fn(),
  follow: vi.fn(),
}));
vi.mock('../../shared/lib/pushAlerts', () => ({
  usePushKey: () => push.key,
  askPushPermission: push.ask,
  followTicketByPush: push.follow,
}));

// The real alarm, with only the browser's audio, vibration and wake lock
// faked, so what the screen shows and what the phone would do stay in step.
vi.mock('../../../../shared/alarm/readyAlarm', async () => {
  const { createAlarm } = await import('../../../../shared/alarm/alarm');
  const { createFakeAlarmEnvironment } =
    await import('../../../../shared/alarm/testing/fakeAlarmEnvironment');
  const fakeAlarm = createFakeAlarmEnvironment();
  return { readyAlarm: createAlarm({ bellSrc: 'bell.wav', env: fakeAlarm.env }), fakeAlarm };
});

import Ticket from './Ticket';
import * as readyAlarmModule from '../../../../shared/alarm/readyAlarm';
import type { createFakeAlarmEnvironment } from '../../../../shared/alarm/testing/fakeAlarmEnvironment';
import { TICKET_POLL_INTERVAL_MS } from '../../shared/lib/ticketPoller';
import { initTestI18n } from '../../shared/testing/i18n';

const { readyAlarm } = readyAlarmModule;
const { bell, navigator: fakeNavigator } = (
  readyAlarmModule as unknown as { fakeAlarm: ReturnType<typeof createFakeAlarmEnvironment> }
).fakeAlarm;

const scanned = {
  code: 'A7',
  color: 2,
  status: 'waiting',
  venueName: 'Acme Cafe',
  serverTime: '2026-11-19T09:58:00.000Z',
};
const waiting = {
  label: '42',
  labelKind: 'number',
  status: 'waiting',
  venueName: 'Acme Cafe',
  ahead: 2,
  serverTime: '2026-11-19T10:00:00.000Z',
};
const ready = {
  label: '42',
  labelKind: 'number',
  status: 'ready',
  venueName: 'Acme Cafe',
  ahead: 0,
  readyAt: '2026-11-19T10:05:00.000Z',
  serverTime: '2026-11-19T10:05:01.000Z',
};

const nextPoll = (): Promise<void> =>
  act(async () => {
    await vi.advanceTimersByTimeAsync(TICKET_POLL_INTERVAL_MS);
  });

beforeEach(async () => {
  await initTestI18n('en');
  vi.useFakeTimers({ shouldAdvanceTime: true });
  params.token = 'tok-1';
  push.key = '';
  push.ask.mockReset();
  push.follow.mockReset().mockResolvedValue(true);
  api.getPublicTicket.mockReset();
  api.acknowledgeTicket.mockReset().mockResolvedValue({});
});

afterEach(() => {
  vi.useRealTimers();
  bell.play.mockClear();
  bell.pause.mockClear();
});

describe('pager ticket page', () => {
  // Runs first, before any test taps a button: sound starts locked, as it
  // does right after the scan.
  it('shows the code to show the cashier and one button to be notified', async () => {
    push.key = 'server-key';
    api.getPublicTicket.mockResolvedValue(scanned);
    render(<Ticket />);

    expect(await screen.findByRole('heading', { name: 'Your code A7' })).toBeInTheDocument();
    expect(screen.getByText('Show this code to the cashier')).toBeInTheDocument();
    expect(screen.queryByText(/ahead of you|You're next/)).not.toBeInTheDocument();
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    // Kept awake and followed by push before the customer taps anything.
    await waitFor(() => expect(fakeNavigator.wakeLock.request).toHaveBeenCalledWith('screen'));
    await waitFor(() => expect(push.follow).toHaveBeenCalled());

    const notify = screen.getByRole('button', { name: "Notify me when it's ready" });
    const followsBeforeTap = push.follow.mock.calls.length;
    fireEvent.click(notify);

    expect(readyAlarm.getSnapshot().unlocked).toBe(true);
    expect(bell.play).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('button', { name: "Notify me when it's ready" })).not.toBeInTheDocument();
    // The same tap asks for lock-screen notifications, inside the click, and
    // the page subscribes again in case the visitor allowed them.
    expect(push.ask).toHaveBeenCalledWith('server-key');
    await waitFor(() => expect(push.follow.mock.calls.length).toBeGreaterThan(followsBeforeTap));
  });

  it('switches from the code to the label once the cashier types it', async () => {
    api.getPublicTicket
      .mockResolvedValueOnce(scanned)
      .mockResolvedValue({ ...waiting, label: 'Sara', labelKind: 'name' });
    render(<Ticket />);
    await screen.findByRole('heading', { name: 'Your code A7' });

    await nextPoll();

    expect(await screen.findByRole('heading', { name: 'Sara' })).toBeInTheDocument();
    expect(screen.getByText("We're preparing your order")).toBeInTheDocument();
    expect(screen.getByText('2 orders ahead of you')).toBeInTheDocument();
    expect(screen.queryByText('Show this code to the cashier')).not.toBeInTheDocument();
  });

  it('names the customer in the lock-screen text when the label is a name', async () => {
    push.key = 'server-key';
    api.getPublicTicket.mockResolvedValue({ ...waiting, label: 'Sara', labelKind: 'name' });
    render(<Ticket />);

    await waitFor(() =>
      expect(push.follow).toHaveBeenLastCalledWith(
        expect.objectContaining({
          text: {
            ready: { title: 'Sara, your order is ready', body: 'Collect it at the counter of Acme Cafe.' },
            remind: {
              title: 'Sara, your order is waiting for you',
              body: 'Acme Cafe is calling your number again.',
            },
          },
        })
      )
    );
  });

  it('asks for a new scan when the code expired before anyone labelled it', async () => {
    api.getPublicTicket.mockResolvedValue({ ...scanned, status: 'expired' });
    render(<Ticket />);

    expect(await screen.findByText('This code has expired')).toBeInTheDocument();
    expect(
      screen.getByText('Scan the QR at the counter again to get a new one.')
    ).toBeInTheDocument();
  });

  it('shows the number, the queue and how to keep the alert alive while waiting', async () => {
    api.getPublicTicket.mockResolvedValue(waiting);
    render(<Ticket />);

    expect(await screen.findByRole('heading', { name: 'Order 42' })).toBeInTheDocument();
    expect(api.getPublicTicket).toHaveBeenCalledWith(
      { publicToken: 'tok-1' },
      expect.objectContaining({ timeoutMs: expect.any(Number) })
    );
    expect(screen.getByText('Acme Cafe')).toBeInTheDocument();
    expect(screen.getByText("We're preparing your order")).toBeInTheDocument();
    expect(screen.getByText('2 orders ahead of you')).toBeInTheDocument();
    expect(
      screen.getByText(
        'Keep this page open with the sound on. Your phone will ring when the order is ready.'
      )
    ).toBeInTheDocument();
    // The alarm is fed from an effect, which can run a moment after the
    // heading is already in the document.
    await waitFor(() => expect(fakeNavigator.wakeLock.request).toHaveBeenCalledWith('screen'));
  });

  it('says "next" when nobody is ahead', async () => {
    api.getPublicTicket.mockResolvedValue({ ...waiting, ahead: 0 });
    render(<Ticket />);

    expect(await screen.findByText("You're next")).toBeInTheDocument();
  });

  it('follows the ticket on the lock screen in the visitor language when push is on', async () => {
    push.key = 'server-key';
    api.getPublicTicket.mockResolvedValue(waiting);
    render(<Ticket />);

    await waitFor(() => expect(push.follow).toHaveBeenCalled());
    expect(push.follow).toHaveBeenLastCalledWith({
      publicKey: 'server-key',
      token: 'tok-1',
      text: {
        ready: { title: 'Order 42 is ready', body: 'Collect it at the counter of Acme Cafe.' },
        remind: {
          title: 'Order 42 is waiting for you',
          body: 'Acme Cafe is calling your number again.',
        },
      },
    });
  });

  it('leaves push alone while the server has it off', async () => {
    api.getPublicTicket.mockResolvedValue(waiting);
    render(<Ticket />);

    await screen.findByText("We're preparing your order");
    await nextPoll();

    expect(push.follow).not.toHaveBeenCalled();
  });

  it('rings when the order becomes ready and stops when the customer has it', async () => {
    api.getPublicTicket.mockResolvedValueOnce(waiting).mockResolvedValue(ready);
    render(<Ticket />);
    await screen.findByText("We're preparing your order");
    expect(bell.play).not.toHaveBeenCalled();

    await nextPoll();

    expect(screen.getByRole('alert')).toHaveTextContent('Your order is ready!');
    expect(readyAlarm.getSnapshot().ringing).toBe(true);
    expect(bell.play).toHaveBeenCalledTimes(1);
    expect(bell.muted).toBe(false);

    // The same ready answer on later polls does not start the ring over.
    await nextPoll();
    expect(bell.play).toHaveBeenCalledTimes(1);

    api.getPublicTicket.mockResolvedValue({
      ...ready,
      acknowledgedAt: '2026-11-19T10:05:20.000Z',
      serverTime: '2026-11-19T10:05:20.000Z',
    });
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: 'I have it' }));
    });

    expect(api.acknowledgeTicket).toHaveBeenCalledWith({ publicToken: 'tok-1' });
    expect(readyAlarm.getSnapshot().ringing).toBe(false);
    expect(bell.pause).toHaveBeenCalled();
    expect(
      await screen.findByText("Thanks. The counter knows you're on your way.")
    ).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'I have it' })).not.toBeInTheDocument();
  });

  it('"stop sound" silences this phone without telling the counter', async () => {
    api.getPublicTicket.mockResolvedValue(ready);
    render(<Ticket />);

    fireEvent.click(await screen.findByRole('button', { name: 'Stop sound' }));

    expect(readyAlarm.getSnapshot().ringing).toBe(false);
    expect(api.acknowledgeTicket).not.toHaveBeenCalled();
    expect(screen.queryByRole('button', { name: 'Stop sound' })).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'I have it' })).toBeInTheDocument();
  });

  it('rings again when the staff call again after an acknowledgement', async () => {
    api.getPublicTicket.mockResolvedValue({
      ...ready,
      acknowledgedAt: '2026-11-19T10:05:20.000Z',
    });
    render(<Ticket />);
    await screen.findByText("Thanks. The counter knows you're on your way.");
    expect(readyAlarm.getSnapshot().ringing).toBe(false);

    api.getPublicTicket.mockResolvedValue({
      ...ready,
      remindedAt: '2026-11-19T10:08:00.000Z',
      serverTime: '2026-11-19T10:08:00.000Z',
    });
    await nextPoll();

    expect(readyAlarm.getSnapshot().ringing).toBe(true);
    expect(screen.getByRole('button', { name: 'I have it' })).toBeInTheDocument();
  });

  it('keeps the buttons and says so when the acknowledgement fails', async () => {
    api.getPublicTicket.mockResolvedValue(ready);
    api.acknowledgeTicket.mockRejectedValue({ statusCode: 500 });
    render(<Ticket />);

    const gotIt = await screen.findByRole('button', { name: 'I have it' });
    await act(async () => {
      fireEvent.click(gotIt);
    });

    expect(await screen.findByText("We couldn't tell the counter. Try again.")).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'I have it' })).toBeEnabled();
  });

  it.each([
    ['collected', 'Order collected. Enjoy!'],
    ['expired', 'This ticket has expired'],
  ])('shows the end of a %s ticket and stops polling', async (status, message) => {
    api.getPublicTicket.mockResolvedValue({ ...waiting, status, ahead: 0 });
    render(<Ticket />);

    expect(await screen.findByText(message)).toBeInTheDocument();
    await nextPoll();

    expect(api.getPublicTicket).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('button', { name: 'I have it' })).not.toBeInTheDocument();
  });

  it('shows a dead end for a ticket that does not exist', async () => {
    api.getPublicTicket.mockRejectedValue({ statusCode: 404, code: 'ticket_not_found' });
    render(<Ticket />);

    expect(await screen.findByText('Ticket not found')).toBeInTheDocument();
  });

  it('offers a retry when the first load fails, and recovers', async () => {
    api.getPublicTicket.mockRejectedValueOnce({ statusCode: 503 }).mockResolvedValue(waiting);
    render(<Ticket />);

    fireEvent.click(await screen.findByRole('button', { name: 'Try again' }));

    expect(await screen.findByText("We're preparing your order")).toBeInTheDocument();
  });

  it('keeps the last known state on screen when a later poll fails', async () => {
    api.getPublicTicket.mockResolvedValueOnce(waiting).mockRejectedValue({ statusCode: 503 });
    render(<Ticket />);
    await screen.findByText("We're preparing your order");

    await nextPoll();

    expect(screen.getByText('Connection lost. Retrying...')).toBeInTheDocument();
    expect(screen.getByRole('heading', { name: 'Order 42' })).toBeInTheDocument();
  });

  it('renders Arabic copy, including the plural for a few orders ahead', async () => {
    await initTestI18n('ar');
    api.getPublicTicket.mockResolvedValue({ ...waiting, ahead: 3 });
    render(<Ticket />);

    expect(await screen.findByText('نحضّر طلبك الآن')).toBeInTheDocument();
    expect(screen.getByText('3 طلبات قبلك')).toBeInTheDocument();
  });

  it('lets go of the alarm when the customer leaves the page', async () => {
    api.getPublicTicket.mockResolvedValue(ready);
    const { unmount } = render(<Ticket />);
    await screen.findByText('Your order is ready!');
    await waitFor(() => expect(readyAlarm.getSnapshot().ringing).toBe(true));

    unmount();

    expect(readyAlarm.getSnapshot().ringing).toBe(false);
  });
});
