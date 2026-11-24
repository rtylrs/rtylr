import React, { StrictMode } from 'react';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

// Real translations: an assertion written in English also proves the key resolves.

vi.mock('@voxire/components', () => ({ HelmetContainer: () => null }));

const navigateSpy = vi.hoisted(() => vi.fn());
vi.mock('@tanstack/react-router', () => ({
  useParams: () => ({ slug: 'Acme' }),
  useNavigate: () => navigateSpy,
}));

const api = vi.hoisted(() => ({
  getPublicVenue: vi.fn(),
  scanTicket: vi.fn(),
  getPublicTicket: vi.fn(),
  getPushKey: vi.fn(),
}));
vi.mock('../../shared/lib/clients', () => ({ publicPagerClient: api }));

const alarmMocks = vi.hoisted(() => ({
  unlock: vi.fn(),
  blip: vi.fn(),
}));
vi.mock('../../../../shared/alarm/readyAlarm', () => ({
  readyAlarm: alarmMocks,
}));

const pushMocks = vi.hoisted(() => ({
  key: 'vapid-public-key',
  ask: vi.fn(),
  askNow: vi.fn(),
}));
vi.mock('../../shared/lib/pushAlerts', () => ({
  askPushPermission: pushMocks.ask,
  askPushPermissionNow: pushMocks.askNow,
  usePushKey: () => pushMocks.key,
}));

vi.mock('../../../../shared/services/i18next', () => ({ selectLanguage: vi.fn() }));

import Scan from './Scan';
import { initTestI18n } from '../../shared/testing/i18n';

const STORAGE_KEY = 'pager:scan:acme';
const scanned = { publicToken: 'tok-1', code: '427', status: 'waiting' };

async function tapStart(): Promise<void> {
  fireEvent.click(await screen.findByRole('button', { name: 'Get my code & enable alerts' }));
}

beforeAll(async () => {
  await initTestI18n('en');
});

beforeEach(() => {
  window.localStorage.clear();
  navigateSpy.mockReset();
  alarmMocks.unlock.mockReset();
  alarmMocks.blip.mockReset();
  pushMocks.key = 'vapid-public-key';
  pushMocks.ask.mockReset();
  pushMocks.askNow.mockReset();
  api.getPublicVenue.mockReset().mockResolvedValue({ name: 'Acme Café', enabled: true });
  api.scanTicket.mockReset().mockResolvedValue(scanned);
  api.getPublicTicket.mockReset();
});

afterEach(() => {
  vi.useRealTimers();
});

describe('pager scan page', () => {
  it('shows the venue and unlocks audio + notifications on a single upfront tap before scanning', async () => {
    render(
      <StrictMode>
        <Scan />
      </StrictMode>
    );

    expect(await screen.findByText('Acme Café')).toBeInTheDocument();
    expect(api.scanTicket).not.toHaveBeenCalled();

    await tapStart();

    expect(alarmMocks.unlock).toHaveBeenCalledTimes(1);
    expect(alarmMocks.blip).toHaveBeenCalledTimes(1);
    expect(pushMocks.ask).toHaveBeenCalledWith('vapid-public-key');
    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true })
    );
    expect(api.scanTicket).toHaveBeenCalledTimes(1);
    expect(api.scanTicket).toHaveBeenCalledWith({ slug: 'acme' });
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('tok-1');
  });

  it('keeps the same ticket and code on a reload or a scan in a new tab', async () => {
    window.localStorage.setItem(STORAGE_KEY, 'tok-old');
    api.getPublicTicket.mockResolvedValue({ status: 'waiting', code: '318' });
    render(<Scan />);

    await tapStart();

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-old', replace: true })
    );
    expect(api.getPublicTicket).toHaveBeenCalledWith({ publicToken: 'tok-old' });
    expect(api.scanTicket).not.toHaveBeenCalled();
  });

  it.each([
    ['expired', { status: 'expired' }],
    ['collected', { status: 'collected' }],
    ['acknowledged', { status: 'ready', acknowledgedAt: '2026-11-24T08:00:00.000Z' }],
  ])('scans again when the remembered ticket is %s', async (_, ticket) => {
    window.localStorage.setItem(STORAGE_KEY, 'tok-old');
    api.getPublicTicket.mockResolvedValue(ticket);
    render(<Scan />);

    await tapStart();

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true })
    );
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('tok-1');
  });

  it('requests notification permission synchronously even before the VAPID key finishes loading', async () => {
    pushMocks.key = '';
    render(<Scan />);

    await tapStart();

    expect(pushMocks.askNow).toHaveBeenCalledTimes(1);
    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true })
    );
  });

  it('lets a returning customer start a fresh order even when an old scan is remembered', async () => {
    window.localStorage.setItem(STORAGE_KEY, 'tok-old');
    api.getPublicTicket.mockResolvedValue({ status: 'waiting', code: '318' });
    render(<Scan />);

    fireEvent.click(await screen.findByRole('button', { name: 'Start a new order instead' }));

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true })
    );
    expect(api.scanTicket).toHaveBeenCalledTimes(1);
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('tok-1');
  });

  it('scans again when the remembered ticket is gone', async () => {
    window.localStorage.setItem(STORAGE_KEY, 'tok-old');
    api.getPublicTicket.mockRejectedValue({ statusCode: 404, code: 'ticket_not_found' });
    render(<Scan />);

    await tapStart();

    await waitFor(() => expect(api.scanTicket).toHaveBeenCalledTimes(1));
    expect(window.localStorage.getItem(STORAGE_KEY)).toBe('tok-1');
  });

  it('keeps the remembered ticket when it cannot be checked, instead of a second phone', async () => {
    window.localStorage.setItem(STORAGE_KEY, 'tok-old');
    api.getPublicTicket.mockRejectedValue({ statusCode: 503 });
    render(<Scan />);

    await tapStart();

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-old', replace: true })
    );
    expect(api.scanTicket).not.toHaveBeenCalled();
  });

  it('still works when the browser refuses storage', async () => {
    const getItem = vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new DOMException('blocked', 'SecurityError');
    });
    const setItem = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new DOMException('blocked', 'SecurityError');
    });
    render(<Scan />);

    await tapStart();

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true })
    );
    getItem.mockRestore();
    setItem.mockRestore();
  });

  it('hands the ticket to the demo instead of navigating, and remembers nothing', async () => {
    const onScanned = vi.fn();
    render(<Scan slug="demo" onScanned={onScanned} />);

    await tapStart();

    await waitFor(() => expect(onScanned).toHaveBeenCalledWith('tok-1'));
    expect(navigateSpy).not.toHaveBeenCalled();
  });

  it('says so when the venue has no pager', async () => {
    api.getPublicVenue.mockRejectedValue({ statusCode: 404, code: 'venue_not_found' });
    render(<Scan />);

    expect(await screen.findByText("This pager isn't available")).toBeInTheDocument();
    expect(screen.queryByRole('button')).not.toBeInTheDocument();
  });

  it.each([
    [{ statusCode: 429 }, 'Too many attempts. Wait a moment, then try again.'],
    [
      { statusCode: 429, code: 'too_many_open_tickets' },
      'This venue has too many phones waiting right now. Please ask at the counter.',
    ],
  ])('explains a failed scan and retries on a tap: %j', async (failure, message) => {
    api.scanTicket.mockRejectedValueOnce(failure);
    render(<Scan />);

    await tapStart();

    expect(await screen.findByText(message)).toBeInTheDocument();
    expect(navigateSpy).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

    await waitFor(() =>
      expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true })
    );
    expect(api.scanTicket).toHaveBeenCalledTimes(2);
  });

  it('rides out a deploy calmly, retrying after 1, 2 and 4 seconds', async () => {
    vi.useFakeTimers();
    api.scanTicket
      .mockRejectedValueOnce(new TypeError('Failed to fetch'))
      .mockRejectedValueOnce({ statusCode: 503 })
      .mockRejectedValueOnce({ statusCode: 404 })
      .mockResolvedValueOnce(scanned);
    render(<Scan />);

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0);
    });
    fireEvent.click(screen.getByRole('button', { name: 'Get my code & enable alerts' }));

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1_000);
    });
    expect(screen.getByRole('status')).toHaveTextContent('Still getting your code, one moment...');
    expect(navigateSpy).not.toHaveBeenCalled();

    await act(async () => {
      await vi.advanceTimersByTimeAsync(6_000);
    });
    expect(navigateSpy).toHaveBeenCalledWith({ to: '/t/tok-1', replace: true });
    expect(api.scanTicket).toHaveBeenCalledTimes(4);
  });
});
