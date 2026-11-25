import React from 'react';
import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

const params = vi.hoisted(() => ({ token: 'tok-1' as string | undefined }));
vi.mock('@tanstack/react-router', () => ({ useParams: () => params }));

const api = vi.hoisted(() => ({ getPublicTicket: vi.fn(), acknowledgeTicket: vi.fn() }));
vi.mock('../../shared/lib/clients', () => ({ publicPagerClient: api }));

vi.mock('../../../../shared/services/i18next', () => ({ selectLanguage: vi.fn() }));

const push = vi.hoisted(() => ({
  key: '',
  perm: 'unsupported' as NotificationPermission | 'unsupported',
  ask: vi.fn(),
  askNow: vi.fn(),
  follow: vi.fn(),
  showLocal: vi.fn(),
  closeLocal: vi.fn(),
}));
// Device detection stays real: the iPhone and Android WebView tests fake the browser instead.
vi.mock('../../shared/lib/pushAlerts', async (importOriginal) => {
  const actual = await importOriginal<typeof PushAlerts>();
  return {
    iosTabWithoutPush: actual.iosTabWithoutPush,
    androidWebView: actual.androidWebView,
    openInChromeUrl: actual.openInChromeUrl,
    usePushKey: () => push.key,
    notificationPermission: () => push.perm,
    askPushPermission: push.ask,
    askPushPermissionNow: push.askNow,
    followTicketByPush: push.follow,
    showTicketNotification: push.showLocal,
    closeTicketNotification: push.closeLocal,
  };
});

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
import type * as PushAlerts from '../../shared/lib/pushAlerts';
import { TICKET_POLL_INTERVAL_MS } from '../../shared/lib/ticketPoller';
import { initTestI18n } from '../../shared/testing/i18n';
import { installFakePushBrowser } from '../../shared/testing/fakePushBrowser';

const { readyAlarm } = readyAlarmModule;
const { bell, navigator: fakeNavigator } = (
  readyAlarmModule as unknown as { fakeAlarm: ReturnType<typeof createFakeAlarmEnvironment> }
).fakeAlarm;

const scanned = {
  code: '427',
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
  window.localStorage.clear();
  vi.useFakeTimers({ shouldAdvanceTime: true });
  params.token = 'tok-1';
  push.key = '';
  push.perm = 'unsupported';
  push.ask.mockReset();
  push.askNow.mockReset().mockReturnValue(null);
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

    expect(await screen.findByRole('heading', { name: 'Your code 427' })).toBeInTheDocument();
    expect(screen.getByText('Show this number to the cashier')).toBeInTheDocument();
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
    expect(
      screen.queryByRole('button', { name: "Notify me when it's ready" })
    ).not.toBeInTheDocument();
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
    await screen.findByRole('heading', { name: 'Your code 427' });

    await nextPoll();

    expect(await screen.findByRole('heading', { name: 'Sara' })).toBeInTheDocument();
    expect(screen.getByText("We're preparing your order")).toBeInTheDocument();
    expect(screen.getByText('2 orders ahead of you')).toBeInTheDocument();
    expect(screen.queryByText('Show this number to the cashier')).not.toBeInTheDocument();
    expect(screen.queryByText('427')).not.toBeInTheDocument();
  });

  it('names the customer in the lock-screen text when the label is a name', async () => {
    push.key = 'server-key';
    api.getPublicTicket.mockResolvedValue({ ...waiting, label: 'Sara', labelKind: 'name' });
    render(<Ticket />);

    await waitFor(() =>
      expect(push.follow).toHaveBeenLastCalledWith(
        expect.objectContaining({
          text: {
            ready: {
              title: 'Sara, your order is ready',
              body: 'Collect it at the counter of Acme Cafe.',
            },
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

  describe('on an iPhone', () => {
    const IPHONE_SAFARI =
      'Mozilla/5.0 (iPhone; CPU iPhone OS 18_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.1 Mobile/15E148 Safari/604.1';

    it('warns that a Safari tab rings only while the page is on screen', async () => {
      const userAgent = vi.spyOn(navigator, 'userAgent', 'get').mockReturnValue(IPHONE_SAFARI);
      try {
        api.getPublicTicket.mockResolvedValue(waiting);
        render(<Ticket />);

        expect(await screen.findByText('On iPhone, keep this page open')).toBeInTheDocument();
        expect(screen.getByText(/can only ring while this page is on screen/)).toBeInTheDocument();
        expect(
          screen.queryByText(/^Keep this page open with the sound on/)
        ).not.toBeInTheDocument();
      } finally {
        userAgent.mockRestore();
      }
    });

    it('does not warn a home screen app, which gets lock-screen notifications', async () => {
      const userAgent = vi.spyOn(navigator, 'userAgent', 'get').mockReturnValue(IPHONE_SAFARI);
      const pushBrowser = installFakePushBrowser();
      try {
        api.getPublicTicket.mockResolvedValue(waiting);
        render(<Ticket />);

        expect(
          await screen.findByText(/^Keep this page open with the sound on/)
        ).toBeInTheDocument();
        expect(screen.queryByText('On iPhone, keep this page open')).not.toBeInTheDocument();
      } finally {
        pushBrowser.restore();
        userAgent.mockRestore();
      }
    });
  });

  it('counts down to the venue estimate, then says it is coming any moment', async () => {
    api.getPublicTicket.mockResolvedValue({
      ...waiting,
      expectedReadyAt: '2026-11-19T10:03:30.000Z',
    });
    render(<Ticket />);

    expect(await screen.findByText('Ready in about 4 min')).toBeInTheDocument();

    api.getPublicTicket.mockResolvedValue({
      ...waiting,
      expectedReadyAt: '2026-11-19T10:03:30.000Z',
      serverTime: '2026-11-19T10:04:00.000Z',
    });
    await nextPoll();
    expect(await screen.findByText('Ready any moment now')).toBeInTheDocument();
  });

  it('offers the venue review page only once the order is in hand', async () => {
    const reviewUrl = 'https://g.page/r/acme/review';
    api.getPublicTicket.mockResolvedValue({ ...waiting, reviewUrl });
    render(<Ticket />);

    expect(await screen.findByRole('heading', { name: 'Order 42' })).toBeInTheDocument();
    expect(screen.queryByRole('link', { name: 'Leave a Google review' })).not.toBeInTheDocument();

    api.getPublicTicket.mockResolvedValue({ ...ready, status: 'collected', reviewUrl });
    await nextPoll();
    const link = await screen.findByRole('link', { name: 'Leave a Google review' });
    expect(link).toHaveAttribute('href', reviewUrl);
    expect(link).toHaveAttribute('target', '_blank');
    expect(link.getAttribute('rel')).toContain('noopener');
  });

  describe('ringtone', () => {
    afterEach(() => {
      window.localStorage.removeItem('pager:ringtone');
    });

    const lastSound = (sync: ReturnType<typeof vi.spyOn>): string =>
      String((sync.mock.calls.at(-1)?.[0] as { bellSrc?: string } | undefined)?.bellSrc);

    it("rings with the venue's tone until the customer picks their own, and keeps it", async () => {
      const sync = vi.spyOn(readyAlarm, 'sync');
      try {
        api.getPublicTicket.mockResolvedValue({ ...waiting, ringtone: 'marimba' });
        const { unmount } = render(<Ticket />);

        const tones = await screen.findByRole('group', { name: 'Your ringtone' });
        expect(within(tones).getByRole('button', { name: 'Marimba' })).toHaveAttribute(
          'aria-pressed',
          'true'
        );
        await waitFor(() => expect(lastSound(sync)).toContain('marimba'));

        fireEvent.click(within(tones).getByRole('button', { name: 'Classic ring' }));
        await waitFor(() => expect(lastSound(sync)).toContain('ring'));
        unmount();

        // The pick outlasts the page: the next visit rings the same.
        render(<Ticket />);
        const again = await screen.findByRole('group', { name: 'Your ringtone' });
        expect(within(again).getByRole('button', { name: 'Classic ring' })).toHaveAttribute(
          'aria-pressed',
          'true'
        );
      } finally {
        sync.mockRestore();
      }
    });
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

  it("shows the venue's logo beside its name, and nothing when the file fails", async () => {
    api.getPublicTicket.mockResolvedValue({
      ...waiting,
      venueLogoUrl: 'https://cdn.example.test/acme.png',
    });
    const { container } = render(<Ticket />);

    await screen.findByText('Acme Cafe');
    const logo = container.querySelector('.pager-venue img');
    expect(logo).toHaveAttribute('src', 'https://cdn.example.test/acme.png');
    fireEvent.error(logo as HTMLImageElement);
    expect(container.querySelector('.pager-venue img')).toBeNull();
    expect(screen.getByText('Acme Cafe')).toBeInTheDocument();
  });

  it('shows a dead end for a ticket that does not exist', async () => {
    api.getPublicTicket.mockRejectedValue({ statusCode: 404, code: 'ticket_not_found' });
    render(<Ticket />);

    expect(await screen.findByText('Ticket not found')).toBeInTheDocument();
  });

  // The cashier's screen may be in another language: both read the same digits.
  it('shows the number in Western digits, left to right, on the Arabic page', async () => {
    await initTestI18n('ar');
    api.getPublicTicket.mockResolvedValue(scanned);
    render(<Ticket />);

    const heading = await screen.findByRole('heading', { name: 'رمزك 427' });
    expect(within(heading).getByText('427')).toHaveAttribute('dir', 'ltr');
    expect(screen.getByText('أرِ هذا الرقم للكاشير')).toBeInTheDocument();
  });

  it.each(['A7', 'PANDA2'])(
    'still shows the code of a phone that scanned before the numbers (%s)',
    async (code) => {
      api.getPublicTicket.mockResolvedValue({ ...scanned, code });
      render(<Ticket />);

      expect(await screen.findByRole('heading', { name: `Your code ${code}` })).toBeInTheDocument();
    }
  );

  it('waits calmly through a deploy, then offers a retry, and recovers', async () => {
    api.getPublicTicket.mockRejectedValue({ statusCode: 503 });
    render(<Ticket />);

    expect(await screen.findByText('Still getting your code, one moment...')).toBeInTheDocument();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(15_000);
    });
    expect(screen.queryByRole('button', { name: 'Try again' })).not.toBeInTheDocument();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5_000);
    });

    api.getPublicTicket.mockResolvedValue(waiting);
    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

    expect(await screen.findByText("We're preparing your order")).toBeInTheDocument();
  });

  it('keeps polling through a 404 the pager did not give a code', async () => {
    api.getPublicTicket.mockRejectedValueOnce({ statusCode: 404 }).mockResolvedValue(waiting);
    render(<Ticket />);
    await screen.findByText('Still getting your code, one moment...');

    await nextPoll();

    expect(await screen.findByText("We're preparing your order")).toBeInTheDocument();
    expect(screen.queryByText('Ticket not found')).not.toBeInTheDocument();
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

  it('keeps the notify button visible when sound is unlocked but notification permission is still default', async () => {
    push.key = 'server-key';
    push.perm = 'default';
    push.askNow.mockReturnValue(Promise.resolve('granted' as NotificationPermission));
    api.getPublicTicket.mockResolvedValue(scanned);
    render(<Ticket />);

    const notify = await screen.findByRole('button', { name: "Notify me when it's ready" });
    fireEvent.click(notify);

    expect(push.askNow).toHaveBeenCalledTimes(1);
    await waitFor(() =>
      expect(
        screen.queryByRole('button', { name: "Notify me when it's ready" })
      ).not.toBeInTheDocument()
    );
  });

  it('clears the remembered venue scan and shows Start a new order once collected', async () => {
    window.localStorage.setItem('pager:scan:acme', 'tok-1');
    window.localStorage.setItem('pager:token:tok-1', 'acme');
    api.getPublicTicket.mockResolvedValue({ ...waiting, status: 'collected' });
    render(<Ticket />);

    const link = await screen.findByRole('link', { name: 'Start a new order' });
    expect(link).toHaveAttribute('href', '/acme?new=1');
    await waitFor(() => expect(window.localStorage.getItem('pager:scan:acme')).toBeNull());
  });

  it('triggers a local Service Worker notification when ready and closes it on acknowledge', async () => {
    push.showLocal.mockReset().mockResolvedValue(true);
    push.closeLocal.mockReset().mockResolvedValue(undefined);
    api.getPublicTicket.mockResolvedValue(ready);
    api.acknowledgeTicket.mockResolvedValue({});
    render(<Ticket />);

    await screen.findByText('Your order is ready!');
    await waitFor(() =>
      expect(push.showLocal).toHaveBeenCalledWith(
        expect.objectContaining({
          token: 'tok-1',
          title: 'Order 42 is ready',
        })
      )
    );

    fireEvent.click(screen.getByRole('button', { name: 'I have it' }));
    await waitFor(() => expect(push.closeLocal).toHaveBeenCalledWith('tok-1'));
  });

  it('shows the Open in Chrome helper when opened inside an Android QR WebView', async () => {
    const originalUA = navigator.userAgent;
    Object.defineProperty(navigator, 'userAgent', {
      configurable: true,
      value:
        'Mozilla/5.0 (Linux; Android 14; Pixel 8; wv) AppleWebKit/537.36 (KHTML, like Gecko) Version/4.0 Chrome/120.0.0.0 Mobile Safari/537.36',
    });
    api.getPublicTicket.mockResolvedValue(waiting);
    try {
      render(<Ticket />);
      expect(
        await screen.findByText('Open in Chrome for vibration & lock-screen alerts')
      ).toBeInTheDocument();
      const link = screen.getByRole('link', { name: 'Open in Chrome' });
      expect(link.getAttribute('href')).toContain('package=com.android.chrome');
    } finally {
      Object.defineProperty(navigator, 'userAgent', { configurable: true, value: originalUA });
    }
  });
});
